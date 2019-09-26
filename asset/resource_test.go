package asset_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/file"
	"io/ioutil"
	"testing"
)

func TestNew(t *testing.T) {
	mgr := file.New()
	baseURL := "file://localhost/tmp/assets"
	_ = asset.Cleanup(mgr, baseURL)

	err := asset.Create(mgr, baseURL, []*asset.Resource{
		asset.NewFile("file1.txt", []byte("123"), 0644),
		asset.NewDir("dir1", 0755),
		asset.NewLink("file2.txt", "file1.txt", 0644),
	})
	assert.Nil(t, err)
	resources, err := asset.Load(mgr, baseURL)
	assert.Nil(t, err)
	assert.NotNil(t, resources["file1.txt"])

	resource := resources["file1.txt"]
	assert.EqualValues(t, resource.Info().Name(), "file1.txt")
	data, err := ioutil.ReadAll(resource.Reader())
	assert.Nil(t, err)
	assert.EqualValues(t, "123", data)
	err = resource.MergeFrom(resource)
	assert.Nil(t, err)
	assert.NotNil(t, resources["file2.txt"])
	assert.NotNil(t, resources["dir1"])

}
