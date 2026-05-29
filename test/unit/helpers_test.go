package unit

import (
	"fmt"
	"time"
)

// mustParseTimeHelper is used in tests that don't want to thread *testing.T through helper code.
func mustParseTimeHelper(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(fmt.Sprintf("mustParseTimeHelper(%q): %v", value, err))
	}
	return parsed.UTC()
}
