package scp

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/file"
	"io"
	"io/ioutil"
	_ "net/http/pprof"
	"os"
	"path"
	"testing"
)

func TestSession_initCmd(t *testing.T) {

	var useCases = []struct {
		description string
		mode        int
		recursive   bool
		location    string
		expect      string
	}{
		{
			description: "write recursive mode",
			mode:        modeWrite,
			recursive:   true,
			location:    "/tmp",
			expect:      "scp -t -p -r /tmp\n",
		},
		{
			description: "write mode",
			mode:        modeWrite,
			location:    "/tmp",
			expect:      "scp -t -p /tmp\n",
		},
		{
			description: "readInBackground recursive mode",
			mode:        modeRead,
			recursive:   true,
			location:    "/tmp",
			expect:      "scp -f -p -r /tmp\n",
		},
		{
			description: "readInBackground mode",
			mode:        modeRead,
			location:    "/tmp",
			expect:      "scp -f -p /tmp\n",
		},
	}

	for _, useCase := range useCases {
		session := &session{mode: useCase.mode, recursive: useCase.recursive}
		actual := session.initCmd(useCase.location)
		assert.EqualValues(t, useCase.expect, actual, useCase.description)
	}

}

func TestSession_download(t *testing.T) {

	client, err := newTestClient("127.0.0.1:22")
	if err != nil {
		t.Skipf("unable to create ssh client %v", err)
		return
	}
	baseDir := os.TempDir()

	var useCases = []struct {
		description string
		location    string
		recursive   bool
		file        string
		assets      []*asset.Resource
		hasError    bool
	}{

		{
			description: "multi download from baseLocation with symlink",
			location:    path.Join(baseDir, "scp_download_04"),
			file:        path.Join(baseDir, "scp_download_04"),
			recursive:   true,
			assets: []*asset.Resource{
				asset.NewFile("foo1.txt", []byte("abc"), 0644),
				asset.NewFile("foo2.txt", []byte("xyz"), 0644),
				asset.NewLink("sym.txt", "foo1.txt", 0644),
			},
		},

		{
			description: "location single download",
			location:    path.Join(baseDir, "scp_download_01"),
			assets: []*asset.Resource{
				asset.NewFile("foo.txt", []byte("abc"), 0644),
			},
			file: path.Join(baseDir, "scp_download_01", "foo.txt"),
		},
		{
			description: "recursive download from baseLocation",
			location:    path.Join(baseDir, "scp_download_02"),
			file:        path.Join(baseDir, "scp_download_02/*"),
			recursive:   true,
			assets: []*asset.Resource{
				asset.NewDir("subfolder", 0744),
				asset.NewFile("foo1.txt", []byte("abc"), 0644),
				asset.NewFile("foo2.txt", []byte("xyz"), 0644),
			},
		},
		{
			description: "missing location",
			location:    path.Join(baseDir, "scp_download_03"),
			file:        path.Join(baseDir, "scp_download_03", "error.txt"),
			hasError:    true,
			assets: []*asset.Resource{
				asset.NewFile("foo.txt", []byte("abc"), 0644),
			},
		},

		{
			description: "multi download from baseLocation - not regular location error",
			location:    path.Join(baseDir, "scp_download_05"),
			file:        path.Join(baseDir, "scp_download_05/*"),
			recursive:   false,
			hasError:    true,
			assets: []*asset.Resource{
				asset.NewDir("subfolder", 0744),
				asset.NewFile("foo1.txt", []byte("abc"), 0644),
				asset.NewFile("foo2.txt", []byte("xyz"), 0644),
			},
		},
	}

	ctx := context.Background()
	fileManager := file.New()

	for _, useCase := range useCases {
		files := map[string][]byte{}
		err = asset.Create(fileManager, useCase.location, useCase.assets)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		session, err := newSession(client, modeRead, useCase.recursive, 0)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		err = session.download(ctx, false, useCase.file,
			func(relative string, info os.FileInfo, reader io.Reader) (bool, error) {
				if reader == nil {
					files[info.Name()] = nil
					return true, nil
				}
				files[info.Name()], err = ioutil.ReadAll(reader)
				return true, err
			})
		_ = session.Close()
		if useCase.hasError {
			assert.NotNil(t, err, useCase.description)
			continue
		}
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		for _, asset := range useCase.assets {
			if asset.Dir || asset.Link != "" {
				continue
			}
			data, ok := files[asset.Name]
			assert.True(t, ok, useCase.description+" "+asset.Name)
			assert.EqualValues(t, string(data), string(asset.Data), useCase.description+" "+asset.Name)
		}

	}
	for _, useCase := range useCases {
		err = asset.Cleanup(fileManager, useCase.location)
		assert.Nil(t, err)
	}
}

