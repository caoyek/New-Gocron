package models

import (
	"encoding/json"
	"errors"
	"strings"
)

type Setting struct {
	Id    int    `xorm:"int pk autoincr"`
	Code  string `xorm:"varchar(32) notnull"`
	Key   string `xorm:"varchar(64) notnull"`
	Value string `xorm:"varchar(4096) notnull default '' "`
}

const slackTemplate = `
任务ID:  {{.TaskId}}
任务名称: {{.TaskName}}
状态:    {{.Status}}
执行结果: {{.Result}}
备注: {{.Remark}}
`
const emailTemplate = `
任务ID:  {{.TaskId}}
任务名称: {{.TaskName}}
状态:    {{.Status}}
执行结果: {{.Result}}
备注: {{.Remark}}
`
const webhookTemplate = `
{
  "task_id": "{{.TaskId}}",
  "task_name": "{{.TaskName}}",
  "status": "{{.Status}}",
  "result": "{{.Result}}",
  "remark": "{{.Remark}}"
}
`

const (
	SlackCode        = "slack"
	SlackUrlKey      = "url"
	SlackTemplateKey = "template"
	SlackChannelKey  = "channel"
)

const (
	MailCode        = "mail"
	MailTemplateKey = "template"
	MailServerKey   = "server"
	MailUserKey     = "user"
)

const (
	WebhookCode              = "webhook"
	WebhookTemplateKey       = "template"
	WebhookTemplateKeyPrefix = "template:"
	WebhookUrlKey            = "url"
	WebhookGroupKey          = "group"
)

const (
	LoginSecurityCode      = "login_security"
	LoginSecurityPolicyKey = "policy"
)

type LoginSecurityPolicy struct {
	BlockEnabled     bool     `json:"block_enabled"`
	WindowMinutes    int      `json:"window_minutes"`
	MaxFailures      int      `json:"max_failures"`
	BlockMinutes     int      `json:"block_minutes"`
	WhitelistEnabled bool     `json:"whitelist_enabled"`
	Whitelist        []string `json:"whitelist"`
}

func DefaultLoginSecurityPolicy() LoginSecurityPolicy {
	return LoginSecurityPolicy{
		BlockEnabled:  true,
		WindowMinutes: 10,
		MaxFailures:   5,
		BlockMinutes:  30,
		Whitelist:     make([]string, 0),
	}
}

// 初始化基本字段 邮件、slack等
func (setting *Setting) InitBasicField() {
	setting.Code = SlackCode
	setting.Key = SlackUrlKey
	setting.Value = ""
	Db.Insert(setting)
	setting.Id = 0

	setting.Code = SlackCode
	setting.Key = SlackTemplateKey
	setting.Value = slackTemplate
	Db.Insert(setting)
	setting.Id = 0

	setting.Code = MailCode
	setting.Key = MailServerKey
	setting.Value = ""
	Db.Insert(setting)
	setting.Id = 0

	setting.Code = MailCode
	setting.Key = MailTemplateKey
	setting.Value = emailTemplate
	Db.Insert(setting)
	setting.Id = 0

	setting.Code = WebhookCode
	setting.Key = WebhookTemplateKey
	setting.Value = webhookTemplate
	Db.Insert(setting)
	setting.Id = 0

	setting.Code = WebhookCode
	setting.Key = WebhookUrlKey
	setting.Value = ""
	Db.Insert(setting)
	setting.Id = 0

	policy, _ := json.Marshal(DefaultLoginSecurityPolicy())
	setting.Code = LoginSecurityCode
	setting.Key = LoginSecurityPolicyKey
	setting.Value = string(policy)
	Db.Insert(setting)
}

// region slack配置

type Slack struct {
	Url      string    `json:"url"`
	Channels []Channel `json:"channels"`
	Template string    `json:"template"`
}

type Channel struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (setting *Setting) Slack() (Slack, error) {
	list := make([]Setting, 0)
	err := Db.Where("code = ?", SlackCode).Find(&list)
	slack := Slack{}
	if err != nil {
		return slack, err
	}

	setting.formatSlack(list, &slack)

	return slack, err
}

