package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/viant/afs"
)

func debug(i os.FileInfo) {
	log.Printf("Name(): %s IsDir(): %t Size(): %d", i.Name(), i.IsDir(), i.Size())
}

func main() {
	var path, subPath, destPath string
	flag.StringVar(&path, "path", "", "path")
	flag.StringVar(&subPath, "subPath", "localhost", "subPath")
	flag.StringVar(&destPath, "destPath", "", "destPath")
	flag.Parse()

	if path == "" {
		log.Fatal("path is required")
	}

	fullPath := fmt.Sprintf("%s/zip://%s", path, subPath)

	service := afs.New()
	ctx := context.Background()

	objects, err := service.List(ctx, fullPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, object := range objects {
		log.Printf("object: %+v", object)
		debug(object)
	}

	if destPath == "" {
		return
	}

	log.Printf("Copying %s to %s", fullPath, destPath)

	err = service.Copy(ctx, fullPath, destPath)
	if err != nil {
		log.Fatal(err)
	}

	objects, err = service.List(ctx, destPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, object := range objects {
		log.Printf("object: %+v", object)
		debug(object)
	}
}
