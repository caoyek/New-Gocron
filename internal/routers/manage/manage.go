package manage

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/caoyek/New-Gocron/internal/models"
	"github.com/caoyek/New-Gocron/internal/modules/logger"
	"github.com/caoyek/New-Gocron/internal/modules/utils"
	"github.com/caoyek/New-Gocron/internal/service"
	"gopkg.in/macaron.v1"
)

func Slack(ctx *macaron.Context) string {
	settingModel := new(models.Setting)
	slack, err := settingModel.Slack()
	jsonResp := utils.JsonResponse{}
	if err != nil {
		logger.Error(err)
		return jsonResp.Success(utils.SuccessContent, nil)

	}

	return jsonResp.Success(utils.SuccessContent, slack)
}

func UpdateSlack(ctx *macaron.Context) string {
	url := ctx.QueryTrim("url")
	template := ctx.QueryTrim("template")
	settingModel := new(models.Setting)
	err := settingModel.UpdateSlack(url, template)

	return utils.JsonResponseByErr(err)
}

func CreateSlackChannel(ctx *macaron.Context) string {
	channel := ctx.QueryTrim("channel")
	settingModel := new(models.Setting)
	if settingModel.IsChannelExist(channel) {
		jsonResp := utils.JsonResponse{}

		return jsonResp.CommonFailure("Channel已存在")
	}
	_, err := settingModel.CreateChannel(channel)

	return utils.JsonResponseByErr(err)
}

func RemoveSlackChannel(ctx *macaron.Context) string {
	id := ctx.ParamsInt(":id")
	settingModel := new(models.Setting)
	_, err := settingModel.RemoveChannel(id)

	return utils.JsonResponseByErr(err)
}

// endregion

// region 邮件
func Mail(ctx *macaron.Context) string {
	settingModel := new(models.Setting)
	mail, err := settingModel.Mail()
	jsonResp := utils.JsonResponse{}
	if err != nil {
		logger.Error(err)
		return jsonResp.Success(utils.SuccessContent, nil)
	}

	return jsonResp.Success("", mail)
}

type MailServerForm struct {
	Host     string `binding:"Required;MaxSize(100)"`
	Port     int    `binding:"Required;Range(1-65535)"`
	User     string `binding:"Required;MaxSize(64);Email"`
	Password string `binding:"Required;MaxSize(64)"`
}

func UpdateMail(ctx *macaron.Context, form MailServerForm) string {
	jsonByte, _ := json.Marshal(form)
	settingModel := new(models.Setting)

	template := ctx.QueryTrim("template")
	err := settingModel.UpdateMail(string(jsonByte), template)

	return utils.JsonResponseByErr(err)
}

func CreateMailUser(ctx *macaron.Context) string {
	username := ctx.QueryTrim("username")
	email := ctx.QueryTrim("email")
	settingModel := new(models.Setting)
	if username == "" || email == "" {
		jsonResp := utils.JsonResponse{}

		return jsonResp.CommonFailure("用户名、邮箱均不能为空")
	}
	_, err := settingModel.CreateMailUser(username, email)

	return utils.JsonResponseByErr(err)
}

func RemoveMailUser(ctx *macaron.Context) string {
	id := ctx.ParamsInt(":id")
	settingModel := new(models.Setting)
	_, err := settingModel.RemoveMailUser(id)

	return utils.JsonResponseByErr(err)
}

func WebHook(ctx *macaron.Context) string {
	settingModel := new(models.Setting)
	webHook, err := settingModel.Webhook()
	jsonResp := utils.JsonResponse{}
	if err != nil {
		logger.Error(err)
		return jsonResp.Success(utils.SuccessContent, nil)
	}

	return jsonResp.Success("", webHook)
}

func UpdateWebHook(ctx *macaron.Context) string {
	template := ctx.QueryTrim("template")
	jsonResp := utils.JsonResponse{}
	if template == "" {
		return jsonResp.CommonFailure("通知模板不能为空")
	}
	settingModel := new(models.Setting)
	err := settingModel.UpdateDefaultWebhookTemplate(template)

	return utils.JsonResponseByErr(err)
}

func CreateWebhookTemplate(ctx *macaron.Context) string {
	name := ctx.QueryTrim("name")
	content := ctx.QueryTrim("content")
	jsonResp := utils.JsonResponse{}
	if errMessage := validateWebhookTemplate(name, content); errMessage != "" {
		return jsonResp.CommonFailure(errMessage)
	}

	settingModel := new(models.Setting)
	exists, err := settingModel.WebhookTemplateNameExists(name, 0)
	if err != nil {
		return jsonResp.CommonFailure("读取通知模板失败", err)
	}
	if exists {
		return jsonResp.CommonFailure("模板名称已存在")
	}
	_, err = settingModel.CreateWebhookTemplate(name, content)

	return utils.JsonResponseByErr(err)
}

