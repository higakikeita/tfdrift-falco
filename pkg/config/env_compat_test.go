package config

import (
	"os"
	"testing"
)

// The project was renamed from tfdrift-falco to driftwire. Anything an existing
// operator has exported is still spelled TFDRIFT_*, so those names must keep
// working — and an explicitly set DRIFTWIRE_* value must win over the old one.
func TestAdoptLegacyEnv(t *testing.T) {
	t.Run("legacy name is adopted", func(t *testing.T) {
		t.Setenv("TFDRIFT_LOG_LEVEL", "debug")
		os.Unsetenv("DRIFTWIRE_LOG_LEVEL")

		adoptLegacyEnv()

		if got := os.Getenv("DRIFTWIRE_LOG_LEVEL"); got != "debug" {
			t.Fatalf("DRIFTWIRE_LOG_LEVEL = %q, want %q (legacy TFDRIFT_LOG_LEVEL was not adopted)", got, "debug")
		}
	})

	t.Run("explicit new name wins", func(t *testing.T) {
		t.Setenv("TFDRIFT_LOG_LEVEL", "debug")
		t.Setenv("DRIFTWIRE_LOG_LEVEL", "warn")

		adoptLegacyEnv()

		if got := os.Getenv("DRIFTWIRE_LOG_LEVEL"); got != "warn" {
			t.Fatalf("DRIFTWIRE_LOG_LEVEL = %q, want %q (legacy value must not overwrite an explicit one)", got, "warn")
		}
	})

	t.Run("unrelated variables are untouched", func(t *testing.T) {
		t.Setenv("PATH_NOT_OURS", "x")
		adoptLegacyEnv()
		if _, ok := os.LookupEnv("DRIFTWIRE_NOT_OURS"); ok {
			t.Fatal("adoptLegacyEnv invented DRIFTWIRE_NOT_OURS from an unrelated variable")
		}
	})
}
