package rules

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/vmaltarello/whyd/internal/journal"
	builtin "github.com/vmaltarello/whyd/rules"
)

func loadBuiltin(t *testing.T) *Set {
	t.Helper()
	s, err := LoadDefault()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// Every [[rule.example]] of the built-in rules.
func TestBuiltinExamples(t *testing.T) {
	s := loadBuiltin(t)
	if len(s.Rules) < 15 {
		t.Fatalf("only %d built-in rules loaded", len(s.Rules))
	}
	for _, r := range s.Rules {
		if len(r.Examples) == 0 {
			t.Errorf("%s has no example", r.ID)
		}
	}
	for _, f := range s.Test() {
		t.Error(f)
	}
}

func TestUserRuleOverrides(t *testing.T) {
	user := fstest.MapFS{
		"mine.toml": {Data: []byte(`
[[rule]]
id = "noise.firewall"
kind = "noise"
category = "noise"
severity = "noise"
title = "nothing"
message = "^never$"

[[rule]]
id = "mine.broken"
kind = "match"
category = "x"
severity = "loud"
title = "broken"
message = "("
`)},
		"garbage.toml": {Data: []byte("this is [ not toml")},
	}
	s, err := Load(Source{Name: "built-in", FS: builtin.Builtin, Strict: true}, Source{Name: "/home/u/.config/whyd/rules", FS: user})
	if err != nil {
		t.Fatalf("a broken user file must not stop loading: %v", err)
	}
	r := s.Rule("noise.firewall")
	if r.Title != "nothing" || r.Overrides != "built-in" || r.Origin != "/home/u/.config/whyd/rules/mine.toml" {
		t.Errorf("override: title %q, overrides %q, origin %q", r.Title, r.Overrides, r.Origin)
	}
	if s.Rule("mine.broken") != nil || len(s.Problems) != 2 {
		t.Errorf("broken rule and file should be skipped and reported, problems: %q", s.Problems)
	}
	if _, err := Load(Source{Name: "built-in", FS: user, Strict: true}); err == nil {
		t.Error("a broken strict source must fail")
	}
}

func TestEngineOnRealFreeze(t *testing.T) {
	e := NewEngine(loadBuiltin(t))
	noise := 0
	err := journal.ReadFile("../../testdata/boot-unclean-oom.jsonl.gz", func(en journal.Entry) error {
		if isNoise, _ := e.Feed(en); isNoise {
			noise++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	events, noiseCounts := e.Finish()

	byRule := map[string]*Event{}
	for _, ev := range events {
		if byRule[ev.Rule.ID] == nil {
			byRule[ev.Rule.ID] = ev
		}
	}
	oom := byRule["memory.oom-report"]
	if oom == nil {
		t.Fatal("OOM not recognised")
	}
	if oom.Count != 4 || oom.Title != "out of memory, 4 processes killed" {
		t.Errorf("OOM event: count %d, title %q", oom.Count, oom.Title)
	}
	if got := oom.First.Format("15:04:05"); got != "10:53:10" {
		t.Errorf("OOM first at %s", got)
	}
	if byRule["memory.oom-kill"] != nil {
		t.Error("kill lines counted twice (also outside the report)")
	}
	if ev := byRule["crash.coredump"]; ev == nil || ev.Title != "sunshine killed by SIGTRAP" {
		t.Errorf("crash event: %+v", ev)
	}
	if ev := byRule["known.rtw88-tx"]; ev == nil || ev.Count != 149 {
		t.Errorf("rtw88 event: %+v", ev)
	}
	if ev := byRule["power.suspend"]; ev == nil {
		t.Error("suspend not recognised")
	}
	if noise < 437 || len(noiseCounts) == 0 || noiseCounts[0].Rule.ID != "noise.firewall" || noiseCounts[0].Count != 437 {
		t.Errorf("noise: %d lines, counts %+v", noise, noiseCounts)
	}
	for _, ev := range events {
		if strings.Contains(ev.Title, "{") {
			t.Errorf("unfilled title %q", ev.Title)
		}
	}
}

// A user rule whose example an earlier built-in rule takes would never fire: Test says so.
func TestShadowedRule(t *testing.T) {
	user := fstest.MapFS{"mine.toml": {Data: []byte(`
[[rule]]
id = "mine.oom"
kind = "match"
category = "memory"
severity = "warning"
title = "oom"
source = "kernel"
message = "Out of memory"
[[rule.example]]
match = "Out of memory: Killed process 1 (x)"
`)}}
	s, err := Load(Source{Name: "built-in", FS: builtin.Builtin, Strict: true}, Source{Name: "u", FS: user})
	if err != nil {
		t.Fatal(err)
	}
	fails := s.Test()
	if len(fails) != 1 || !strings.Contains(fails[0], "taken by memory.oom-kill") {
		t.Errorf("fails %q", fails)
	}
}

// A mistake in a rule is reported with what to write instead, never silently ignored.
func TestRuleMistakes(t *testing.T) {
	base := "id = \"mine.x\"\nkind = \"match\"\ncategory = \"x\"\nseverity = \"warning\"\ntitle = \"x\"\nmessage = 'x'\n"
	cases := map[string]string{
		"sorce = \"kernel\"":               `unknown field "sorce", did you mean "source"?`,
		"merge = \"rules\"":                `merge must be "rule" or left out`,
		"priority = \"<=12\"":              `priority must be "<=N" with N from 0 to 7`,
		"priority = \"4\"":                 `priority must be "<=N"`,
		"extract = ['(?P<a>x)']":           "extract has no effect in a match rule",
		"start = 'x'":                      "start has no effect in a match rule",
		"merge_window = \"10 minutes\"":    "merge_window must be a duration",
		"note = \"{proc}\"":                "note uses {proc}, but no pattern has a group named proc",
		"merge_title = \"{count} times\"":  "",
		"[[rule.example]]\nmtch = \"x\"":   `unknown field "mtch" in [[rule.example]]`,
		"[[rule.example]]\nnomatch = \"\"": "each [[rule.example]] needs either match or nomatch",
		"severity = \"noise\"\n#":          `severity "noise" is for kind = "noise"`,
	}
	for extra, want := range cases {
		text := base + extra
		if strings.HasPrefix(extra, "severity") { // replace, not add
			text = strings.Replace(base, "severity = \"warning\"\n", "", 1) + extra
		}
		s, err := Load(Source{Name: "u", FS: fstest.MapFS{"r.toml": {Data: []byte("[[rule]]\n" + text)}}})
		if err != nil {
			t.Fatal(err)
		}
		got := strings.Join(s.Problems, "\n")
		if want == "" && got != "" || want != "" && !strings.Contains(got, want) {
			t.Errorf("%s\n got %q\nwant %q", extra, got, want)
		}
	}
	for text, want := range map[string]string{
		"id = \"g\"\nkind = \"group\"\ncategory = \"x\"\nseverity = \"info\"\ntitle = \"g\"\nstart = 'a'":   "missing end",
		"id = \"g\"\nkind = \"group\"\ncategory = \"x\"\nseverity = \"info\"\ntitle = \"g\"\nmessage = 'a'": "message has no effect in a group rule",
		"id = \"n\"\nkind = \"noise\"\ntitle = \"n\"\nmessage = 'a'\nseverity = \"info\"":                   `severity must be "noise" in a noise rule`,
		"id = \"n\"\nkind = \"noise\"\ntitle = \"n\"\nmessage = 'a'":                                        "",
		"id = \"m\"\nkind = \"match\"\nseverity = \"info\"\ntitle = \"m\"\nmessage = 'a'":                   "missing category",
		"id = \"m\"\ncategory = \"x\"\nseverity = \"info\"\ntitle = \"m\"\nmessage = 'a'":                   "missing kind",
	} {
		s, err := Load(Source{Name: "u", FS: fstest.MapFS{"r.toml": {Data: []byte("[[rule]]\n" + text)}}})
		if err != nil {
			t.Fatal(err)
		}
		got := strings.Join(s.Problems, "\n")
		if want == "" && got != "" || want != "" && !strings.Contains(got, want) {
			t.Errorf("%s\n got %q\nwant %q", text, got, want)
		}
	}
}
