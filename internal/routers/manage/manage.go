package manage

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

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
	url := ctx.QueryTrim("url")
	template := ctx.QueryTrim("template")
	settingModel := new(models.Setting)
	err := settingModel.UpdateWebHook(url, template)

	return utils.JsonResponseByErr(err)
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
