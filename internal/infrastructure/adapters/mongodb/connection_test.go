package mongodb

import (
	"strings"
	"testing"
)

func TestBuildURI(t *testing.T) {
	uri, err := buildURI("db.example", "user@example", "p@ss", "z=last&a=admin%20users", "go-template", 27018)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(uri, "user%40example:p%40ss@db.example:27018/go-template?a=admin+users&z=last") {
		t.Fatalf("uri = %q", uri)
	}
}

func TestBuildURIRejectsUnsafeHostAndCredentials(t *testing.T) {
	for _, test := range []struct {
		name, host, user, pass string
	}{
		{"scheme", "mongodb://db", "", ""},
		{"port", "db:27017", "", ""},
		{"path", "db/path", "", ""},
		{"paired credentials", "db", "user", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := buildURI(test.host, test.user, test.pass, "", "db", 27017); err == nil {
				t.Fatal("expected URI validation error")
			}
		})
	}
}

func TestBuildURIDefaultsHost(t *testing.T) {
	uri, err := buildURI("", "", "", "", "go-template", 27017)
	if err != nil || uri != "mongodb://localhost:27017/go-template" {
		t.Fatalf("uri = %q, err = %v", uri, err)
	}
}

func TestBuildURICredentiallessHasNoQuery(t *testing.T) {
	uri, err := buildURI("db.example", "", "", "", "go-template", 27017)
	if err != nil || uri != "mongodb://db.example:27017/go-template" {
		t.Fatalf("uri = %q, err = %v", uri, err)
	}
}

func TestBuildURIRejectsUnsafeArgs(t *testing.T) {
	for _, args := range []string{"bad=%zz", "a=1;b=2"} {
		if _, err := buildURI("db.example", "", "", args, "go-template", 27017); err == nil {
			t.Fatalf("expected args validation error for %q", args)
		}
	}
}

func TestBuildURIAppliesArgsWithoutCredentials(t *testing.T) {
	uri, err := buildURI("db.example", "", "", "z=last&a=first", "go-template", 27017)
	if err != nil || uri != "mongodb://db.example:27017/go-template?a=first&z=last" {
		t.Fatalf("uri = %q, err = %v", uri, err)
	}
}
