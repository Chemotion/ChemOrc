package cli

import (
	"slices"
	"testing"
)

func TestOfferedVersions(t *testing.T) {
	order := []string{"4.0.1", "3.1.2", "3.0.0", "2.2.0", "2.1.1"}
	cases := []struct {
		name, current, minimum string
		want                   []string
	}{
		// new installations: nothing older than oldestInstallableELN
		{"install", "1.3.1", oldestInstallableELN, []string{"4.0.1", "3.1.2", "3.0.0"}},
		// upgrades keep every newer version, including the jump from 2.x
		{"upgrade from 2.1.1", "2.1.1", "", []string{"4.0.1", "3.1.2", "3.0.0", "2.2.0", "2.1.1 - current version"}},
		{"upgrade from 2.2.0", "2.2.0", "", []string{"4.0.1", "3.1.2", "3.0.0", "2.2.0 - current version"}},
		{"upgrade from 3.1.2", "3.1.2", "", []string{"4.0.1", "3.1.2 - current version"}},
		{"already newest", "4.0.1", "", []string{"4.0.1 - current version"}},
	}
	for _, c := range cases {
		if got := offeredVersions(c.current, c.minimum, order); !slices.Equal(got, c.want) {
			t.Errorf("%s: offeredVersions(%q, %q) = %q, want %q", c.name, c.current, c.minimum, got, c.want)
		}
	}
}
