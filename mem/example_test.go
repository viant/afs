package mem_test

import (
	"context"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/afs/mem"
	"log"
	"strings"
)

func ExampleService() {
	service := afs.New()
	ctx := context.Background()
	err := service.Upload(ctx, "mem://localhost/folder1/asset.txt", 0644, strings.NewReader("some content"))
	if err != nil {
		log.Fatal(err)
	}
	objects, err := service.List(ctx, "mem://localhost/folder1/")
	if err != nil {
		log.Fatal(err)
	}
	for _, object := range objects {
		fmt.Printf("%v %v\n", object.URL(), object.Name())
	}
}

func ExampleNew() {
	manager := mem.New()
	ctx := context.Background()
	err := manager.Upload(ctx, "mem://localhost/folder1/asset.txt", 0644, strings.NewReader("some content"))
	if err != nil {
		log.Fatal(err)
	}
	objects, err := manager.List(ctx, "mem://localhost/folder1/")
	if err != nil {
		log.Fatal(err)
	}
	for _, object := range objects {
		fmt.Printf("%v %v\n", object.URL(), object.Name())
	}
}

func ExampleNewStorager() {
	ctx := context.Background()
	storager := mem.NewStorager("mem://localhost/")
	err := storager.Upload(ctx, "folder1/asset1", 0644, strings.NewReader("some content"))
	if err != nil {
		log.Fatal(err)
	}
	err = storager.Upload(ctx, "folder1/asset2", 0644, strings.NewReader("some content"))
	if err != nil {
		log.Fatal(err)
	}

	fileInfos, err := storager.List(ctx, "folder1/", 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	for _, info := range fileInfos {
		fmt.Printf("%v\n", info.Name())
	}

}
