# file - file storage 

This package provides a file system manager that wraps os operations.


### Usage

- **[Service](../service.go)**
```go
func main() {
	service := afs.New()
	ctx := context.Background()
	err := service.Upload(ctx, "file:///folder1/asset.txt", 0644, strings.NewReader("some content"))
	if err != nil {
		log.Fatal(err)
	}
	objects, err := service.List(ctx, "file:///folder1/")
	if err != nil {
		log.Fatal(err)
	}
	for _, object := range objects {
		fmt.Printf("%v %v\n", object.URL(), object.Name())
	}
}
``` 


