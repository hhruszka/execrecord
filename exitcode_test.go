package execrecord

import (
	"fmt"
	"testing"
)

// TestConstantValues pins the numeric value of every constant. These values are a
// wire contract: they are serialised into JSON reports and compared against shell
// exit statuses, so an accidental change must fail loudly.
func TestConstantValues(t *testing.T) {
	tests := []struct {
		name string
		code ExitCode
		want int
	}{
		{"Success", Success, 0},
		{"GeneralError", GeneralError, 1},
		{"IncorrectUsageExitCode", IncorrectUsageExitCode, 2},
		{"CommandCannotExecute", CommandCannotExecute, 126},
		{"CommandNotFound", CommandNotFound, 127},
		{"InvalidArgumentToExit", InvalidArgumentToExit, 128},
		{"FatalErrorSignal1", FatalErrorSignal1, 129},
		{"ScriptTerminatedByControlC", ScriptTerminatedByControlC, 130},
		{"FatalErrorSignal15", FatalErrorSignal15, 143},
		{"ExitStatusOutOfRange", ExitStatusOutOfRange, 255},
	}

	for _, tt := range tests {
		if int(tt.code) != tt.want {
			t.Errorf("%s = %d, want %d", tt.name, int(tt.code), tt.want)
		}
	}
}

// TestNoConstantIsNegative guards the invariant this type exists to hold: an ExitCode
// describes a process that ran, so every constant must be a real POSIX status. A
// negative constant would mean some outcome-that-never-ran had been smuggled back in,
// which is exactly the conflation the negative sentinels used to cause.
func TestNoConstantIsNegative(t *testing.T) {
	for code, description := range exitCodeDescriptions {
		if code < 0 {
			t.Errorf("ExitCode %d (%q) is negative; outcomes without an exit status belong in an error or a verdict, not here", int(code), description)
		}
		if code > 255 {
			t.Errorf("ExitCode %d (%q) exceeds the POSIX range 0-255", int(code), description)
		}
	}
}

// TestSignalCodesFollow128Plus checks the 128+n convention holds for every signal
// constant, which is what lets callers map a signal number to a code arithmetically.
func TestSignalCodesFollow128Plus(t *testing.T) {
	signals := map[int]ExitCode{
		1:  FatalErrorSignal1,
		3:  FatalErrorSignal3,
		4:  FatalErrorSignal4,
		5:  FatalErrorSignal5,
		6:  FatalErrorSignal6,
		7:  FatalErrorSignal7,
		8:  FatalErrorSignal8,
		9:  FatalErrorSignal9,
		10: FatalErrorSignal10,
		11: FatalErrorSignal11,
		12: FatalErrorSignal12,
		13: FatalErrorSignal13,
		14: FatalErrorSignal14,
		15: FatalErrorSignal15,
	}

	for n, code := range signals {
		if want := 128 + n; int(code) != want {
			t.Errorf("FatalErrorSignal%d = %d, want %d", n, int(code), want)
		}
	}
}

func TestExitCodeString(t *testing.T) {
	tests := []struct {
		code ExitCode
		want string
	}{
		{Success, "Success"},
		{GeneralError, "General error, unspecified error"},
		{CommandNotFound, "Command not found"},
		{FatalErrorSignal9, "Fatal error signal 9 (SIGKILL)"},
		{ExitStatusOutOfRange, "Exit status out of range"},
		{ExitCode(77), "77"},   // unmapped positive code renders as its number
		{ExitCode(-99), "-99"}, // no constant is negative now, but String must not mangle one
	}

	for _, tt := range tests {
		if got := tt.code.String(); got != tt.want {
			t.Errorf("ExitCode(%d).String() = %q, want %q", int(tt.code), got, tt.want)
		}
	}
}

