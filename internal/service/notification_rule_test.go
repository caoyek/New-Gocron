package service

import "testing"

func TestMatchNotificationRulesLegacyKeyword(t *testing.T) {
	matched, err := MatchNotificationRules("task finished successfully", "finished")
	if err != nil || !matched {
		t.Fatalf("expected legacy keyword to match, matched=%v err=%v", matched, err)
	}
	matched, err = MatchNotificationRules("task FINISHED successfully", "finished")
	if err != nil || matched {
		t.Fatalf("expected legacy keyword to remain case-sensitive, matched=%v err=%v", matched, err)
	}
	matched, err = MatchNotificationRules("result: {error", "{error")
	if err != nil || !matched {
		t.Fatalf("expected legacy keyword beginning with brace to match, matched=%v err=%v", matched, err)
	}
}

func TestMatchNotificationRulesAnyAndAll(t *testing.T) {
	any := `{"format":"notification_rules","version":1,"mode":"any","rules":[{"type":"contains","value":"failed"},{"type":"contains","value":"warning"}]}`
	matched, err := MatchNotificationRules("WARNING: retrying", any)
	if err != nil || !matched {
		t.Fatalf("expected any rule to match case-insensitively, matched=%v err=%v", matched, err)
	}

	all := `{"format":"notification_rules","version":1,"mode":"all","rules":[{"type":"contains","value":"warning"},{"type":"not_contains","value":"failed"}]}`
	matched, err = MatchNotificationRules("WARNING: retrying", all)
	if err != nil || !matched {
		t.Fatalf("expected all rules to match, matched=%v err=%v", matched, err)
	}
}

func TestMatchNotificationRulesWildcardAndRegexp(t *testing.T) {
	wildcard := `{"format":"notification_rules","version":1,"mode":"any","rules":[{"type":"wildcard","value":"order*failed"}]}`
	matched, err := MatchNotificationRules("Order\n1528 FAILED", wildcard)
	if err != nil || !matched {
		t.Fatalf("expected wildcard rule to match, matched=%v err=%v", matched, err)
	}

	regularExpression := `{"format":"notification_rules","version":1,"mode":"any","rules":[{"type":"regex","value":"库存[：:]?\\s*\\d+"}]}`
	matched, err = MatchNotificationRules("库存：12860", regularExpression)
	if err != nil || !matched {
		t.Fatalf("expected regexp rule to match, matched=%v err=%v", matched, err)
	}
}

func TestMatchNotificationRulesNumber(t *testing.T) {
	rules := `{"format":"notification_rules","version":1,"mode":"any","rules":[{"type":"number","field":"库存数量","operator":">","number":10000}]}`
	for _, output := range []string{"库存数量：12860", "库存数量 = 12860.5", "库存数量 9999\n库存数量: 10001"} {
		matched, err := MatchNotificationRules(output, rules)
		if err != nil || !matched {
			t.Fatalf("expected number rule to match %q, matched=%v err=%v", output, matched, err)
		}
	}

	matched, err := MatchNotificationRules("库存数量：9999", rules)
	if err != nil || matched {
		t.Fatalf("expected number rule not to match, matched=%v err=%v", matched, err)
	}
}

func TestValidateNotificationRulesRejectsInvalidRules(t *testing.T) {
	tests := []string{
		"",
		`{"format":"notification_rules","version":1,"mode":"any","rules":[]}`,
		`{"format":"notification_rules","version":1,"mode":"wrong","rules":[{"type":"contains","value":"x"}]}`,
		`{"format":"notification_rules","version":1,"mode":"any","rules":[{"type":"regex","value":"["}]}`,
		`{"format":"notification_rules","version":1,"mode":"any","rules":[{"type":"number","field":"库存","operator":"~","number":1}]}`,
	}
	for _, value := range tests {
		if _, err := ParseNotificationRules(value); err == nil {
			t.Fatalf("expected invalid rule set to fail: %s", value)
		}
	}
}

func TestParseNotificationRulesKeepsJSONKeywordCompatible(t *testing.T) {
	matched, err := MatchNotificationRules(`response: {"version":1}`, `{"version":1}`)
	if err != nil || !matched {
		t.Fatalf("expected JSON-shaped legacy keyword to match, matched=%v err=%v", matched, err)
	}
}
