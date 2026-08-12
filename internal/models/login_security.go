package models

import "time"

const (
	LoginResultSuccess           = "success"
	LoginResultPasswordFailure   = "password_failure"
	LoginResultBlocked           = "blocked"
	LoginResultWhitelistRejected = "whitelist_rejected"
)

type LoginSecurityEvent struct {
	Id        int       `json:"id" xorm:"pk autoincr notnull"`
	Username  string    `json:"username" xorm:"varchar(64) notnull default ''"`
	Ip        string    `json:"ip" xorm:"varchar(45) notnull default '' index"`
	Result    string    `json:"result" xorm:"varchar(32) notnull index"`
	Message   string    `json:"message" xorm:"varchar(255) notnull default ''"`
	Created   time.Time `json:"created" xorm:"datetime notnull created index"`
	BaseModel `json:"-" xorm:"-"`
}

type LoginBlock struct {
	Id           int       `json:"id" xorm:"pk autoincr notnull"`
	Scope        string    `json:"scope" xorm:"varchar(16) notnull index"`
	Value        string    `json:"value" xorm:"varchar(128) notnull index"`
	BlockedUntil time.Time `json:"blocked_until" xorm:"datetime notnull index"`
	Created      time.Time `json:"created" xorm:"datetime notnull created"`
	Updated      time.Time `json:"updated" xorm:"datetime updated"`
}

func (event *LoginSecurityEvent) Create() error {
	_, err := Db.Insert(event)

	return err
}

func (event *LoginSecurityEvent) List(params CommonMap) ([]LoginSecurityEvent, error) {
	event.parsePageAndPageSize(params)
	list := make([]LoginSecurityEvent, 0)
	err := Db.Desc("id").Limit(event.PageSize, event.pageLimitOffset()).Find(&list)

	return list, err
}

func (event *LoginSecurityEvent) Total() (int64, error) {
	return Db.Count(event)
}

func (event *LoginSecurityEvent) CountPasswordFailures(scope, value string, since time.Time) (int64, error) {
	session := Db.Where("result = ? AND created >= ?", LoginResultPasswordFailure, since)
	if scope == "ip" {
		session.And("ip = ?", value)
	} else {
		session.And("username = ?", value)
	}

	return session.Count(event)
}

func (event *LoginSecurityEvent) LatestSuccess(scope, value string) (time.Time, error) {
	latest := new(LoginSecurityEvent)
	session := Db.Where("result = ?", LoginResultSuccess)
	if scope == "ip" {
		session.And("ip = ?", value)
	} else {
		session.And("username = ?", value)
	}
	has, err := session.Desc("id").Get(latest)
	if err != nil || !has {
		return time.Time{}, err
	}

	return latest.Created, nil
}

func (block *LoginBlock) Active(scope, value string, now time.Time) (*LoginBlock, error) {
	active := new(LoginBlock)
	has, err := Db.Where("scope = ? AND value = ? AND blocked_until > ?", scope, value, now).Get(active)
	if err != nil || !has {
		return nil, err
	}

	return active, nil
}

func (block *LoginBlock) Upsert(scope, value string, blockedUntil time.Time) error {
	existing := new(LoginBlock)
	has, err := Db.Where("scope = ? AND value = ?", scope, value).Get(existing)
	if err != nil {
		return err
	}
	if has {
		existing.BlockedUntil = blockedUntil
		_, err = Db.ID(existing.Id).Cols("blocked_until").Update(existing)
		return err
	}

	block.Scope = scope
	block.Value = value
	block.BlockedUntil = blockedUntil
	_, err = Db.Insert(block)

	return err
}

func (block *LoginBlock) ListActive(now time.Time) ([]LoginBlock, error) {
	list := make([]LoginBlock, 0)
	err := Db.Where("blocked_until > ?", now).Asc("blocked_until").Find(&list)

	return list, err
}

func (block *LoginBlock) Remove(id int) error {
	_, err := Db.ID(id).Delete(new(LoginBlock))

	return err
}
