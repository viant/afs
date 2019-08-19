package tar_test

import (
	"context"
	"github.com/viant/afs"
	"github.com/viant/afs/file"
	"github.com/viant/afs/tar"
	"log"
)

func ExampleNewWalker() {
	ctx := context.Background()
	service := afs.New()
	walker := tar.NewWalker(file.New())
	err := service.Copy(ctx, "/tmp/test.tar", "mem://dest/folder/test", walker)
	if err != nil {
		log.Fatal(err)
	}

}

func ExampleNewBatchUploader() {
	ctx := context.Background()
	service := afs.New()
	uploader := tar.NewBatchUploader(file.New())
	err := service.Copy(ctx, "/tmp/test/data", "/tmp/data.tar", uploader)
	if err != nil {
		log.Fatal(err)
	}
}