func UpdateWebhookTemplate(ctx *macaron.Context) string {
	id := ctx.ParamsInt(":id")
	name := ctx.QueryTrim("name")
	content := ctx.QueryTrim("content")
	jsonResp := utils.JsonResponse{}
	if id <= 0 {
		return jsonResp.CommonFailure("通知模板不存在")
	}
	if errMessage := validateWebhookTemplate(name, content); errMessage != "" {
		return jsonResp.CommonFailure(errMessage)
	}

	settingModel := new(models.Setting)
	exists, err := settingModel.WebhookTemplateNameExists(name, id)
	if err != nil {
		return jsonResp.CommonFailure("读取通知模板失败", err)
	}
	if exists {
		return jsonResp.CommonFailure("模板名称已存在")
	}
	updated, err := settingModel.UpdateWebhookTemplate(id, name, content)
	if err != nil {
		return jsonResp.CommonFailure("保存通知模板失败", err)
	}
	if updated == 0 {
		return jsonResp.CommonFailure("通知模板不存在")
	}

	return jsonResp.Success("保存成功", nil)
}

func RemoveWebhookTemplate(ctx *macaron.Context) string {
	id := ctx.ParamsInt(":id")
	jsonResp := utils.JsonResponse{}
	if id <= 0 {
		return jsonResp.CommonFailure("通知模板不存在")
	}
	settingModel := new(models.Setting)
	webHook, err := settingModel.Webhook()
	if err != nil {
		return jsonResp.CommonFailure("读取通知模板失败", err)
	}
	templateExists := false
	for _, template := range webHook.Templates {
		if template.Id != id {
			continue
		}
		if template.IsDefault {
			return jsonResp.CommonFailure("默认模板不能删除")
		}
		templateExists = true
		break
	}
	if !templateExists {
		return jsonResp.CommonFailure("通知模板不存在")
	}
	inUse, err := new(models.Task).WebhookTemplateInUse(id)
	if err != nil {
		return jsonResp.CommonFailure("检查通知模板使用情况失败", err)
	}
	if inUse {
		return jsonResp.CommonFailure("该通知模板仍被任务使用，请先调整任务通知配置")
	}

	removed, err := settingModel.RemoveWebhookTemplate(id)
	if err != nil {
		return jsonResp.CommonFailure("删除通知模板失败", err)
	}
	if removed == 0 {
		return jsonResp.CommonFailure("通知模板不存在")
	}

	return jsonResp.Success("删除成功", nil)
}

func CreateWebhookGroup(ctx *macaron.Context) string {
	name := ctx.QueryTrim("name")
	webhookUrl := ctx.QueryTrim("url")
	jsonResp := utils.JsonResponse{}
	if errMessage := validateWebhookGroup(name, webhookUrl); errMessage != "" {
		return jsonResp.CommonFailure(errMessage)
	}

	settingModel := new(models.Setting)
	exists, err := settingModel.WebhookGroupNameExists(name, 0)
	if err != nil {
		return jsonResp.CommonFailure("读取企微群配置失败", err)
	}
	if exists {
		return jsonResp.CommonFailure("企微群名称已存在")
	}
	_, err = settingModel.CreateWebhookGroup(name, webhookUrl)

	return utils.JsonResponseByErr(err)
}

func UpdateWebhookGroup(ctx *macaron.Context) string {
	id := ctx.ParamsInt(":id")
	name := ctx.QueryTrim("name")
	webhookUrl := ctx.QueryTrim("url")
	jsonResp := utils.JsonResponse{}
	if id <= 0 {
		return jsonResp.CommonFailure("企微群配置不存在")
	}
	if errMessage := validateWebhookGroup(name, webhookUrl); errMessage != "" {
		return jsonResp.CommonFailure(errMessage)
	}

	settingModel := new(models.Setting)
	exists, err := settingModel.WebhookGroupNameExists(name, id)
	if err != nil {
		return jsonResp.CommonFailure("读取企微群配置失败", err)
	}
	if exists {
		return jsonResp.CommonFailure("企微群名称已存在")
	}
	updated, err := settingModel.UpdateWebhookGroup(id, name, webhookUrl)
	if err != nil {
		return jsonResp.CommonFailure("保存企微群配置失败", err)
	}
	if updated == 0 {
		return jsonResp.CommonFailure("企微群配置不存在")
	}

	return jsonResp.Success("保存成功", nil)
}

