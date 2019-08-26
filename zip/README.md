# zip - zip archives

## Usage


* **Service**

```go
    service := afs.New()
    ctx := context.Background()
    objects, err := service.List(ctx, "file:/tmp/app.war/zip://localhost/WEB-INF")
    if err != nil {
        log.Fatal(err)
    }
    for _, object := range objects {
        fmt.Printf("%v %v\n", object.Name(), object.URL())
        if object.IsDir() {
            continue
        }
        reader, err := service.Download(ctx, object)
        if err != nil {
            log.Fatal(err)
        }
        data, err := ioutil.ReadAll(reader)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("%s\n", data)
    }
```

* **Walker**

```go
    ctx := context.Background()
	service := afs.New()
	walker := zip.NewWalker(file.New())
	err := service.Copy(ctx, "/tmp/test.zip", "mem://dest/folder/test", walker)
	if err != nil {
		log.Fatal(err)
	}
```

* **Uploader**

```go
    ctx := context.Background()
	service := afs.New()
	uploader := zip.NewBatchUploader(file.New())
	err := service.Copy(ctx, "/tmp/test/data", "/tmp/data.zip", uploader)
	if err != nil {
		log.Fatal(err)
	}
```
