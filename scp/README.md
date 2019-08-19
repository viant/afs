# SCP - scp base storage 

- [Usage](#usage)
- [Options](#options)

### Usage

- **[Service](../service.go)**

```go
func main() {
	auth, err := scp.LocalhostKeyAuth("")
	if err != nil {
		log.Fatal(err)
	}
	service := afs.New()
	ctx := context.Background()
	reader, err := service.DownloadWithURL(ctx, "scp://127.0.0.1:22/etc/hosts", auth, option.NewTimeout(2000))
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("data: %s\n", data)

    //next call with the same base URL reuses auth unless Close is called
    reader, err := service.DownloadWithURL(ctx, "scp://127.0.0.1:22/etc/hosts")

}
```

- **[Storager](../storage/storager.go)**

```go
func main() {

	var config *ssh.ClientConfig

	//load config ...
	
	
	timeoutMS := 15000
	service, err := scp.NewStorager("127.0.0.1:22", timeoutMS, config)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	location := "/tmp/myfile"
	err = service.Upload(ctx,  location, 0644, []byte("somedata"))
	if err != nil {
		log.Fatal(err)
	}
	reader, err := service.Download(ctx,  location)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(reader)
	fmt.Printf("data: %s\n", data)

	has, _ := service.Exists(ctx, location)
	fmt.Printf("%v %v", location, has)

	files, err  := service.List(ctx, location, 3)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fmt.Printf("file: %v\n", file.Name())
	}
	err = service.Delete(ctx, location)
	if err != nil {
		log.Fatal(err)
	}

}	
```

### Options


- **[AuthProvider](auth.go)**

```go
    var authProvider AuthProvider
    //... initialise AuthProvider

    service := afs.New()
    ctx := context.Background()
    reader, err := service.DownloadWithURL(ctx, "scp://127.0.0.1:22/etc/hosts", authProvider, option.NewTimeout(2000))
    if err != nil {
        log.Fatal(err)
    }
    
```

- **[BasicAuth](../option/cred.go)**

```go
    auth := option.NewBasicAuth("myuser", "nypass")
	service := afs.New()
	ctx := context.Background()
	reader, err := service.DownloadWithURL(ctx, "scp://127.0.0.1:22/etc/hosts", auth, option.NewTimeout(2000))
	if err != nil {
		log.Fatal(err)
	}
```

- **[KeyAuth](auth.go)**

```go
    auth, err := scp.LocalhostKeyAuth("pathToMyAuthorizedKeyContainer_id_rsa")
	if err != nil {
		log.Fatal(err)
	}
	service := afs.New()
	ctx := context.Background()
	reader, err := service.DownloadWithURL(ctx, "scp://127.0.0.1:22/etc/hosts", auth)
	if err != nil {
		log.Fatal(err)
	}

```

- **[Timeout](../option/timeout.go)**

