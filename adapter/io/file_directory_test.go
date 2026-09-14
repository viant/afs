package io_test

import (
	"context"
	"github.com/viant/afs"
	adapter "github.com/viant/afs/adapter/io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestFileDirectoryIdentity(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "docs"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ name, content string }{{"plain.yaml", "Columns: {}"}, {"empty", ""}, {"docs/nested.yaml", "Paths: {}"}} {
		if err := os.WriteFile(filepath.Join(root, test.name), []byte(test.content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	source := adapter.NewFS(context.Background(), afs.New(), "file://"+root)
	for _, test := range []struct{ name, content string }{{"plain.yaml", "Columns: {}"}, {"empty", ""}, {"docs/nested.yaml", "Paths: {}"}} {
		t.Run(test.name, func(t *testing.T) {
			data, err := fs.ReadFile(source, test.name)
			if err != nil || string(data) != test.content {
				t.Fatalf("read=%q err=%v", data, err)
			}
			info, err := fs.Stat(source, test.name)
			if err != nil || info.IsDir() {
				t.Fatalf("stat=%v err=%v", info, err)
			}
		})
	}
	for _, name := range []string{".", "docs"} {
		info, err := fs.Stat(source, name)
		if err != nil || !info.IsDir() {
			t.Fatalf("dir=%v err=%v", info, err)
		}
	}
	for _, name := range []string{"missing", "docs/missing", "../plain.yaml", "/plain.yaml", "docs/../plain.yaml"} {
		if _, err := source.Open(name); err == nil {
			t.Fatalf("invalid/missing %q accepted", name)
		}
	}
	if err := fstest.TestFS(source, "plain.yaml", "empty", "docs/nested.yaml"); err != nil {
		t.Fatal(err)
	}
}
