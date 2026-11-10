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

// write a compose file whose eln service has the given ports block (YAML lines under `ports:`)
func writeCompose(t *testing.T, ports string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "docker-compose.yml")
	compose := "services:\n  eln:\n    image: example/eln\n"
	if ports != "" {
		compose += "    ports:\n" + ports
	}
	if err := os.WriteFile(file, []byte(compose), 0o644); err != nil {
		t.Fatal(err)
	}
	return file
}

// read back the first port mapping of the eln service, exactly as written
func firstMapping(t *testing.T, file string) string {
	t.Helper()
	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	got, err := evaluateYamlExpression(string(content), ".services.eln.ports[0]")
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(got)
}

func TestChangeExposedPort(t *testing.T) {
	cases := []struct{ before, newPort, after string }{
		{"4000:4000", "4101", "4101:4000"},
		{"4100:4000", "80", "80:4000"},
		{"4000:4000", "4180", "4180:4000"}, // must not be taken for 80:4000 or 4100:4000
		{"0.0.0.0:4000:4000", "10000", "0.0.0.0:10000:4000"},
		{"4000:4000/tcp", "4101", "4101:4000/tcp"},
	}
	for _, c := range cases {
		file := writeCompose(t, "      - "+c.before+"\n")
		if err := changeExposedPort(file, c.newPort); err != nil {
			t.Errorf("changeExposedPort(%q -> %s): unexpected error: %v", c.before, c.newPort, err)
			continue
		}
		if got := firstMapping(t, file); got != c.after {
			t.Errorf("changeExposedPort(%q -> %s) = %q, want %q", c.before, c.newPort, got, c.after)
		}
	}
}

// an upgrade must keep the old bind address and host port, whatever the new file publishes
func TestCarryOverExposure(t *testing.T) {
	cases := []struct{ old, newFile, after string }{
		{"127.0.0.1:4100:4000", "4000:4000", "127.0.0.1:4100:4000"},
		{"4100:4000", "0.0.0.0:4000:4000", "4100:4000"},
		{"80:4000", "4000:4000", "80:4000"},
		{"10000:4000", "4000:4000", "10000:4000"},
	}
	for _, c := range cases {
		old, err := readYAML(writeCompose(t, "      - "+c.old+"\n"))
		if err != nil {
			t.Fatal(err)
		}
		bindPrefix, hostPort, set, err := publishedExposure(&old)
		if !set || err != nil {
			t.Errorf("publishedExposure(%q): set=%v err=%v", c.old, set, err)
			continue
		}
		file := writeCompose(t, "      - "+c.newFile+"\n")
		if err := carryOverExposure(file, bindPrefix, hostPort); err != nil {
			t.Errorf("carryOverExposure(%q into %q): unexpected error: %v", c.old, c.newFile, err)
			continue
		}
		if got := firstMapping(t, file); got != c.after {
			t.Errorf("carryOverExposure(%q into %q) = %q, want %q", c.old, c.newFile, got, c.after)
		}
	}
}

// ports that cannot be carried over must be an error, never "not set" (which would keep the new file's default port)
func TestPublishedExposure(t *testing.T) {
	cases := []struct {
		name, ports    string
		wantSet, isErr bool
	}{
		{"no ports entry", "", false, false},
		{"short form", "      - 4101:4000\n", true, false},
		{"long syntax", "      - target: 4000\n        published: 4101\n", true, true},
		{"container port only", "      - \"4000\"\n", true, true},
		{"no fixed host port", "      - 127.0.0.1::4000\n", true, true},
	}
	for _, c := range cases {
		compose, err := readYAML(writeCompose(t, c.ports))
		if err != nil {
			t.Fatal(err)
		}
		_, _, set, err := publishedExposure(&compose)
		if set != c.wantSet || (err != nil) != c.isErr {
			t.Errorf("%s: publishedExposure() set=%v err=%v, want set=%v error=%v", c.name, set, err, c.wantSet, c.isErr)
		}
	}
}

func TestChangeExposedPortWithoutPorts(t *testing.T) {
	file := writeCompose(t, "")
	if err := changeExposedPort(file, "4101"); err == nil {
		t.Errorf("expected an error for a compose file without ports")
	}
}

// Rewriting the port must leave the rest of the file intact: trailing and full-line comments included.
// (The goccy decoder used before turned `cpus: 1 # …` into `!!int '1 # …'`, which Docker Compose rejects.)
func TestRewriteKeepsComments(t *testing.T) {
	file := writeCompose(t, "      - 4000:4000\n")
	extra := "    # a full-line comment\n    cpus: 1 # a trailing comment\n    mem_limit: 1g\n"
	content, _ := os.ReadFile(file)
	if err := os.WriteFile(file, append(content, []byte(extra)...), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := carryOverExposure(file, "", "4101"); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(file)
	compose, err := readYAML(file)
	if err != nil {
		t.Fatalf("rewritten file does not parse: %v\n%s", err, out)
	}
	if got := compose.Get("services.eln.cpus"); got != 1 {
		t.Errorf("cpus = %#v after the rewrite, want 1\n%s", got, out)
	}
	for _, want := range []string{"# a full-line comment", "# a trailing comment", "4101:4000"} {
		if !strings.Contains(string(out), want) {
			t.Errorf("rewritten file lost %q:\n%s", want, out)
		}
	}
}

// The payload shipped to users must survive the rewrite every install and upgrade applies to it.
func TestPayloadSurvivesPortRewrite(t *testing.T) {
	payload, err := os.ReadFile(filepath.Join("..", "payload", "docker-compose.yml"))
	if err != nil {
		t.Skip("payload not found:", err)
	}
	file := filepath.Join(t.TempDir(), "docker-compose.yml")
	if err := os.WriteFile(file, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := carryOverExposure(file, "127.0.0.1:", "4101"); err != nil {
		t.Fatal(err)
	}
	compose, err := readYAML(file)
	if err != nil {
		t.Fatalf("rewritten payload does not parse: %v", err)
	}
	if got := firstMapping(t, file); got != "127.0.0.1:4101:4000" {
		t.Errorf("port mapping = %q, want 127.0.0.1:4101:4000", got)
	}
	for _, service := range getSubHeadings(&compose, "services") {
		key := "services." + service + ".image"
		if compose.GetString(key) == "" {
			t.Errorf("%s lost its image after the rewrite", service)
		}
	}
	if v := compose.Get("services.pdftext.cpus"); v != nil && v != 1 {
		t.Errorf("pdftext cpus = %#v after the rewrite, want 1", v)
	}
}
