package version

import (
	"strings"
	"testing"
	"time"
)

func restoreVars() func() {
	serviceName := ServiceName
	commitTag := CommitTag
	commitSHA := CommitSHA
	commitBranch := CommitBranch
	originURL := OriginURL
	buildDate := BuildDate

	return func() {
		ServiceName = serviceName
		CommitTag = commitTag
		CommitSHA = commitSHA
		CommitBranch = commitBranch
		OriginURL = originURL
		BuildDate = buildDate
	}
}

func TestNewVersionFallsBackToDefaultTag(t *testing.T) {
	t.Cleanup(restoreVars())

	ServiceName = "svc"
	CommitTag = unspecified
	CommitSHA = "deadbeef"
	CommitBranch = "main"
	OriginURL = "git@example.com/repo.git"
	BuildDate = "2024-01-02T03:04:05ZUTC"

	ver, err := NewVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ver.Tag != "v0.0.0" {
		t.Fatalf("expected fallback tag v0.0.0, got %s", ver.Tag)
	}

	if ver.URL != "git@example.com/repo" {
		t.Fatalf("expected trimmed origin URL, got %s", ver.URL)
	}

	if !ver.IfSpecified() {
		t.Fatalf("expected version to be specified when fields set")
	}

	if !strings.Contains(ver.String(), "Branch main, commit hash: deadbeef") {
		t.Fatalf("version string missing branch/commit info: %s", ver.String())
	}

	expectedTime := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	if !ver.Date.Equal(expectedTime) {
		t.Fatalf("unexpected build date: %v", ver.Date)
	}
}

func TestNewVersionHandlesUnspecifiedValues(t *testing.T) {
	t.Cleanup(restoreVars())

	ServiceName = unspecified
	CommitTag = unspecified
	CommitSHA = unspecified
	CommitBranch = unspecified
	OriginURL = unspecified
	BuildDate = unspecified

	ver, err := NewVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ver.Tag != "v0.0.0" {
		t.Fatalf("expected fallback tag v0.0.0, got %s", ver.Tag)
	}

	if ver.IfSpecified() {
		t.Fatalf("expected unspecified values to be treated as not specified")
	}

	if ver.String() != "Unspecified v0.0.0\n\n" {
		t.Fatalf("unexpected version string: %q", ver.String())
	}
}

func TestNewVersionBadDate(t *testing.T) {
	t.Cleanup(restoreVars())

	BuildDate = "bad-date"

	if _, err := NewVersion(); err == nil {
		t.Fatalf("expected error for bad date, got nil")
	}
}

func TestIfSpecifiedMatrix(t *testing.T) {
	t.Cleanup(restoreVars())

	ServiceName = "svc"
	CommitTag = "v1.0.0"
	CommitSHA = unspecified
	CommitBranch = "main"
	OriginURL = "git@example.com/repo.git"
	BuildDate = unspecified

	ver, err := NewVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ver.IfSpecified() {
		t.Fatalf("expected IfSpecified to be false when commit is unspecified")
	}

	CommitSHA = "cafebabe"

	ver, err = NewVersion()
	if err != nil {
		t.Fatalf("unexpected error after setting commit: %v", err)
	}

	if !ver.IfSpecified() {
		t.Fatalf("expected IfSpecified to be true when all fields set")
	}
}
