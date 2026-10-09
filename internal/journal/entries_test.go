package journal

import (
	"strings"
	"testing"
)

func TestReadFile(t *testing.T) {
	var n, bytesMsgs, kernel int
	var first, last Entry
	err := ReadFile("../../testdata/boot-unclean-oom.jsonl.gz", func(e Entry) error {
		if n == 0 {
			first = e
		}
		last = e
		n++
		if e.Kernel {
			kernel++
		}
		if strings.HasPrefix(e.Message, "[UFW BLOCK]") && e.Source != "kernel" {
			t.Errorf("UFW line with source %q", e.Source)
		}
		if strings.ContainsRune(e.Message, '�') {
			bytesMsgs++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 6216 {
		t.Errorf("read %d entries, want 6216", n)
	}
	if kernel == 0 {
		t.Error("no kernel entries")
	}
	if !strings.HasPrefix(first.Message, "Linux version 7.2.9-arch1-1") {
		t.Errorf("first message = %q", first.Message)
	}
	if got := last.Time.Format("15:04:05"); got != "12:11:56" {
		t.Errorf("last entry at %s, want 12:11:56", got)
	}
	if first.Priority < 0 || first.Source == "" {
		t.Errorf("first entry not decoded: %+v", first)
	}
}

func TestFieldValue(t *testing.T) {
	cases := map[string]string{
		`"plain"`:         "plain",
		`[104,105]`:       "hi",  // byte array (not valid UTF-8 in general)
		`[104,255]`:       "h�",  // invalid UTF-8 is replaced, not dropped
		`["one","two"]`:   "one", // repeated field: first value
		`[[111,107],"x"]`: "ok",  // repeated field whose first value is bytes
		`null`:            "",
	}
	for in, want := range cases {
		if got := fieldValue([]byte(in)); got != want {
			t.Errorf("fieldValue(%s) = %q, want %q", in, got, want)
		}
	}
}

func TestReadOneBootOnly(t *testing.T) {
	two := `{"_BOOT_ID":"aa","MESSAGE":"one","__REALTIME_TIMESTAMP":"1"}
{"_BOOT_ID":"aa","MESSAGE":"two","__REALTIME_TIMESTAMP":"2"}
{"_BOOT_ID":"bb","MESSAGE":"three","__REALTIME_TIMESTAMP":"3"}
`
	n := 0
	err := Read(strings.NewReader(two), func(Entry) error { n++; return nil })
	if err == nil || !strings.Contains(err.Error(), "more than one boot") || n != 2 {
		t.Errorf("err = %v after %d entries", err, n)
	}
}
