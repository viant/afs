package zip_test

import (
	"context"
	"github.com/viant/afs"
	"github.com/viant/afs/file"
	"github.com/viant/afs/zip"
	"log"
)

func ExampleNewWalker() {
	ctx := context.Background()
	service := afs.New()
	walker := zip.NewWalker(file.New())
	err := service.Copy(ctx, "/tmp/test.zip", "mem://dest/folder/test", walker)
	if err != nil {
		log.Fatal(err)
	}

}

func ExampleNewBatchUploader() {
	ctx := context.Background()
	service := afs.New()
	uploader := zip.NewBatchUploader(file.New())
	err := service.Copy(ctx, "/tmp/test/data", "/tmp/data.zip", uploader)
	if err != nil {
		log.Fatal(err)
	}
}
