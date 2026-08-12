package service

import (
	"errors"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/caoyek/New-Gocron/internal/models"
	macaron "gopkg.in/macaron.v1"
)

var ErrLoginBlocked = errors.New("登录已被临时封禁")

type LoginBlockStatus struct {
	BlockedUntil time.Time
}

func RequestPeerIP(ctx *macaron.Context) string {
	address := strings.TrimSpace(ctx.Req.RemoteAddr)
	host, _, err := net.SplitHostPort(address)
	if err == nil {
		return strings.Trim(host, "[]")
	}

	return strings.Trim(address, "[]")
}

func NormalizeWhitelist(raw string) ([]string, error) {
	items := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == ';'
	})
	unique := make(map[string]struct{})
	result := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		if ip := net.ParseIP(value); ip != nil {
			value = ip.String()
		} else {
			_, network, err := net.ParseCIDR(value)
			if err != nil {
				return nil, errors.New("白名单包含无效的 IP 或 CIDR: " + value)
			}
			value = network.String()
		}
		if _, exists := unique[value]; exists {
			continue
		}
		unique[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)

	return result, nil
}

func IPAllowed(ip string, whitelist []string) bool {
	parsedIP := net.ParseIP(strings.TrimSpace(ip))
	if parsedIP == nil {
		return false
	}
	for _, entry := range whitelist {
		if allowedIP := net.ParseIP(entry); allowedIP != nil {
			if allowedIP.Equal(parsedIP) {
				return true
			}
			continue
		}
		_, network, err := net.ParseCIDR(entry)
		if err == nil && network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

func LoginSecurityPolicy() (models.LoginSecurityPolicy, error) {
	return new(models.Setting).LoginSecurity()
}

func ActiveLoginBlock(ip, username string, now time.Time) (*models.LoginBlock, error) {
	blockModel := new(models.LoginBlock)
	if ip != "" {
		block, err := blockModel.Active("ip", ip, now)
		if err != nil || block != nil {
			return block, err
		}
	}
	if username != "" {
		return blockModel.Active("account", normalizeUsername(username), now)
	}

	return nil, nil
}

func RecordLoginEvent(username, ip, result, message string) error {
	event := &models.LoginSecurityEvent{
		Username: normalizeUsername(username),
		Ip:       ip,
		Result:   result,
		Message:  message,
	}

	return event.Create()
}

func RecordLoginFailure(policy models.LoginSecurityPolicy, username, ip string, now time.Time) (*models.LoginBlock, error) {
	username = normalizeUsername(username)
	if err := RecordLoginEvent(username, ip, models.LoginResultPasswordFailure, "用户名或密码错误"); err != nil {
		return nil, err
	}
	if !policy.BlockEnabled {
		return nil, nil
	}

	since := now.Add(-time.Duration(policy.WindowMinutes) * time.Minute)
	eventModel := new(models.LoginSecurityEvent)
	blockedUntil := now.Add(time.Duration(policy.BlockMinutes) * time.Minute)
	var firstBlock *models.LoginBlock
	for _, target := range []struct {
		scope string
		value string
	}{
		{"ip", ip},
		{"account", username},
	} {
		if target.value == "" {
			continue
		}
		latestSuccess, err := eventModel.LatestSuccess(target.scope, target.value)
		if err != nil {
			return nil, err
		}
		targetSince := since
		if latestSuccess.After(targetSince) {
			targetSince = latestSuccess
		}
		failures, err := eventModel.CountPasswordFailures(target.scope, target.value, targetSince)
		if err != nil {
			return nil, err
		}
		if failures < int64(policy.MaxFailures) {
			continue
		}
		if err = new(models.LoginBlock).Upsert(target.scope, target.value, blockedUntil); err != nil {
			return nil, err
		}
		if firstBlock == nil {
			firstBlock = &models.LoginBlock{
				Scope:        target.scope,
				Value:        target.value,
				BlockedUntil: blockedUntil,
			}
		}
	}

	return firstBlock, nil
}

func normalizeUsername(username string) string {
	const maxUsernameLength = 64

	normalized := []rune(strings.ToLower(strings.TrimSpace(username)))
	if len(normalized) > maxUsernameLength {
		normalized = normalized[:maxUsernameLength]
	}

	return string(normalized)
}
