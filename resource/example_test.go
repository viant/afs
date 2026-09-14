package resource_test

import (
	"context"
	"fmt"

	"github.com/viant/afs"
	"github.com/viant/afs/resource"
)

func ExampleNewProvider() {
	provider, err := resource.NewProvider(resource.Config{
		Scheme: "skill",
		Roots: map[string]resource.Root{
			"datly-reader": {FS: embedded, Path: "testdata/skills/datly-reader"},
		},
	})
	if err != nil {
		panic(err)
	}
	files := afs.NewWithProviders(map[string]afs.Provider{"skill": provider})
	data, err := files.DownloadWithURL(context.Background(), "skill://datly-reader/references/guide.md")
	if err != nil {
		panic(err)
	}
	fmt.Print(string(data))
	// Output: Supporting guide.
}