func TestSession_upload(t *testing.T) {
	client, err := newTestClient("127.0.0.1:22")
	if err != nil {
		t.Skipf("unable to create ssh client %v", err)
		return
	}
	baseDir := os.TempDir()
	fileManager := file.New()
	var useCases = []struct {
		description    string
		baseLocation   string
		recursive      bool
		location       string
		createLocation bool
		assets         []*asset.Resource
	}{
		{
			description:    "hidden location upload",
			recursive:      true,
			createLocation: true,
			baseLocation:   path.Join(baseDir, "scp_upload_00/"),
			location:       path.Join(baseDir, "scp_upload_00/"),
			assets: []*asset.Resource{
				asset.NewDir(".bin", 0644),
				asset.NewFile(".bin/foo.txt", []byte("abc"), 0644),
			},
		},
		//{
		//	description:    "single file location upload",
		//	recursive:      true,
		//	createLocation: true,
		//	baseLocation:   path.Join(baseDir, "scp_upload_01"),
		//	location:       path.Join(baseDir, "scp_upload_01", "foo.txt"),
		//	assets: []*asset.Resource{
		//		asset.NewFile("foo.txt", []byte("abc"), 0644),
		//	},
		//},
		//{
		//	description:    "multi file location upload",
		//	recursive:      true,
		//	createLocation: true,
		//	baseLocation:   path.Join(baseDir, "scp_upload_02"),
		//	location:       path.Join(baseDir, "scp_upload_02"),
		//	assets: []*asset.Resource{
		//		asset.NewFile("foo1.txt", []byte("abc"), 0644),
		//		asset.NewFile("foo2.txt", []byte("xyz"), 0644),
		//	},
		//},
		//
		//{
		//	description:    "multi download from baseLocation",
		//	baseLocation:   path.Join(baseDir, "scp_upload_03"),
		//	location:       path.Join(baseDir, "scp_upload_03"),
		//	createLocation: true,
		//	recursive:      true,
		//	assets: []*asset.Resource{
		//		asset.NewFile("foo1.txt", []byte("abc"), 0644),
		//		asset.NewFile("foo2.txt", []byte("xyz"), 0644),
		//		asset.NewDir("sub", 0744),
		//		asset.NewFile("sub/bar1.txt", []byte("xyz"), 0644),
		//		asset.NewFile("sub/bar2.txt", []byte("xyz"), 0644),
		//		asset.NewFile("bar.txt", []byte("xyz"), 0644),
		//	},
		//},
		//{
		//	description:    "multi download - unordered",
		//	baseLocation:   path.Join(baseDir, "scp_upload_05"),
		//	location:       path.Join(baseDir, "scp_upload_05"),
		//	createLocation: true,
		//	recursive:      true,
		//	assets: []*asset.Resource{
		//		asset.NewDir("test", 0744),
		//		asset.NewDir("test/folder2/sub", 0744),
		//
		//		asset.NewFile("test/asset1.txt", []byte("xyz"), 0644),
		//		asset.NewFile("test/asset2.txt", []byte("xyz"), 0644),
		//		asset.NewDir("test/folder1", 0744),
		//		asset.NewFile("test/folder1/res.txt", []byte("xyz"), 0644),
		//		asset.NewFile("test/folder2/res1.txt", []byte("xyz"), 0644),
		//	},
		//},
		//{
		//	description:    "multi download from baseLocation - 2 depth",
		//	baseLocation:   path.Join(baseDir, "scp_upload_04"),
		//	location:       path.Join(baseDir, "scp_upload_04"),
		//	createLocation: true,
		//	recursive:      true,
		//	assets: []*asset.Resource{
		//		asset.NewFile("foo1.txt", []byte("abc"), 0644),
		//		asset.NewDir("s1", 0744),
		//		asset.NewFile("s1/bar1.txt", []byte("xyz"), 0644),
		//		asset.NewDir("s1/s2", 0744),
		//		asset.NewFile("s1/s2/bar1.txt", []byte("xyz"), 0644),
		//		asset.NewFile("s1/bar2.txt", []byte("xyz"), 0644),
		//		asset.NewFile("foo2.txt", []byte("abc"), 0644),
		//	},
		//},
	}

	for _, useCase := range useCases {
		ctx := context.Background()
		session, err := newSession(client, modeWrite, useCase.recursive, 0)
		assert.Nil(t, err, useCase.description)
		_ = asset.Cleanup(fileManager, useCase.baseLocation)

		if useCase.createLocation {
			_ = fileManager.Create(ctx, useCase.baseLocation, 0744, true)
		}

		fmt.Printf("%v\n", useCase.location)

		uploader, closer, err := session.upload(useCase.location)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		for _, asset := range useCase.assets {
			relative, _ := path.Split(asset.Name)
			var reader io.Reader
			if len(asset.Data) > 0 {
				reader = bytes.NewReader(asset.Data)
			}
			err = uploader(ctx, relative, asset.Info(), reader)
			assert.Nil(t, err, useCase.description)
		}

		actuals, err := asset.Load(fileManager, useCase.location)

		assert.Nil(t, err, useCase.description)
		for _, asset := range useCase.assets {
			_, ok := actuals[asset.Name]
			assert.True(t, ok, useCase.description+" "+asset.Name)
		}

		err = closer.Close()
		assert.Nil(t, err, useCase.description)

	}
	for _, useCase := range useCases {
		_ = asset.Cleanup(fileManager, useCase.baseLocation)
	}
}
