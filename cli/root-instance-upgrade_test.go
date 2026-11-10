package cli

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// labels added by an admin (e.g. for a reverse proxy) must survive an upgrade; the project label is set once
func TestWithProjectLabel(t *testing.T) {
	file := filepath.Join(t.TempDir(), "docker-compose.cli.yml")
	yaml := `services:
  eln:
    labels:
      - traefik.enable=true
      - traefik.http.routers.eln.rule=Host(` + "`eln.example.org`" + `)
      - net.chemotion.cli.project=old-name
  worker:
    environment:
      - X=1
  converter:
    labels:
      traefik.enable: "true"
`
	if err := os.WriteFile(file, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	compose, err := readYAML(file)
	if err != nil {
		t.Fatal(err)
	}

	got, ok := withProjectLabel(compose.Get("services.eln.labels"), "inst-1234")
	want := []string{"traefik.enable=true", "traefik.http.routers.eln.rule=Host(`eln.example.org`)", "net.chemotion.cli.project=inst-1234"}
	if !ok || !slices.Equal(got, want) {
		t.Errorf("existing list: got %q (ok=%v), want %q", got, ok, want)
	}

	got, ok = withProjectLabel(compose.Get("services.worker.labels"), "inst-1234")
	if !ok || !slices.Equal(got, []string{"net.chemotion.cli.project=inst-1234"}) {
		t.Errorf("no labels yet: got %q (ok=%v)", got, ok)
	}

	if _, ok = withProjectLabel(compose.Get("services.converter.labels"), "inst-1234"); ok {
		t.Errorf("map-form labels must be reported, not overwritten")
	}

	// a list the CLI itself just set (as []string) is handled the same way
	got, _ = withProjectLabel([]string{"a=b", "net.chemotion.cli.project=x"}, "y")
	if !slices.Equal(got, []string{"a=b", "net.chemotion.cli.project=y"}) {
		t.Errorf("[]string input: got %q", got)
	}
}
