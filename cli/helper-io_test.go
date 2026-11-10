package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitPortMapping(t *testing.T) {
	cases := []struct {
		mapping, prefix, host, suffix string
		wantErr                       bool
	}{
		{mapping: "4000:4000", host: "4000", suffix: ":4000"},
		{mapping: "4101:4000", host: "4101", suffix: ":4000"},
		{mapping: "80:4000", host: "80", suffix: ":4000"},
		{mapping: "10000:4000", host: "10000", suffix: ":4000"},
		{mapping: "0.0.0.0:4100:4000", prefix: "0.0.0.0:", host: "4100", suffix: ":4000"},
		{mapping: "127.0.0.1:4100:4000/tcp", prefix: "127.0.0.1:", host: "4100", suffix: ":4000/tcp"},
		{mapping: " 4100:4000 ", host: "4100", suffix: ":4000"},
		{mapping: "4000", wantErr: true},           // no host port published
		{mapping: "abc:4000", wantErr: true},       // host port not a number
		{mapping: "4000-4005:4000", wantErr: true}, // ranges are not supported
		{mapping: "70000:4000", wantErr: true},     // out of range
	}
	for _, c := range cases {
		prefix, host, suffix, err := splitPortMapping(c.mapping)
		if c.wantErr {
			if err == nil {
				t.Errorf("splitPortMapping(%q): expected an error, got host %q", c.mapping, host)
			}
			continue
		}
		if err != nil {
			t.Errorf("splitPortMapping(%q): unexpected error: %v", c.mapping, err)
			continue
		}
		if prefix != c.prefix || host != c.host || suffix != c.suffix {
			t.Errorf("splitPortMapping(%q) = (%q, %q, %q), want (%q, %q, %q)", c.mapping, prefix, host, suffix, c.prefix, c.host, c.suffix)
		}
	}
}

func TestChangeExposedPort(t *testing.T) {
	cases := []struct{ before, newPort, after string }{
		{"4000:4000", "4101", "4101:4000"},
		{"4100:4000", "80", "80:4000"},
		{"0.0.0.0:4000:4000", "10000", "0.0.0.0:10000:4000"},
	}
	for _, c := range cases {
		file := filepath.Join(t.TempDir(), "docker-compose.yml")
		compose := "services:\n  eln:\n    image: example/eln\n    ports:\n      - " + c.before + "\n"
		if err := os.WriteFile(file, []byte(compose), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := changeExposedPort(file, c.newPort); err != nil {
			t.Errorf("changeExposedPort(%q -> %s): unexpected error: %v", c.before, c.newPort, err)
			continue
		}
		got, _ := os.ReadFile(file)
		if !strings.Contains(string(got), c.after) || strings.Contains(string(got), "- "+c.before+"\n") {
			t.Errorf("changeExposedPort(%q -> %s): got file\n%s\nwant mapping %q", c.before, c.newPort, got, c.after)
		}
	}
}

func TestChangeExposedPortWithoutPorts(t *testing.T) {
	file := filepath.Join(t.TempDir(), "docker-compose.yml")
	if err := os.WriteFile(file, []byte("services:\n  eln:\n    image: example/eln\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := changeExposedPort(file, "4101"); err == nil {
		t.Errorf("expected an error for a compose file without ports")
	}
}