func RemoveWebhookGroup(ctx *macaron.Context) string {
	id := ctx.ParamsInt(":id")
	jsonResp := utils.JsonResponse{}
	if id <= 0 {
		return jsonResp.CommonFailure("企微群配置不存在")
	}
	settingModel := new(models.Setting)
	webHook, err := settingModel.Webhook()
	if err != nil {
		return jsonResp.CommonFailure("读取企微群配置失败", err)
	}
	groupExists := false
	defaultGroupId := 0
	if len(webHook.Groups) > 0 {
		defaultGroupId = webHook.Groups[0].Id
	}
	for _, group := range webHook.Groups {
		if group.Id == id {
			groupExists = true
			break
		}
	}
	if !groupExists {
		return jsonResp.CommonFailure("企微群配置不存在")
	}
	inUse, err := new(models.Task).WebhookGroupInUse(id, defaultGroupId)
	if err != nil {
		return jsonResp.CommonFailure("检查企微群使用情况失败", err)
	}
	if inUse {
		return jsonResp.CommonFailure("该企微群仍被任务使用，请先调整任务通知配置")
	}

	removed, err := settingModel.RemoveWebhookGroup(id)
	if err != nil {
		return jsonResp.CommonFailure("删除企微群配置失败", err)
	}
	if removed == 0 {
		return jsonResp.CommonFailure("企微群配置不存在")
	}

	return jsonResp.Success("删除成功", nil)
}

func validateWebhookGroup(name, webhookUrl string) string {
	if name == "" {
		return "企微群名称不能为空"
	}
	if utf8.RuneCountInString(name) > 64 {
		return "企微群名称不能超过64个字符"
	}
	if webhookUrl == "" {
		return "Webhook URL不能为空"
	}
	if len(webhookUrl) > 2048 {
		return "Webhook URL不能超过2048个字符"
	}
	parsedUrl, err := url.ParseRequestURI(webhookUrl)
	if err != nil || parsedUrl.Host == "" || (parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https") {
		return "请输入有效的Webhook URL"
	}

	return ""
}

func validateWebhookTemplate(name, content string) string {
	if name == "" {
		return "模板名称不能为空"
	}
	if utf8.RuneCountInString(name) > 48 {
		return "模板名称不能超过48个字符"
	}
	if content == "" {
		return "通知模板不能为空"
	}
	if utf8.RuneCountInString(content) > 4096 {
		return "通知模板不能超过4096个字符"
	}

	return ""
}

func LoginSecurity(ctx *macaron.Context) string {
	jsonResp := utils.JsonResponse{}
	policy, err := service.LoginSecurityPolicy()
	if err != nil {
		return jsonResp.CommonFailure("读取登录安全配置失败", err)
	}
	blocks, err := new(models.LoginBlock).ListActive(time.Now())
	if err != nil {
		return jsonResp.CommonFailure("读取封禁列表失败", err)
	}

	return jsonResp.Success(utils.SuccessContent, map[string]interface{}{
		"policy":  policy,
		"blocks":  blocks,
		"peer_ip": service.RequestPeerIP(ctx),
	})
}

func UpdateLoginSecurity(ctx *macaron.Context) string {
	jsonResp := utils.JsonResponse{}
	windowMinutes, err := strconv.Atoi(ctx.QueryTrim("window_minutes"))
	if err != nil || windowMinutes < 1 || windowMinutes > 1440 {
		return jsonResp.CommonFailure("统计周期必须为 1-1440 分钟")
	}
	maxFailures, err := strconv.Atoi(ctx.QueryTrim("max_failures"))
	if err != nil || maxFailures < 1 || maxFailures > 100 {
		return jsonResp.CommonFailure("失败次数必须为 1-100 次")
	}
	blockMinutes, err := strconv.Atoi(ctx.QueryTrim("block_minutes"))
	if err != nil || blockMinutes < 1 || blockMinutes > 10080 {
		return jsonResp.CommonFailure("封禁时长必须为 1-10080 分钟")
	}
	whitelist, err := service.NormalizeWhitelist(ctx.Query("whitelist"))
	if err != nil {
		return jsonResp.CommonFailure(err.Error())
	}
	whitelistEnabled := parseBool(ctx.QueryTrim("whitelist_enabled"))
	peerIP := service.RequestPeerIP(ctx)
	if whitelistEnabled && !service.IPAllowed(peerIP, whitelist) {
		return jsonResp.CommonFailure("当前访问 IP " + peerIP + " 不在白名单中，无法启用")
	}
	policy := models.LoginSecurityPolicy{
		BlockEnabled:     parseBool(ctx.QueryTrim("block_enabled")),
		WindowMinutes:    windowMinutes,
		MaxFailures:      maxFailures,
		BlockMinutes:     blockMinutes,
		WhitelistEnabled: whitelistEnabled,
		Whitelist:        whitelist,
	}
	if err = new(models.Setting).UpdateLoginSecurity(policy); err != nil {
		return jsonResp.CommonFailure("保存登录安全配置失败", err)
	}

	return jsonResp.Success("保存成功", nil)
}

func RemoveLoginBlock(ctx *macaron.Context) string {
	jsonResp := utils.JsonResponse{}
	id := ctx.ParamsInt(":id")
	if id <= 0 {
		return jsonResp.CommonFailure("封禁记录不存在")
	}
	if err := new(models.LoginBlock).Remove(id); err != nil {
		return jsonResp.CommonFailure("解除封禁失败", err)
	}

	return jsonResp.Success("已解除封禁", nil)
}

func parseBool(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))

	return value == "1" || value == "true" || value == "on"
}

// endregion