func (setting *Setting) formatSlack(list []Setting, slack *Slack) {
	for _, v := range list {
		switch v.Key {
		case SlackUrlKey:
			slack.Url = v.Value
		case SlackTemplateKey:
			slack.Template = v.Value
		default:
			slack.Channels = append(slack.Channels, Channel{
				v.Id, v.Value,
			})
		}
	}
}

func (setting *Setting) UpdateSlack(url, template string) error {
	setting.Value = url

	Db.Cols("value").Update(setting, Setting{Code: SlackCode, Key: SlackUrlKey})

	setting.Value = template
	Db.Cols("value").Update(setting, Setting{Code: SlackCode, Key: SlackTemplateKey})

	return nil
}

// 创建slack渠道
func (setting *Setting) CreateChannel(channel string) (int64, error) {
	setting.Code = SlackCode
	setting.Key = SlackChannelKey
	setting.Value = channel

	return Db.Insert(setting)
}

func (setting *Setting) IsChannelExist(channel string) bool {
	setting.Code = SlackCode
	setting.Key = SlackChannelKey
	setting.Value = channel

	count, _ := Db.Count(setting)

	return count > 0
}

// 删除slack渠道
func (setting *Setting) RemoveChannel(id int) (int64, error) {
	setting.Code = SlackCode
	setting.Key = SlackChannelKey
	setting.Id = id
	return Db.Delete(setting)
}

// endregion

type Mail struct {
	Host      string     `json:"host"`
	Port      int        `json:"port"`
	User      string     `json:"user"`
	Password  string     `json:"password"`
	MailUsers []MailUser `json:"mail_users"`
	Template  string     `json:"template"`
}

type MailUser struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// region 邮件配置
func (setting *Setting) Mail() (Mail, error) {
	list := make([]Setting, 0)
	err := Db.Where("code = ?", MailCode).Find(&list)
	mail := Mail{MailUsers: make([]MailUser, 0)}
	if err != nil {
		return mail, err
	}

	setting.formatMail(list, &mail)

	return mail, err
}

func (setting *Setting) formatMail(list []Setting, mail *Mail) {
	mailUser := MailUser{}
	for _, v := range list {
		switch v.Key {
		case MailServerKey:
			json.Unmarshal([]byte(v.Value), mail)
		case MailUserKey:
			json.Unmarshal([]byte(v.Value), &mailUser)
			mailUser.Id = v.Id
			mail.MailUsers = append(mail.MailUsers, mailUser)
		case MailTemplateKey:
			mail.Template = v.Value
		}

	}
}

func (setting *Setting) UpdateMail(config, template string) error {
	setting.Value = config
	Db.Cols("value").Update(setting, Setting{Code: MailCode, Key: MailServerKey})

	setting.Value = template
	Db.Cols("value").Update(setting, Setting{Code: MailCode, Key: MailTemplateKey})

	return nil
}

func (setting *Setting) CreateMailUser(username, email string) (int64, error) {
	setting.Code = MailCode
	setting.Key = MailUserKey
	mailUser := MailUser{0, username, email}
	jsonByte, err := json.Marshal(mailUser)
	if err != nil {
		return 0, err
	}
	setting.Value = string(jsonByte)

	return Db.Insert(setting)
}

func (setting *Setting) RemoveMailUser(id int) (int64, error) {
	setting.Code = MailCode
	setting.Key = MailUserKey
	setting.Id = id
	return Db.Delete(setting)
}

type WebHook struct {
	Url       string            `json:"url"`
	Groups    []WebHookGroup    `json:"groups"`
	Template  string            `json:"template"`
	Templates []WebHookTemplate `json:"templates"`
}

type WebHookGroup struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type WebHookTemplate struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	IsDefault bool   `json:"is_default"`
}

type WebhookTarget struct {
	Format     string `json:"format"`
	Version    int    `json:"version"`
	GroupIds   []int  `json:"group_ids"`
	TemplateId int    `json:"template_id"`
}

func DecodeWebhookTarget(value string) (WebhookTarget, error) {
	target := WebhookTarget{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(value)), &target); err != nil {
		return target, err
	}
	if target.Format != "webhook_target" || target.Version != 1 {
		return WebhookTarget{}, errors.New("企微通知配置格式无效")
	}

	return target, nil
}

