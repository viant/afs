package file

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/matcher"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

func TestManager_List(t *testing.T) {

	baseDir := os.TempDir()
	fileManager := New()

	ignoreMatcher, _ := matcher.NewIgnore([]string{"*.txt", ".ignore"})

	var useCases = []struct {
		description  string
		location     string
		baseLocation string
		assets       map[string]string
		matcher      option.Match
	}{
		{
			description:  "location download",
			location:     path.Join(baseDir, "file_list_001/"),
			baseLocation: path.Join(baseDir, "file_list_001"),
			assets: map[string]string{
				"file1.txt": "abc",
				"file2.txt": "abc",
			},
		},
		{
			description:  "location download with filter",
			location:     path.Join(baseDir, "file_list_002/"),
			baseLocation: path.Join(baseDir, "file_list_002/"),
			assets: map[string]string{
				"file1.txt": "abc1",
				"file2.txt": "abc2",
				"asset.csv": "abc3",
				"asset.tsv": "abc4",
				".ignore":   "abc4",
			},
			matcher: ignoreMatcher.Match,
		},
	}

	ctx := context.Background()
	for _, useCase := range useCases {
		manager := New()
		assert.EqualValues(t, Scheme, manager.Scheme(), useCase.description)
		_ = fileManager.Delete(ctx, useCase.baseLocation)
		_ = fileManager.Create(ctx, useCase.baseLocation, 0744, true)

		for name, content := range useCase.assets {
			filename := path.Join(useCase.baseLocation, name)
			_ = ioutil.WriteFile(filename, []byte(content), 0744)
		}

		actuals := map[string]string{}
		var options = make([]storage.Option, 0)
		if useCase.matcher != nil {
			options = append(options, useCase.matcher)
		}
		objects, err := manager.List(ctx, useCase.location, options...)
		assert.Nil(t, err, useCase.description)
		for _, object := range objects {
			content := ""
			if !object.IsDir() {
				reader, err := manager.Open(ctx, object)
				assert.Nil(t, err)
				if data, err := ioutil.ReadAll(reader); err == nil {
					content = string(data)
				}
			}
			actuals[object.Name()] = content
		}

		for name := range useCase.assets {
			if useCase.matcher != nil {
				info := NewInfo(name, 0, 0644, time.Now(), false)
				if !useCase.matcher("", info) {
					continue
				}
			}
			_, ok := actuals[name]
			assert.True(t, ok, fmt.Sprintf(useCase.description+"  %v, %v", name, actuals))
		}

	}

}
