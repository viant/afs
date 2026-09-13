package embed

import (
	iofs "io/fs"
	"reflect"
	"testing"
)

func TestHolderMatchesEmbeddedFileOrdering(t *testing.T) {
	for _, paths := range [][]string{{"a.txt", "z/query.sql"}, {"a/b/query.sql", "a/c/query.sql", "root.sql"}, {"z/query.sql", "a.txt"}} {
		t.Run(paths[0], func(t *testing.T) {
			holder := NewHolder()
			for _, name := range paths {
				holder.Add(name, name)
			}
			snapshot := holder.EmbedFs()
			for _, name := range paths {
				content, err := iofs.ReadFile(snapshot, name)
				if err != nil || string(content) != name {
					t.Fatalf("read %s: %q %v", name, content, err)
				}
			}
		})
	}
}

func TestHolderCreatesAllParentDirectories(t *testing.T) {
	holder := NewHolder()
	holder.Add("a/b/c/query.sql", "SELECT 1")
	snapshot := holder.EmbedFs()
	for _, tt := range []struct {
		dir   string
		names []string
	}{{".", []string{"a"}}, {"a", []string{"b"}}, {"a/b", []string{"c"}}, {"a/b/c", []string{"query.sql"}}} {
		entries, err := iofs.ReadDir(snapshot, tt.dir)
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		if !reflect.DeepEqual(names, tt.names) {
			t.Fatalf("directory %s: %v want %v", tt.dir, names, tt.names)
		}
	}
}
