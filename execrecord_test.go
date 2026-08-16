package execrecord

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestNewSplitsLines pins the line-splitting contract: empty output must yield a nil
// slice (not []string{""}), trailing newlines are dropped, and leading whitespace is
// kept because indentation is meaningful in command output.
func TestNewSplitsLines(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"empty output yields nil, not one blank line", "", nil},
		{"newline only yields nil", "\n", nil},
		{"whitespace only yields nil", "  \t\n", nil},
		{"trailing newline dropped", "line\n", []string{"line"}},
		{"multiple trailing newlines dropped", "line\n\n\n", []string{"line"}},
		{"no trailing newline", "line", []string{"line"}},
		{"multiple lines", "a\nb\n", []string{"a", "b"}},
		{"leading indentation preserved", "  indented\nnext\n", []string{"  indented", "next"}},
		{"interior blank line preserved", "a\n\nb\n", []string{"a", "", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Stdout and Stderr share a trimming rule. Error is trimmed on both sides
			// instead, so it is covered by TestNewErrorTrimsLeadingWhitespace.
			record := New(Success, "", tt.in, tt.in, time.Now())

			for _, field := range []struct {
				name string
				got  []string
			}{
				{"Stdout", record.Stdout},
				{"Stderr", record.Stderr},
			} {
				if len(field.got) != len(tt.want) {
					t.Errorf("%s = %#v, want %#v", field.name, field.got, tt.want)
					continue
				}
				for i := range tt.want {
					if field.got[i] != tt.want[i] {
						t.Errorf("%s[%d] = %q, want %q", field.name, i, field.got[i], tt.want[i])
					}
				}
			}
		})
	}
}

// TestNewErrorTrimsLeadingWhitespace documents where Error differs from the output
// streams: it is fully trimmed, whereas stdout/stderr keep leading indentation.
func TestNewErrorTrimsLeadingWhitespace(t *testing.T) {
	record := New(GeneralError, "  spaced error\n", "  spaced out\n", "", time.Now())

	if len(record.Error) != 1 || record.Error[0] != "spaced error" {
		t.Errorf("Error = %#v, want [\"spaced error\"]", record.Error)
	}
	if len(record.Stdout) != 1 || record.Stdout[0] != "  spaced out" {
		t.Errorf("Stdout = %#v, want [\"  spaced out\"] — indentation must survive", record.Stdout)
	}
}

// TestNewErrorSplitsLines covers the Error field's own splitting, since
// TestNewSplitsLines exercises only the two streams that share a trimming rule.
func TestNewErrorSplitsLines(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"empty yields nil", "", nil},
		{"whitespace only yields nil", " \n\t ", nil},
		{"single line", "boom\n", []string{"boom"}},
		{"multiple lines", "first\nsecond\n", []string{"first", "second"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(GeneralError, tt.in, "", "", time.Now()).Error
			if len(got) != len(tt.want) {
				t.Fatalf("Error = %#v, want %#v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("Error[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestNewPreservesRetCodeAndTime(t *testing.T) {
	now := time.Now()
	record := New(CommandNotFound, "", "", "", now)

	if record.RetCode != CommandNotFound {
		t.Errorf("RetCode = %v, want %v", record.RetCode, CommandNotFound)
	}
	if !record.ExecTime.Equal(now) {
		t.Errorf("ExecTime = %v, want %v", record.ExecTime, now)
	}
}

// TestEmptyStreamsAreNilNotEmptySlice is the property that makes len() checks safe
// downstream, and is the reason New exists rather than a bare struct literal.
func TestEmptyStreamsAreNilNotEmptySlice(t *testing.T) {
	record := New(Success, "", "", "", time.Now())

	if record.Stdout != nil {
		t.Errorf("Stdout = %#v, want nil", record.Stdout)
	}
	if record.Stderr != nil {
		t.Errorf("Stderr = %#v, want nil", record.Stderr)
	}
	if record.Error != nil {
		t.Errorf("Error = %#v, want nil", record.Error)
	}
}

func TestRecordString(t *testing.T) {
	now := time.Now()
	record := New(Success, "", "hello\nworld\n", "warn\n", now)

	got := record.String()
	for _, want := range []string{"RetCode: 0", "hello\nworld", "warn"} {
		if !strings.Contains(got, want) {
			t.Errorf("String() = %q, want it to contain %q", got, want)
		}
	}

	// The identity fields are gone; the record must not claim to know them.
	for _, unwanted := range []string{"Namespace:", "Pod:", "Container:"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("String() = %q, should not contain %q", got, unwanted)
		}
	}
}

// TestRecordStringTruncatesLongOutput covers the elision that keeps String usable in a log
// line, and confirms Raw does not elide.
func TestRecordStringTruncatesLongOutput(t *testing.T) {
	long := strings.Repeat("x", 200)
	record := New(Success, "", long, long, time.Now())

	got := record.String()
	if !strings.Contains(got, "...") {
		t.Errorf("String() should elide output longer than the limit, got %q", got)
	}
	if strings.Contains(got, long) {
		t.Error("String() should not contain the full 200-char output")
	}

	raw := record.Raw()
	if !strings.Contains(raw, long) {
		t.Error("Raw() should contain the full output without eliding")
	}
	if strings.Contains(raw, "...") {
		t.Errorf("Raw() should not elide, got %q", raw)
	}
}

// TestRecordStringExactlyAtLimit checks the boundary: output exactly at the limit is not
// truncated, one byte over is.
func TestRecordStringExactlyAtLimit(t *testing.T) {
	atLimit := New(Success, "", strings.Repeat("x", 80), "", time.Now()).String()
	if strings.Contains(atLimit, "...") {
		t.Error("output of exactly 80 bytes should not be elided")
	}

	overLimit := New(Success, "", strings.Repeat("x", 81), "", time.Now()).String()
	if !strings.Contains(overLimit, "...") {
		t.Error("output of 81 bytes should be elided")
	}
}

// TestJSONRoundTrip guards the wire format, which reports depend on.
func TestJSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	original := New(CommandNotFound, "bad", "out\n", "err\n", now)

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// The identity fields must not appear on the wire.
	for _, unwanted := range []string{`"Namespace"`, `"Pod"`, `"Container"`} {
		if strings.Contains(string(encoded), unwanted) {
			t.Errorf("encoded JSON %s should not contain %s", encoded, unwanted)
		}
	}

	var decoded ExecutionRecord
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.RetCode != original.RetCode {
		t.Errorf("RetCode = %v, want %v", decoded.RetCode, original.RetCode)
	}
	if !decoded.ExecTime.Equal(original.ExecTime) {
		t.Errorf("ExecTime = %v, want %v", decoded.ExecTime, original.ExecTime)
	}
	if len(decoded.Stdout) != 1 || decoded.Stdout[0] != "out" {
		t.Errorf("Stdout = %#v, want [\"out\"]", decoded.Stdout)
	}
}
