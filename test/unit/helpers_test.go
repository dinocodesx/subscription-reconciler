package unit

import "time"

// mustParseTimeHelper is used in tests that don't want to thread *testing.T through helper code.
func mustParseTimeHelper(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic("mustParseTimeHelper: " + err.Error())
	}
	return parsed.UTC()
}
