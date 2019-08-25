package archive

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/base"
	"github.com/viant/afs/mem"
	"github.com/viant/afs/walker"
	"testing"
)

func TestRewrite(t *testing.T) {

	ctx := context.Background()
	mgr := mem.New()

	var useCases = []struct {
		description string
		source      string
		dest        string
		resources   []*asset.Resource
		modifier    Modifier
		expect      map[string]*asset.Resource
	}{
		{
			description: "rewrite all",
			source:      "mem://localhost/rewrite_001/folder1",
			dest:        "mem://localhost/rewrite_001/folder2",
			resources: []*asset.Resource{
				asset.NewFile("res1.txt", []byte("abc"), 0644),
				asset.NewDir("folder", 0750),
				asset.NewFile("folder/res2.txt", []byte("xyz"), 0644),
			},
			modifier: func(resources []*asset.Resource) ([]*asset.Resource, error) {
				return resources, nil
			},
			expect: map[string]*asset.Resource{
				"res1.txt":        asset.NewFile("res1.txt", []byte("abc"), 0644),
				"folder":          asset.NewDir("folder", 0750),
				"folder/res2.txt": asset.NewFile("folder/res2.txt", []byte("xyz"), 0644),
			},
		},
		{
			description: "delete folder resource",
			source:      "mem://localhost/rewrite_002/folder1",
			dest:        "mem://localhost/rewrite_002/folder2",
			resources: []*asset.Resource{
				asset.NewFile("res1.txt", []byte("abc"), 0644),
				asset.NewDir("f1", 0750),
				asset.NewFile("f1/res2.txt", []byte("xyz"), 0644),
				asset.NewFile("f1/res3.txt", []byte("xyz"), 0644),
				asset.NewDir("f2", 0750),
				asset.NewDir("f2/sub", 0750),
				asset.NewFile("f2/sub/res4.txt", []byte("xyz"), 0644),
				asset.NewFile("f2/sub/res5.txt", []byte("xyz"), 0644),
			},
			modifier: DeleteHandler("f2/sub"),
			expect: map[string]*asset.Resource{
				"res1.txt":    asset.NewFile("res1.txt", []byte("abc"), 0644),
				"f1":          asset.NewDir("f1", 0750),
				"f1/res2.txt": asset.NewFile("f1/res2.txt", []byte("xyz"), 0644),
				"f1/res3.txt": asset.NewFile("f1/res3.txt", []byte("xyz"), 0644),
				"f2":          asset.NewDir("f2", 0750),
			},
		},

		{
			description: "override resource",
			source:      "mem://localhost/rewrite_003/folder1",
			dest:        "mem://localhost/rewrite_003/folder2",
			resources: []*asset.Resource{
				asset.NewFile("res1.txt", []byte("abc1"), 0644),
				asset.NewFile("res2.txt", []byte("a"), 0644),
				asset.NewFile("res3.txt", []byte("abc3"), 0644),
			},
			modifier: CreateHandler("res2.txt", 0643, []byte("abc20"), false),
			expect: map[string]*asset.Resource{
				"res1.txt": asset.NewFile("res1.txt", []byte("abc1"), 0644),
				"res2.txt": asset.NewFile("res2.txt", []byte("abc20"), 0643),
				"res3.txt": asset.NewFile("res3.txt", []byte("abc3"), 0644),
			},
		},
		{
			description: "upload resource",
			source:      "mem://localhost/rewrite_004/folder1",
			dest:        "mem://localhost/rewrite_004/folder2",
			resources: []*asset.Resource{
				asset.NewFile("res1.txt", []byte("abc1"), 0644),
				asset.NewFile("res3.txt", []byte("abc3"), 0644),
			},
			modifier: CreateHandler("res2.txt", 0643, []byte("abc20"), false),
			expect: map[string]*asset.Resource{
				"res1.txt": asset.NewFile("res1.txt", []byte("abc1"), 0644),
				"res2.txt": asset.NewFile("res2.txt", []byte("abc20"), 0643),
				"res3.txt": asset.NewFile("res3.txt", []byte("abc3"), 0644),
			},
		},

		{
			description: "upload resource with partial path match",
			source:      "mem://localhost/rewrite_005/folder1",
			dest:        "mem://localhost/rewrite_005/folder2",
			resources: []*asset.Resource{
				asset.NewDir("f1", 0755),
				asset.NewDir("f1/f2", 0755),

				asset.NewFile("f1/f2/res1.txt", []byte("abc1"), 0644),
				asset.NewFile("f1/f2/res3.txt", []byte("abc3"), 0644),
			},
			modifier: CreateHandler("f1/f5/res2.txt", 0643, []byte("abc20"), false),
			expect: map[string]*asset.Resource{
				"f1":             asset.NewDir("f1", 0755),
				"f1/f2":          asset.NewDir("f1/f2", 0755),
				"f1/f2/res1.txt": asset.NewFile("f1/f2/res1.txt", []byte("abc1"), 0644),
				"f1/f5":          asset.NewDir("f1/f5", 0755),
				"f1/f5/res2.txt": asset.NewFile("f1/f5/res2.txt", []byte("abc20"), 0643),
				"f1/f2/res3.txt": asset.NewFile("f1/f2/res3.txt", []byte("abc3"), 0644),
			},
		},

		{
			description: "multi resource upload",
			source:      "mem://localhost/rewrite_006/folder1",
			dest:        "mem://localhost/rewrite_006/folder2",
			resources: []*asset.Resource{
				asset.NewDir("f1", 0755),
				asset.NewDir("f1/f2", 0755),

				asset.NewFile("f1/f2/res1.txt", []byte("abc1"), 0644),
				asset.NewFile("f1/f2/res3.txt", []byte("abc3"), 0644),
			},
			modifier: UploadHandler([]*asset.Resource{
				asset.NewFile("f1/f5/res2.txt", []byte("abc20"), 0643),
				asset.NewFile("f3/res4.txt", []byte("abc20"), 0643),
				asset.NewFile("f1/f5/f6/res5.txt", []byte("abc20"), 0643),
			}),
			expect: map[string]*asset.Resource{
				"f1":                asset.NewDir("f1", 0755),
				"f1/f2":             asset.NewDir("f1/f2", 0755),
				"f3":                asset.NewDir("f3", 0755),
				"f3/res4.txt":       asset.NewFile("f3/res4.txt", []byte("abc20"), 0643),
				"f1/f2/res1.txt":    asset.NewFile("f1/f2/res1.txt", []byte("abc1"), 0644),
				"f1/f5":             asset.NewDir("f1/f5", 0755),
				"f1/f5/f6":          asset.NewDir("f1/f5/f6", 0755),
				"f1/f5/f6/res5.txt": asset.NewFile("f1/f5/f6/res5.txt", []byte("abc20"), 0643),
				"f1/f5/res2.txt":    asset.NewFile("f1/f5/res2.txt", []byte("abc20"), 0643),
				"f1/f2/res3.txt":    asset.NewFile("f1/f2/res3.txt", []byte("abc3"), 0644),
			},
		},
	}

	for _, useCase := range useCases {
		memWalker := walker.New(mgr)
		err := asset.Create(mgr, useCase.source, useCase.resources)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		uploader := base.NewUploader(mgr)
		upload, closer, err := uploader.Uploader(ctx, useCase.dest)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		err = Rewrite(ctx, memWalker, useCase.source, upload, useCase.modifier)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		err = closer.Close()

		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		actuals, err := asset.Load(mgr, useCase.dest)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		assert.EqualValues(t, len(actuals), len(useCase.expect), useCase.description)
		for _, actual := range actuals {
			expect, ok := useCase.expect[actual.Name]
			description := useCase.description + " " + actual.Name
			if !assert.True(t, ok, description) {
				continue
			}
			if !expect.Dir {
				assert.EqualValues(t, expect.Data, actual.Data, description)
			}
			assert.EqualValues(t, expect.Dir, actual.Dir, description)

			assert.EqualValues(t, expect.Mode, actual.Mode, description+" mode: expected: %o, had: %o", expect.Mode, actual.Mode)

		}
	}
}
