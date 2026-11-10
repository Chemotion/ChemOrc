package cli

import (
	"regexp"
	"slices"
	"testing"

	vercompare "github.com/hashicorp/go-version"
)

// the table of compose files must stay a clean record: valid versions, newest first, no duplicates,
// and every entry but the first (the version being prepared) pinned to a full commit hash
func TestELNComposeFiles(t *testing.T) {
	fullHash := regexp.MustCompile(`^[0-9a-f]{40}$`)
	var previous *vercompare.Version
	seen := map[string]bool{}
	for i, e := range elnComposeFiles {
		v, err := vercompare.NewVersion(e.version)
		if err != nil {
			t.Errorf("entry %d: invalid version %q", i, e.version)
			continue
		}
		if seen[e.version] {
			t.Errorf("entry %d: duplicate version %s", i, e.version)
		}
		seen[e.version] = true
		if previous != nil && !v.LessThan(previous) {
			t.Errorf("entry %d: %s is not older than %s (the table must be newest first)", i, e.version, previous)
		}
		previous = v
		if i > 0 && !fullHash.MatchString(e.commit) {
			t.Errorf("entry %d (%s): %q is not a full commit hash", i, e.version, e.commit)
		}
	}
}

func TestOfferedVersions(t *testing.T) {
	order := []string{"4.0.1", "3.1.2", "3.0.0", "2.2.0", "2.1.1"} // a fixed list, independent of elnComposeFiles
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
