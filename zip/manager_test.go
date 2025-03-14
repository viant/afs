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
	}{
		{
			description: "list war classes",
			URL:         fmt.Sprintf("file:%v/test/app.war/zip://localhost/WEB-INF/classes", baseDir),
			expect: map[string]bool{
				"classes":           true,
				"HelloWorld.class":  true,
				"config.properties": true,
			},
		},

		{
			description: "list war classes",
			URL:         fmt.Sprintf("file:%v/test/app.war/zip://localhost/WEB-INF/classes/config.properties", baseDir),
			expect: map[string]bool{
				"config.properties": true,
			},
		},
	}

	for _, useCase := range useCases {
		service := afs.New()
		objects, err := callList(service, ctx, useCase.URL)
		assert.Nil(t, err, useCase.description)
		assert.EqualValues(t, len(useCase.expect), len(objects))
		for _, obj := range objects {
			assert.True(t, useCase.expect[obj.Name()], useCase.description+" "+obj.URL())
		}

	}
}

func TestCopy(t *testing.T) {
	ctx := context.Background()

	service := afs.New()

	t.Run("copy_from_zip_to_filesystem", func(t *testing.T) {
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

		// Create a temporary destination directory
		destURL := path.Join(os.TempDir(), "nosubdir_test")
		defer os.RemoveAll(destURL)

		// Copy from zip to filesystem
		srcURL := fmt.Sprintf("file:%s/zip://localhost", tempFile.Name())
		err = service.Copy(ctx, srcURL, destURL)
		assert.Nil(t, err)

		// Verify zip contents
		zipObjects, err := service.List(ctx, srcURL)
		assert.Nil(t, err)
		assert.EqualValues(t, 2, len(zipObjects), "expected 2 objects in zip")

		// Verify destination contents
		objects, err := service.List(ctx, destURL)
		assert.Nil(t, err)
		assert.EqualValues(t, 2, len(objects), "expected 2 objects in dest")
	})

	// Test creating a zip with depth 2 directory structure without explicit directory entries
	t.Run("create_zip_without_dir_entries", func(t *testing.T) {
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
		objects, err := service.List(ctx, zipURL)
		assert.Nil(t, err)

		// We should see both directories and files, even though we didn't explicitly create directory entries
		expectedPaths := map[string]bool{
			"dir1":        true,
			"dir2":        true,
			"dir1/subdir": true,
			"file1.txt":   true,
			"file2.txt":   true,
			"file3.txt":   true,
			"file4.txt":   true,
			"file5.txt":   true,
		}

		// Check that we can find all expected paths
		for _, obj := range objects {
			name := obj.Name()
			assert.True(t, expectedPaths[name], "Found unexpected object: "+name)
			delete(expectedPaths, name)
		}

		// Check that we found all expected paths
		assert.Equal(t, 0, len(expectedPaths), "Some expected paths: %v were not found", expectedPaths)
	})
}
