package execrecord

import "fmt"

// ExitCode is the status a process exited with. Values are strictly POSIX (0-255):
// an ExitCode only ever describes a command that actually ran to completion.
//
// Outcomes that produce no exit status are deliberately not representable here.
// An executor that never obtained a status reports it as an error, and a caller that
// decided an outcome without running anything records that decision as a verdict of
// its own. Encoding either as a negative "exit code" conflates a process result with
// a judgement about it, which is what this type used to do.
type ExitCode int

const (
	// POSIX exit statuses.
	Success                ExitCode = 0
	GeneralError           ExitCode = 1
	IncorrectUsageExitCode ExitCode = 2

	CommandCannotExecute       ExitCode = 126
	CommandNotFound            ExitCode = 127
	InvalidArgumentToExit      ExitCode = 128
	ScriptTerminatedByControlC ExitCode = 130
	ExitStatusOutOfRange       ExitCode = 255
	// Signal based exit codes (128+n)
	FatalErrorSignal1 ExitCode = 129
	// FatalErrorSignal2 is omitted as it overlaps with ScriptTerminatedByControlC
	FatalErrorSignal3  ExitCode = 131
	FatalErrorSignal4  ExitCode = 132
	FatalErrorSignal5  ExitCode = 133
	FatalErrorSignal6  ExitCode = 134
	FatalErrorSignal7  ExitCode = 135
	FatalErrorSignal8  ExitCode = 136
	FatalErrorSignal9  ExitCode = 137
	FatalErrorSignal10 ExitCode = 138
	FatalErrorSignal11 ExitCode = 139
	FatalErrorSignal12 ExitCode = 140
	FatalErrorSignal13 ExitCode = 141
	FatalErrorSignal14 ExitCode = 142
	FatalErrorSignal15 ExitCode = 143
)

// String returns the description for the exit code, or its decimal value when
// the code has no registered description.
func (e ExitCode) String() string {
	if description, ok := exitCodeDescriptions[e]; ok {
		return description
	}
	return fmt.Sprintf("%d", int(e))
}

// exitCodeDescriptions maps possible exit codes with descriptive names.
var exitCodeDescriptions = map[ExitCode]string{
	Success:                "Success",
	GeneralError:           "General error, unspecified error",
	IncorrectUsageExitCode: "Incorrect usage or syntax of the command",
	CommandCannotExecute:   "Command cannot execute",
	CommandNotFound:        "Command not found",
	InvalidArgumentToExit:  "Invalid argument to exit",

	ScriptTerminatedByControlC: "Script terminated by Control-C (SIGINT)",
	ExitStatusOutOfRange:       "Exit status out of range",

	// Signal based exit codes (128+n)
	FatalErrorSignal1: "Fatal error signal 1 (SIGHUP)",
	// FatalErrorSignal2 overlaps with ScriptTerminatedByControlC
	FatalErrorSignal3:  "Fatal error signal 3 (SIGQUIT)",
	FatalErrorSignal4:  "Fatal error signal 4 (SIGILL)",
	FatalErrorSignal5:  "Fatal error signal 5 (SIGTRAP)",
	FatalErrorSignal6:  "Fatal error signal 6 (SIGABRT/SIGIOT)",
	FatalErrorSignal7:  "Fatal error signal 7 (SIGBUS)",
	FatalErrorSignal8:  "Fatal error signal 8 (SIGFPE)",
	FatalErrorSignal9:  "Fatal error signal 9 (SIGKILL)",
	FatalErrorSignal10: "Fatal error signal 10 (SIGUSR1)",
	FatalErrorSignal11: "Fatal error signal 11 (SIGSEGV)",
	FatalErrorSignal12: "Fatal error signal 12 (SIGUSR2)",
	FatalErrorSignal13: "Fatal error signal 13 (SIGPIPE)",
	FatalErrorSignal14: "Fatal error signal 14 (SIGALRM)",
	FatalErrorSignal15: "Fatal error signal 15 (SIGTERM)",
}

// Description returns a string description for a given exit code.
// It looks up the code in the predefined exitCodeDescriptions map. If the code is
// found, it returns the corresponding description. If not, it returns the code's
// decimal value.
func Description(code ExitCode) string {
	return code.String()
}

// HasDescription reports whether the code has a registered description. Callers
// that need to distinguish a known code from an unrecognised one cannot do so from
// String alone, because an unknown code renders as its decimal value.
func HasDescription(code ExitCode) bool {
	_, ok := exitCodeDescriptions[code]
	return ok
}
