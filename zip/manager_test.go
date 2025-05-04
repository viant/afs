package zip_test

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/afs"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
)

type useCaseFn func(s afs.Service, ctx context.Context, url string) ([]storage.Object, error)

func TestNew(t *testing.T) {
	testCases(t, func(service afs.Service, ctx context.Context, url string) ([]storage.Object, error) {
		return service.List(ctx, url)
	})
}

func TestNoCache(t *testing.T) {
	testCases(t, func(service afs.Service, ctx context.Context, url string) ([]storage.Object, error) {
		return service.List(ctx, url, &option.NoCache{Source: option.NoCacheBaseURL})
	})
}

func testCases(t *testing.T, callList useCaseFn) {
	_, filename, _, _ := runtime.Caller(0)
	baseDir, _ := path.Split(filename)
	ctx := context.Background()

	var useCases = []struct {
		description string
		URL         string
		expect      map[string]bool
		isDir       bool
	}{
		{
			description: "list war web-inf directory",
			URL:         fmt.Sprintf("file:%vtest/app.war/zip://localhost/WEB-INF", baseDir),
			expect: map[string]bool{
				// this is the listed directory
				"WEB-INF": true,

				"web.xml":           true,
				"HelloWorld.class":  true,
				"config.properties": true,
			},
			isDir: true,
		},

		{
			description: "list war classes directory",
			URL:         fmt.Sprintf("file:%vtest/app.war/zip://localhost/WEB-INF/classes", baseDir),
			expect: map[string]bool{
				// this is the listed directory
				"classes": true,

				"HelloWorld.class":  true,
				"config.properties": true,
			},
			isDir: true,
		},

		{
			description: "list war file only",
			URL:         fmt.Sprintf("file:%vtest/app.war/zip://localhost/WEB-INF/classes/config.properties", baseDir),
			expect: map[string]bool{
				"config.properties": true,
			},
		},
	}

	for _, useCase := range useCases {
		service := afs.New()
		objects, err := callList(service, ctx, useCase.URL)
		assert.Nil(t, err, useCase.description)

		cleanObjects := make([]storage.Object, 0)
		for _, obj := range objects {
			if obj.Name() == "" {
				continue
			}
			cleanObjects = append(cleanObjects, obj)
		}

		assert.EqualValues(t, len(useCase.expect), len(cleanObjects))
		for _, obj := range cleanObjects {
			assert.True(t, useCase.expect[obj.Name()], fmt.Sprintf("%s missing [%s] [%s]", useCase.description, obj.Name(), obj.URL()))
		}

	}
}

func TestCopyFromZip(t *testing.T) {
	ctx := context.Background()

	service := afs.New()

	// Create a temporary file for the zip
	tempFile, err := os.CreateTemp("", "test_nosubdir_*.zip")
	assert.Nil(t, err)

	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Create a new zip writer
	zipWriter := zip.NewWriter(tempFile)

	// Add files to the zip
	files := []struct {
		path    string
		content string
	}{
		{"a.txt", "a"},
		{"b.txt", "b"},
	}

	// Write each file to the zip
	for _, file := range files {
		w, err := zipWriter.Create(file.path)
		assert.Nil(t, err)
		_, err = w.Write([]byte(file.content))
		assert.Nil(t, err)
	}

	// Close the zip writer to flush the contents
	err = zipWriter.Close()
	assert.Nil(t, err)

	srcURL := fmt.Sprintf("file:%s/zip://localhost", tempFile.Name())

	// Verify zip contents
	zipObjects, err := service.List(ctx, srcURL)
	assert.Nil(t, err)
	// 3 objects - root dir, 2 files
	assert.EqualValues(t, 3, len(zipObjects), "expected 3 objects in zip, got %+V", zipObjects)

	// Manually create files like a ZIP might but without using ZIP
	manualURL := path.Join(os.TempDir(), "nosubdir_manual/")
	os.MkdirAll(manualURL, 0755)
	for _, file := range files {
		os.WriteFile(path.Join(manualURL, file.path), []byte(file.content), 0644)
	}

	// Create a temporary destination directory
	destURL := path.Join(os.TempDir(), "nosubdir_test/")
	defer os.RemoveAll(destURL)

	// Copy from zip to filesystem
	err = service.Copy(ctx, srcURL, destURL)
	assert.Nil(t, err)

	// Verify destination contents
	objects, err := service.List(ctx, destURL)
	assert.Nil(t, err)
	// 3 objects - root dir, 2 files
	assert.EqualValues(t, 3, len(objects), "expected 3 objects in dest, got %+V", objects)
}

func TestCreateZipWithoutDirEntries(t *testing.T) {
	ctx := context.Background()
	service := afs.New()

	// Test creating a zip with depth 2 directory structure without explicit directory entries
	// Create a temporary file for the zip
	tempFile, err := os.CreateTemp("", "test_no_dir_entries_*.zip")
	assert.Nil(t, err)
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Create a new zip writer
	zipWriter := zip.NewWriter(tempFile)
	defer zipWriter.Close()

	// Add files with directory paths but no explicit directory entries
	files := []struct {
		path    string
		content string
	}{
		{"dir1/file1.txt", "content of file1"},
		{"dir1/file2.txt", "content of file2"},
		{"dir1/subdir/file3.txt", "content of file3"},
		{"dir1/subdir/file4.txt", "content of file4"},
		{"dir2/file5.txt", "content of file5"},
	}

	// Write each file to the zip
	for _, file := range files {
		w, err := zipWriter.Create(file.path)
		assert.Nil(t, err)
		_, err = w.Write([]byte(file.content))
		assert.Nil(t, err)
	}

	// Close the zip writer to flush the contents
	err = zipWriter.Close()
	assert.Nil(t, err)

	// Verify the zip contents
	zipURL := fmt.Sprintf("file:%s/zip://localhost", tempFile.Name())
	// this test unfortunately is invalid since we're changing how List works...
	objects, err := service.List(ctx, zipURL)
	assert.Nil(t, err)

	// We should see both directories and files, even though we didn't explicitly create directory entries
	expectedPaths := map[string]bool{
		"file1.txt": true,
		"file2.txt": true,
		"file3.txt": true,
		"file4.txt": true,
		"file5.txt": true,
	}

	foundPaths := map[string]bool{}

	// Check that we can find all expected paths
	for _, obj := range objects {
		name := obj.Name()
		if name == "" {
			// directories don't get names in List()
			continue
		}

		assert.True(t, expectedPaths[name], "Found unexpected object: %v, url: %v", name, obj.URL())
		assert.False(t, foundPaths[name], "Found repeat object: %v, url: %v", name, obj.URL())
		foundPaths[name] = true
	}

	// Check that we found all expected paths
	assert.Equal(t, len(expectedPaths), len(foundPaths), "Some expected paths were not found: %v ", foundPaths)
}
