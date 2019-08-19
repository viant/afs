# tar - tar archives 

## Usage

* **Walker**

```go
    ctx := context.Background()
	service := afs.New()
	walker := tar.NewWalker(file.New())
	err := service.Copy(ctx, "/tmp/test.tar", "mem://dest/folder/test", walker)
	if err != nil {
		log.Fatal(err)
	}
```

* **Uploader**

```go
    ctx := context.Background()
	service := afs.New()
	uploader := tar.NewBatchUploader(file.New())
	err := service.Copy(ctx, "/tmp/test/data", "/tmp/data.tar", uploader)
	if err != nil {
		log.Fatal(err)
	}
```