// TestStringUsesDecimalNotIota guards against ExitCode's String method being used
// implicitly by fmt in a way that recurses or renders the wrong base.
func TestExitCodeStringUsesDecimalNotIota(t *testing.T) {
	if got := fmt.Sprintf("%v", ExitCode(42)); got != "42" {
		t.Errorf("fmt %%v of unmapped code = %q, want %q", got, "42")
	}
	if got := fmt.Sprintf("%v", Success); got != "Success" {
		t.Errorf("fmt %%v of Success = %q, want %q", got, "Success")
	}
}

// TestDescriptionMatchesString keeps the two entry points from diverging.
func TestDescriptionMatchesString(t *testing.T) {
	codes := []ExitCode{
		Success, GeneralError, IncorrectUsageExitCode, CommandCannotExecute,
		CommandNotFound, InvalidArgumentToExit, ScriptTerminatedByControlC,
		ExitStatusOutOfRange, ExitCode(77),
	}

	for _, code := range codes {
		if got, want := Description(code), code.String(); got != want {
			t.Errorf("Description(%d) = %q, String() = %q", int(code), got, want)
		}
	}
}

// TestEveryConstantHasDescription ensures a newly added constant cannot silently
// render as a bare number.
func TestEveryConstantHasDescription(t *testing.T) {
	named := map[string]ExitCode{
		"Success":                    Success,
		"GeneralError":               GeneralError,
		"IncorrectUsageExitCode":     IncorrectUsageExitCode,
		"CommandCannotExecute":       CommandCannotExecute,
		"CommandNotFound":            CommandNotFound,
		"InvalidArgumentToExit":      InvalidArgumentToExit,
		"ScriptTerminatedByControlC": ScriptTerminatedByControlC,
		"ExitStatusOutOfRange":       ExitStatusOutOfRange,
		"FatalErrorSignal1":          FatalErrorSignal1,
		"FatalErrorSignal3":          FatalErrorSignal3,
		"FatalErrorSignal4":          FatalErrorSignal4,
		"FatalErrorSignal5":          FatalErrorSignal5,
		"FatalErrorSignal6":          FatalErrorSignal6,
		"FatalErrorSignal7":          FatalErrorSignal7,
		"FatalErrorSignal8":          FatalErrorSignal8,
		"FatalErrorSignal9":          FatalErrorSignal9,
		"FatalErrorSignal10":         FatalErrorSignal10,
		"FatalErrorSignal11":         FatalErrorSignal11,
		"FatalErrorSignal12":         FatalErrorSignal12,
		"FatalErrorSignal13":         FatalErrorSignal13,
		"FatalErrorSignal14":         FatalErrorSignal14,
		"FatalErrorSignal15":         FatalErrorSignal15,
	}

	for name, code := range named {
		if _, ok := exitCodeDescriptions[code]; !ok {
			t.Errorf("%s (%d) has no entry in exitCodeDescriptions", name, int(code))
		}
	}
}

// TestHasDescription covers the predicate callers use to tell a known code from an
// unrecognised one, which String cannot express because unknown codes render as a
// number rather than failing.
func TestHasDescription(t *testing.T) {
	tests := []struct {
		code ExitCode
		want bool
	}{
		{Success, true},
		{CommandNotFound, true},
		{FatalErrorSignal15, true},
		{ExitCode(77), false},  // valid POSIX status, no description registered
		{ExitCode(-99), false}, // not a POSIX status at all
		{ExitCode(125), false},
	}

	for _, tt := range tests {
		if got := HasDescription(tt.code); got != tt.want {
			t.Errorf("HasDescription(%d) = %v, want %v", int(tt.code), got, tt.want)
		}
	}
}

// TestNoDuplicateDescriptions catches two constants collapsing onto one value,
// which is how FatalErrorSignal2 would silently shadow ScriptTerminatedByControlC.
func TestNoDuplicateDescriptions(t *testing.T) {
	seen := make(map[string]ExitCode, len(exitCodeDescriptions))
	for code, description := range exitCodeDescriptions {
		if prior, ok := seen[description]; ok {
			t.Errorf("description %q used by both %d and %d", description, int(prior), int(code))
		}
		seen[description] = code
	}
}
