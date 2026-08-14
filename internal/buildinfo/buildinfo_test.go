package buildinfo

import "testing"

func TestCurrentUsesInjectedBuildMetadata(t *testing.T) {
	originalVersion := version
	originalCommit := commit
	t.Cleanup(func() {
		version = originalVersion
		commit = originalCommit
	})
	version = "v0.1.0"
	commit = "abc1234"

	got := Current()
	if got.Version != "v0.1.0" {
		t.Fatalf("Version=%q, want %q", got.Version, "v0.1.0")
	}
	if got.Commit != "abc1234" {
		t.Fatalf("Commit=%q, want %q", got.Commit, "abc1234")
	}
}

func TestIsPseudoVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{name: "release", version: "v0.1.0", want: false},
		{name: "release candidate", version: "v0.1.0-rc.1", want: false},
		{name: "pseudo", version: "v0.0.0-20260603232640-51f272dbd384", want: true},
		{name: "dirty pseudo", version: "v0.0.0-20260603232640-51f272dbd384+dirty", want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isPseudoVersion(test.version); got != test.want {
				t.Fatalf("isPseudoVersion(%q)=%t, want %t", test.version, got, test.want)
			}
		})
	}
}
