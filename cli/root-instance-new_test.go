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

// an unreadable current version (e.g. tag `eln-latest`, or a registry host with a hyphen) must not panic
func TestOfferedVersionsUnknownCurrent(t *testing.T) {
	order := []string{"4.0.1", "3.1.2", "3.0.0"}
	for _, current := range []string{"latest", "", "registry.example/chemotion:eln-4.0.1"} {
		if got := offeredVersions(current, "", order); !slices.Equal(got, order) {
			t.Errorf("offeredVersions(%q) = %q, want every version %q", current, got, order)
		}
	}
}

func TestWithLatestELN(t *testing.T) {
	newest := "4.0.1"
	known := func() map[string]string {
		return map[string]string{"4.0.1": "pinned-4.0.1", "3.1.2": "pinned-3.1.2", "3.0.0": "pinned-3.0.0"}
	}
	cases := []struct {
		name, current, latest string
		offered, want         []string
		wantURL               map[string]string // version -> expected URL afterwards
	}{
		{"known latest is marked, URL kept", "3.1.2", "4.0.1",
			[]string{"4.0.1", "3.1.2 - current version"}, []string{"4.0.1 - latest stable version", "3.1.2 - current version"},
			map[string]string{"4.0.1": "pinned-4.0.1"}},
		{"stale latest (older than this CLI) does not hijack its pinned entry", "3.1.2", "3.1.2",
			[]string{"4.0.1", "3.1.2 - current version"}, []string{"4.0.1", "3.1.2 - current version"},
			map[string]string{"3.1.2": "pinned-3.1.2"}},
		{"stale known latest is not marked as latest for an older instance", "3.0.0", "3.1.2",
			[]string{"4.0.1", "3.1.2", "3.0.0 - current version"}, []string{"4.0.1", "3.1.2", "3.0.0 - current version"},
			map[string]string{"3.1.2": "pinned-3.1.2"}},
		{"stale latest is not offered to a newer instance", "4.0.1", "3.1.2",
			[]string{"4.0.1 - current version"}, []string{"4.0.1 - current version"},
			map[string]string{"3.1.2": "pinned-3.1.2"}},
		{"release newer than this CLI is offered first, from the latest release", "3.1.2", "4.1.0",
			[]string{"4.0.1", "3.1.2 - current version"}, []string{"4.1.0 - latest stable version", "4.0.1", "3.1.2 - current version"},
			map[string]string{"4.1.0": composeURL, "4.0.1": "pinned-4.0.1"}},
		{"newer release not offered when the instance already runs it", "4.1.0", "4.1.0",
			nil, nil, map[string]string{}},
		{"unreadable latest is ignored", "3.1.2", "garbage",
			[]string{"4.0.1", "3.1.2 - current version"}, []string{"4.0.1", "3.1.2 - current version"}, map[string]string{}},
	}
	for _, c := range cases {
		versions := known()
		got := withLatestELN(slices.Clone(c.offered), versions, newest, c.current, c.latest)
		if !slices.Equal(got, c.want) {
			t.Errorf("%s: offered %q, want %q", c.name, got, c.want)
		}
		for v, url := range c.wantURL {
			if versions[v] != url {
				t.Errorf("%s: version %s resolves to %q, want %q", c.name, v, versions[v], url)
			}
		}
	}
}
