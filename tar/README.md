# tar - tar archives 

## Usage



* **Service**

```go
    service := afs.New()
    ctx := context.Background()
    objects, err := service.List(ctx, "file:/tmp/app.war/tar://localhost/WEB-INF")
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
