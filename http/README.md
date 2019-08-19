# Http storage 

This package defines http base storage

- [Usage](#usage)
- [Options](#options)
    * [Http Client Provider](#http-client-provider)
    * [Basic Auth](#basic-auth)
    * [Custom Header](#custom-header)
    * [Response](#response)

### Usage

### afs.Service

```go

    ctx := context.Background()
    service := afs.New()
    service.Copy()
    
    reader, err := service.DownloadWithURL(ctx, URL)
    err := service.Create(ctx, useCase.URL, 0744, false, reader)
    err := service.Upload(ctx, useCase.URL, 0744, reader)
    err := service.Delete(ctx, URL)
```





### Options

##### Http Client Provider

```go

    ctx := context.Background()
    var clientProvider = func(baseURL string, options ...storage.Option) (*http.Client, error) {
        return http.DefaultClient, nil 
    }
    service := http.New(clientProvider)
    reader, err := service.DownloadWithURL(ctx, URL)
    err := service.Delete(ctx, URL, clientProvider)

```

##### Basic Auth

```go

    ctx := context.Background()
    authProvider :=  option.NewBasicAuth("user", "password")
    service := http.New()
    reader, err := service.DownloadWithURL(ctx, URL, authProvider)
```

##### Custom Header

```go

    ctx := context.Background()
    header := htttp.Header{}
    header.Set("Set-Cookie", "id=a3fWa; Expires=Wed, 21 Oct 2035 07:28:00 GMT; Secure; HttpOnly")
    reader, err := manager.DownloadWithURL(ctx, URL, header)
    

```

##### Response

```go
    ctx := context.Background()
    response := &http.Response{}
    service := http.New()
    reader, err := service.DownloadWithURL(ctx, URL, response)
    err := service.Create(ctx, useCase.URL, 0744, false, reader, response)
    err := service.Upload(ctx, useCase.URL, 0744, reader, response)
    err := service.Delete(ctx, URL. response)
   
```
