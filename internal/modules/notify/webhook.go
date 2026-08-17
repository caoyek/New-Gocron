package notify

import (
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/caoyek/New-Gocron/internal/models"
	"github.com/caoyek/New-Gocron/internal/modules/httpclient"
	"github.com/caoyek/New-Gocron/internal/modules/logger"
	"github.com/caoyek/New-Gocron/internal/modules/utils"
)

type WebHook struct{}

func (webHook *WebHook) Send(msg Message) {
	model := new(models.Setting)
	webHookSetting, err := model.Webhook()
	if err != nil {
		logger.Error("#webHook#从数据库获取webHook配置失败", err)
		return
	}
	if len(webHookSetting.Groups) == 0 {
		logger.Error("#webHook#企微群配置为空")
		return
	}
	groups := webHook.getActiveGroups(webHookSetting, msg)
	if len(groups) == 0 {
		logger.Error("#webHook#未找到任务配置的企微群")
		return
	}
	template, ok := webHook.getActiveTemplate(webHookSetting, msg)
	if !ok {
		logger.Error("#webHook#通知模板配置为空")
		return
	}
	logger.Debugf("#webHook#发送企微群数量-%d", len(groups))
	msg["name"] = utils.EscapeJson(msg["name"].(string))
	msg["output"] = utils.EscapeJson(msg["output"].(string))
	msg["content"] = parseNotifyTemplate(template.Content, msg)
	msg["content"] = html.UnescapeString(msg["content"].(string))
	for _, group := range groups {
		webHook.send(msg, group.Url)
	}
}

func (webHook *WebHook) getActiveGroups(webHookSetting models.WebHook, msg Message) []models.WebHookGroup {
	receiverIds, _ := msg["task_receiver_id"].(string)
	if strings.TrimSpace(receiverIds) == "" {
		if len(webHookSetting.Groups) == 0 {
			return []models.WebHookGroup{}
		}
		return webHookSetting.Groups[:1]
	}

	activeGroups := make([]models.WebHookGroup, 0)
	selectedIds := strings.Split(receiverIds, ",")
	if target, err := models.DecodeWebhookTarget(receiverIds); err == nil {
		selectedIds = make([]string, 0, len(target.GroupIds))
		for _, id := range target.GroupIds {
			selectedIds = append(selectedIds, strconv.Itoa(id))
		}
	}
	for _, group := range webHookSetting.Groups {
		if utils.InStringSlice(selectedIds, strconv.Itoa(group.Id)) {
			activeGroups = append(activeGroups, group)
		}
	}

	return activeGroups
}

func (webHook *WebHook) getActiveTemplate(webHookSetting models.WebHook, msg Message) (models.WebHookTemplate, bool) {
	receiverIds, _ := msg["task_receiver_id"].(string)
	if target, err := models.DecodeWebhookTarget(receiverIds); err == nil && target.TemplateId > 0 {
		for _, template := range webHookSetting.Templates {
			if template.Id == target.TemplateId {
				return template, true
			}
		}
	}
	if len(webHookSetting.Templates) > 0 {
		return webHookSetting.Templates[0], true
	}
	if webHookSetting.Template != "" {
		return models.WebHookTemplate{Content: webHookSetting.Template}, true
	}

	return models.WebHookTemplate{}, false
}

func (webHook *WebHook) send(msg Message, url string) {
	content := msg["content"].(string)
	timeout := 30
	maxTimes := 3
	i := 0
	for i < maxTimes {
		resp := httpclient.PostJson(url, content, timeout)
		if resp.StatusCode == 200 {
			break
		}
		i += 1
		time.Sleep(2 * time.Second)
		if i < maxTimes {
			logger.Errorf("webHook#发送消息失败#%s#消息内容-%s", resp.Body, msg["content"])
		}
	}
}