func (webHook WebHook) ValidateTarget(target WebhookTarget) error {
	if len(target.GroupIds) == 0 {
		return errors.New("至少选择一个企微群")
	}
	if target.TemplateId <= 0 {
		return errors.New("请选择通知模板")
	}

	groups := make(map[int]struct{}, len(webHook.Groups))
	for _, group := range webHook.Groups {
		groups[group.Id] = struct{}{}
	}
	for _, id := range target.GroupIds {
		if id <= 0 {
			return errors.New("企微群配置不存在，请重新选择")
		}
		if _, ok := groups[id]; !ok {
			return errors.New("企微群配置不存在，请重新选择")
		}
	}

	for _, template := range webHook.Templates {
		if template.Id == target.TemplateId {
			return nil
		}
	}

	return errors.New("通知模板不存在，请重新选择")
}

func (setting *Setting) Webhook() (WebHook, error) {
	list := make([]Setting, 0)
	err := Db.Where("code = ?", WebhookCode).Asc("id").Find(&list)
	webHook := WebHook{
		Groups:    make([]WebHookGroup, 0),
		Templates: make([]WebHookTemplate, 0),
	}
	if err != nil {
		return webHook, err
	}

	err = setting.formatWebhook(list, &webHook)

	return webHook, err
}

func (setting *Setting) formatWebhook(list []Setting, webHook *WebHook) error {
	for _, v := range list {
		switch {
		case v.Key == WebhookUrlKey:
			webHook.Url = v.Value
			if strings.TrimSpace(v.Value) != "" {
				webHook.Groups = append(webHook.Groups, WebHookGroup{
					Id:   v.Id,
					Name: "默认企微群",
					Url:  v.Value,
				})
			}
		case v.Key == WebhookGroupKey:
			group := WebHookGroup{}
			if err := json.Unmarshal([]byte(v.Value), &group); err != nil {
				return err
			}
			group.Id = v.Id
			webHook.Groups = append(webHook.Groups, group)
		case v.Key == WebhookTemplateKey:
			webHook.Template = v.Value
			webHook.Templates = append(webHook.Templates, WebHookTemplate{
				Id:        v.Id,
				Name:      "默认模板",
				Content:   v.Value,
				IsDefault: true,
			})
		case strings.HasPrefix(v.Key, WebhookTemplateKeyPrefix):
			name := strings.TrimSpace(strings.TrimPrefix(v.Key, WebhookTemplateKeyPrefix))
			if name != "" {
				webHook.Templates = append(webHook.Templates, WebHookTemplate{
					Id:      v.Id,
					Name:    name,
					Content: v.Value,
				})
			}
		}
	}
	for i, template := range webHook.Templates {
		if template.IsDefault && i > 0 {
			ordered := make([]WebHookTemplate, 0, len(webHook.Templates))
			ordered = append(ordered, template)
			ordered = append(ordered, webHook.Templates[:i]...)
			ordered = append(ordered, webHook.Templates[i+1:]...)
			webHook.Templates = ordered
			break
		}
	}
	if webHook.Template == "" && len(webHook.Templates) > 0 {
		webHook.Template = webHook.Templates[0].Content
	}

	return nil
}

func (setting *Setting) UpdateDefaultWebhookTemplate(template string) error {
	setting.Value = template
	updated, err := Db.Cols("value").Update(setting, Setting{Code: WebhookCode, Key: WebhookTemplateKey})
	if err != nil || updated > 0 {
		return err
	}

	setting.Id = 0
	setting.Code = WebhookCode
	setting.Key = WebhookTemplateKey
	setting.Value = template
	_, err = Db.Insert(setting)

	return err
}

func (setting *Setting) CreateWebhookTemplate(name, content string) (int64, error) {
	setting.Id = 0
	setting.Code = WebhookCode
	setting.Key = WebhookTemplateKeyPrefix + strings.TrimSpace(name)
	setting.Value = content

	return Db.Insert(setting)
}

func (setting *Setting) UpdateWebhookTemplate(id int, name, content string) (int64, error) {
	stored := Setting{}
	has, err := Db.ID(id).Where("code = ?", WebhookCode).Get(&stored)
	if err != nil || !has {
		return 0, err
	}
	if stored.Key != WebhookTemplateKey && !strings.HasPrefix(stored.Key, WebhookTemplateKeyPrefix) {
		return 0, nil
	}

	setting.Value = content
	if stored.Key == WebhookTemplateKey {
		return Db.ID(id).Where("code = ?", WebhookCode).Cols("value").Update(setting)
	}
	setting.Key = WebhookTemplateKeyPrefix + strings.TrimSpace(name)

	return Db.ID(id).Where("code = ?", WebhookCode).Cols("key", "value").Update(setting)
}

