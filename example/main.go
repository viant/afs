package main

import (
	"context"
	"github.com/viant/afs"
	"github.com/viant/afs/modifier"
	"github.com/viant/afs/option"
	"github.com/viant/afsc/gs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
)

func main() {
	_, filename, _, _ := runtime.Caller(0)
	baseDir, _ := path.Split(filename)
	runArchiveSubstitution(baseDir)
	copyIntoArchive(baseDir)
}

func copyIntoArchive(baseDir string) {
	secretPath := path.Join(os.Getenv("HOME"), ".secret", "gcp-e2e.json")
	auth, err := gs.NewJwtConfig(option.NewLocation(secretPath))
	if err != nil {
		return
	}
	sourceURL := path.Clean(path.Join(baseDir, "../base"))
	destURL := "gs:e2etst/test.zip/zip://localhost/dir1"
	service := afs.New()
	ctx := context.Background()
	err = service.Copy(ctx, sourceURL, destURL, option.NewDest(auth))
	if err != nil {
		log.Fatal(err)
	}
}

func runArchiveSubstitution(baseDir string) {
	testFile := path.Join(baseDir, "test/app.war")
	if data, err := ioutil.ReadFile(testFile); err == nil {
		_ = ioutil.WriteFile("/tmp/app.war", data, 0644)
	}
	service := afs.New()
	ctx := context.Background()
	sourceURL := "file:/tmp/app.war/zip://localhost/WEB-INF/classes/config.properties"
	destURL := "file:/tmp/app.war/zip://localhost/"
	err := service.Copy(ctx, sourceURL, destURL, modifier.Replace(map[string]string{
		"${changeMe}": os.Getenv("USER"),
	}))
	if err != nil {
		log.Fatal(err)
	}
}
