# mem - in memory file storage

This package defines in memory file storage.

### Usage

- **[Service](../service.go)**
```go
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
``` 

- **[Manager](../storage/manager.go)**

```go
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
```

- **[Storager](../storage/storager.go)**

```go
func main() {
	ctx := context.Background()
    storager := mem.NewStorager("mem://localhost/")
    err := storager.Upload(ctx, "folder1/asset1", 0644, []byte("some content"))
    if err != nil {
        log.Fatal(err)
    }
    err = storager.Upload(ctx, "folder1/asset2", 0644, []byte("some content"))
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
```

### Options
 
 - [Errors](../option/error.go)