func (setting *Setting) RemoveWebhookTemplate(id int) (int64, error) {
	stored := Setting{}
	has, err := Db.ID(id).Where("code = ?", WebhookCode).Get(&stored)
	if err != nil || !has || !strings.HasPrefix(stored.Key, WebhookTemplateKeyPrefix) {
		return 0, err
	}

	return Db.ID(id).Where("code = ?", WebhookCode).Delete(new(Setting))
}

func (setting *Setting) WebhookTemplateNameExists(name string, excludeId int) (bool, error) {
	webHook, err := setting.Webhook()
	if err != nil {
		return false, err
	}
	for _, template := range webHook.Templates {
		if template.Id != excludeId && strings.EqualFold(strings.TrimSpace(template.Name), strings.TrimSpace(name)) {
			return true, nil
		}
	}

	return false, nil
}

func (setting *Setting) CreateWebhookGroup(name, url string) (int64, error) {
	group := WebHookGroup{Name: name, Url: url}
	value, err := json.Marshal(group)
	if err != nil {
		return 0, err
	}

	setting.Id = 0
	setting.Code = WebhookCode
	setting.Key = WebhookGroupKey
	setting.Value = string(value)

	return Db.Insert(setting)
}

func (setting *Setting) UpdateWebhookGroup(id int, name, url string) (int64, error) {
	group := WebHookGroup{Name: name, Url: url}
	value, err := json.Marshal(group)
	if err != nil {
		return 0, err
	}

	setting.Key = WebhookGroupKey
	setting.Value = string(value)

	return Db.ID(id).
		Where("code = ?", WebhookCode).
		In("key", WebhookUrlKey, WebhookGroupKey).
		Cols("key", "value").
		Update(setting)
}

func (setting *Setting) RemoveWebhookGroup(id int) (int64, error) {
	return Db.ID(id).
		Where("code = ?", WebhookCode).
		In("key", WebhookUrlKey, WebhookGroupKey).
		Delete(new(Setting))
}

func (setting *Setting) WebhookGroupNameExists(name string, excludeId int) (bool, error) {
	webHook, err := setting.Webhook()
	if err != nil {
		return false, err
	}
	for _, group := range webHook.Groups {
		if group.Id != excludeId && strings.EqualFold(strings.TrimSpace(group.Name), strings.TrimSpace(name)) {
			return true, nil
		}
	}

	return false, nil
}

func (setting *Setting) LoginSecurity() (LoginSecurityPolicy, error) {
	policy := DefaultLoginSecurityPolicy()
	stored, has, err := findLoginSecuritySetting()
	if err != nil || !has || stored.Value == "" {
		return policy, err
	}
	if err = json.Unmarshal([]byte(stored.Value), &policy); err != nil {
		return DefaultLoginSecurityPolicy(), err
	}
	if policy.WindowMinutes <= 0 || policy.MaxFailures <= 0 || policy.BlockMinutes <= 0 {
		return DefaultLoginSecurityPolicy(), errors.New("登录安全配置无效")
	}
	if policy.Whitelist == nil {
		policy.Whitelist = make([]string, 0)
	}

	return policy, nil
}

func (setting *Setting) UpdateLoginSecurity(policy LoginSecurityPolicy) error {
	value, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	stored, has, err := findLoginSecuritySetting()
	if err != nil {
		return err
	}
	if has {
		stored.Value = string(value)
		_, err = Db.ID(stored.Id).Cols("value").Update(stored)
		return err
	}

	setting.Code = LoginSecurityCode
	setting.Key = LoginSecurityPolicyKey
	setting.Value = string(value)
	_, err = Db.Insert(setting)

	return err
}

func findLoginSecuritySetting() (*Setting, bool, error) {
	list := make([]Setting, 0)
	if err := Db.Where("code = ?", LoginSecurityCode).Find(&list); err != nil {
		return nil, false, err
	}
	for i := range list {
		if list[i].Key == LoginSecurityPolicyKey {
			return &list[i], true, nil
		}
	}

	return new(Setting), false, nil
}

// endregion
