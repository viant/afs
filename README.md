# afs - abstract file storage

[![GoReportCard](https://goreportcard.com/badge/github.com/viant/afs)](https://goreportcard.com/report/github.com/viant/afs)
[![GoDoc](https://godoc.org/github.com/viant/afs?status.svg)](https://godoc.org/github.com/viant/afs)
![goversion-image](https://img.shields.io/badge/Go-1.11+-00ADD8.svg)
<a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-78%25-brightgreen.svg?longCache=true&style=flat)</a>


Please refer to [`CHANGELOG.md`](CHANGELOG.md) if you encounter breaking changes.

- [Motivation](#motivation)
- [Introduction](#introduction)
- [Usage](#usage)
- [Matchers](#matchers)
- [Content modifiers](#content-modifiers)
- [Options](#options)
- [Storage Implementations](#storage-implementations)
- [Testing mode](#testing-mode)
- [Storage Manager](#storage-managers)
- [GoCover](#gocover)
- [License](#license)
- [Credits and Acknowledgements](#credits-and-acknowledgements)

## Motivation

When dealing with various storage systems, like cloud storage, SCP, container or local file system, using shared API for typical storage operation provides an excellent simplification.
What's more, the ability to simulate storage-related errors like Auth or EOF allows you to test an app error handling.

## Introduction

This library uses a storage manager abstraction to provide an implementation for a specific storage system with following 

* **CRUD Operation:**

```go
List(ctx context.Context, URL string, options ...Option) ([]Object, error)

Walk(ctx context.Context, URL string, handler OnVisit, options ...Option) error

Download(ctx context.Context, object Object, options ...Option) (io.ReadCloser, error)

DownloadWithURL(ctx context.Context, URL string, options ...Option) (io.ReadCloser, error)

Upload(ctx context.Context, URL string, mode os.FileMode, reader io.Reader, options ...Option) error

Create(ctx context.Context, URL string, mode os.FileMode, isContainer bool, options ...Option) error

Delete(ctx context.Context, URL string, options ...Option) error
``` 

* **Batch uploader:**

```go
type Upload func(ctx context.Context, parent string, info os.FileInfo, reader io.Reader) error
 
Uploader(ctx context.Context, URL string, options ...Option) (Upload, io.Closer, error)
```

* **Utilities:**

```go

Copy(ctx context.Context, sourceURL, destURL string, options ...Option) error

Move(ctx context.Context, sourceURL, destURL string, options ...Option) error
```


URL scheme is used to identify storage system, or alternatively relative/absolute path can be used for local file storage.
By default all operations using the same baseURL share the same corresponding storage manager instance.
For example, instead supplying SCP auth details for all operations, auth option can be used only once.

```go

func main() {

    ctx := context.Background()
    {
        //auth with first call 
        service := afs.New()
        defer service.Close()
        keyAuth, err := scp.LocalhostKeyAuth("")
        if err != nil {
           log.Fatal(err)
        }
        reader1, err := service.DownloadWithURL(ctx, "scp://host1:22/myfolder/asset.txt", keyAuth)
        if err != nil {
               log.Fatal(err)
        }
        ...
        reader2, err := service.DownloadWithURL(ctx, "scp://host1:22/myfolder/asset.txt", keyAuth)
    }
    
    {
        //auth per baseURL 
        service := afs.New()
        err = service.Init(ctx, "scp://host1:22/", keyAuth)
        if err != nil {
            log.Fatal(err)
        }
        defer service.Destroy("scp://host1:22/")
        reader, err := service.DownloadWithURL(ctx, "scp://host1:22/myfolder/asset.txt")
     }
}

```

## Usage

* **Downloading location content**

```go
func main() {
	
    service := afs.New()
    ctx := context.Background()
    objects, err := service.List(ctx, "/tmp/folder")
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
}
```

* **Uploading Content**

```go
func main() {
	
    service := afs.New()
    ctx := context.Background()
    keyAuth, err := scp.LocalhostKeyAuth("")
    if err != nil {
        log.Fatal(err)
    }
    err  = service.Init(ctx, "scp://127.0.0.1:22/", keyAuth)
    if err != nil {
        log.Fatal(err)
    }	
    err = service.Upload(ctx, "scp://127.0.0.1:22/folder/asset.txt", 0644, strings.NewReader("test me"))
    if err != nil {
        log.Fatal(err)
    }
    ok, err := service.Exists(ctx, "scp://127.0.0.1:22/folder/asset.txt")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("has file: %v\n", ok)
    _ = service.Delete(ctx, "scp://127.0.0.1:22/folder/asset.txt")
}
```


* **Data Copy**

```go
func main() {

    service := afs.New()
    ctx := context.Background()
    keyAuth, err := scp.LocalhostKeyAuth("")
    if err != nil {
        log.Fatal(err)
    }
    err = service.Copy(ctx, "s3://mybucket/myfolder", "scp://127.0.0.1/tmp", option.NewSource(), option.NewDest(keyAuth))
    if err != nil {
        log.Fatal(err)
    }
}
```

* **Archiving content**

```go

func main() {
	
    secretPath := path.Join(os.Getenv("HOME"), ".secret", "gcp-e2e.json")
    auth, err := gs.NewJwtConfig(option.NewLocation(secretPath))
    if err != nil {
        return
    }
    sourceURL := "mylocalPath/"
    destURL := "gs:mybucket/test.zip/zip://localhost/dir1"
    service := afs.New()
    ctx := context.Background()
    err = service.Copy(ctx, sourceURL, destURL, option.NewDest(auth))
    if err != nil {
        log.Fatal(err)
    }

}	
```


* **Archive Walker**

Walker can be created for tar or zip archive.

```go
func main() {
	
    ctx := context.Background()
	service := afs.New()
	walker := tar.NewWalker(s3afs.New())
	err := service.Copy(ctx, "/tmp/test.tar", "s3:///dest/folder/test", walker)
	if err != nil {
		log.Fatal(err)
	}
```


* **Archive Uploader**

Uploader can be created for tar or zip archive.

```go
func main() {
	
    ctx := context.Background()
	service := afs.New()
	uploader := zip.NewBatchUploader(gsafs.New())
	err := service.Copy(ctx, "gs:///tmp/test/data", "/tmp/data.zip", uploader)
	if err != nil {
		log.Fatal(err)
	}
}
```


* **Data Move**

```go
func main() {
	
    service := afs.New()
	ctx := context.Background()
	keyAuth, err := scp.LocalhostKeyAuth("")
	if err != nil {
		log.Fatal(err)
	}
	err = service.Move(ctx, "/tmp/transient/app", "scp://127.0.0.1/tmp", option.NewSource(), option.NewDest(keyAuth))
	if err != nil {
		log.Fatal(err)
	}
}	
```

* **Batch Upload**

```go
func main() {
	
    service := afs.New()
	ctx := context.Background()
	upload, closer, err := service.Uploader(ctx, "/tmp/clone")
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()
	assets := []*asset.Resource{
		asset.NewFile("asset1.txt", []byte("test 1"), 0644),
		asset.NewFile("asset2.txt", []byte("test 2"), 0644),
		asset.NewDir("folder1", file.DefaultDirOsMode),
		asset.NewFile("folder1/asset1.txt", []byte("test 3"), 0644),
		asset.NewFile("folder1/asset2.txt", []byte("test 4"), 0644),
	}
	for _, asset := range assets {
		relative := ""
		var reader io.Reader
		if strings.Contains(asset.Name, "/") {
			relative, _ = path.Split(asset.Name)
		}
		if ! asset.Dir {
			reader = bytes.NewReader(asset.Data)
		}
		err = upload(ctx, relative, asset.Info(), reader)
		if err != nil {
			log.Fatal(err)
		}
	}
}
```
## Matchers

To filter source content you can use [Matcher](option/matcher.go) option. 
The following have been implemented.


**[Basic Matcher](matcher/basic.go)**

```go
func main() {
	
    matcher, err := NewBasic("/data", ".avro", nil)
    service := afs.New()
    ctx := context.Background()
    err := service.Copy(ctx, "/tmp/data", "s3://mybucket/data/", matcher.Match)
    if err != nil {
        log.Fatal(err)
    }
}
```

**[Filepath matcher](matcher/filepath.go)**

OS style filepath match, with the following terms:
- '*'         matches any sequence of non-Separator characters
- '?'         matches any single non-Separator character
- '[' [ '^' ] { character-range } ']'

```go

func main() {
	
    matcher := matcher.Filepath("*.avro")
    service := afs.New()
    ctx := context.Background()
    err := service.Copy(ctx, "/tmp/data", "gs://mybucket/data/", matcher)
    if err != nil {
        log.Fatal(err)
    }
}	
		

```

**[Ignore Matcher](matcher/ignore.go)**

Ignore matcher represents matcher that matches file that are not in the ignore rules.
The syntax of ignore borrows heavily from that of .gitignore; see https://git-scm.com/docs/gitignore or man gitignore for a full reference.


```go
func mian(){
	
	ignoreMatcher, err := matcher.NewIgnore([]string{"*.txt", ".ignore"})
  	//or matcher.NewIgnore(option.NewLocation(".cloudignore"))
	if err != nil {
		log.Fatal(err)
	}
	service := afs.New()
	ctx := context.Background()
	objects, err := service.List(ctx, "/tmp/folder", ignoreMatcher.Match)
	if err != nil {
		log.Fatal(err)
	}
	for _, object := range objects {
		fmt.Printf("%v %v\n", object.Name(), object.URL())
		if object.IsDir() {
			continue
		}
	}
}	
```



**[Modification Time Matcher](matcher/modification.go)**

Modification Time Matcher represents matcher that matches file that were modified either before or after specified time.

```go
func mian(){
	
	before, err := toolbox.TimeAt("2 days ago in UTC")
    if err != nil {
		log.Fatal(err)
	}	
	modTimeMatcher, err := matcher.NewModification(before, nil)
	if err != nil {
		log.Fatal(err)
	}
	service := afs.New()
	ctx := context.Background()
	objects, err := service.List(ctx, "/tmp/folder", modTimeMatcher.Match)
	if err != nil {
		log.Fatal(err)
	}
	for _, object := range objects {
		fmt.Printf("%v %v\n", object.Name(), object.URL())
		if object.IsDir() {
			continue
		}
	}
}	
```


## Content modifiers

To modify resource content on the fly you can use [Modifier](option/modifier.go) option.

```go
func main() {
	service := afs.New()
	ctx := context.Background()
	sourceURL := "file:/tmp/app.war/zip://localhost/WEB-INF/classes/config.properties"
	destURL := "file:/tmp/app.war/zip://localhost/"
	err := service.Copy(ctx, sourceURL, destURL, modifier.Replace(map[string]string{
		"${changeMe}": os.Getenv("USER"),
	}))
	if err != nil {
		log.Fatal(err)
	}
}
```

```go
package main

import (
	"bytes"
	"context"
	"log"
	"github.com/viant/afs"
	"io"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func modifyContent(info os.FileInfo, reader io.ReadCloser) (closer io.ReadCloser, e error) {
   if strings.HasSuffix(info.Name() ,".info") {
       data, err := ioutil.ReadAll(reader)
       if err != nil {
           return nil, err
       }
       _ = reader.Close()
       expanded := strings.Replace(string(data), "$os.User", os.Getenv("USER"), 1)
       reader = ioutil.NopCloser(strings.NewReader(expanded))
   }
   return reader, nil
}                           

func main() {

    service := afs.New()
    reader ,err := service.DownloadWithURL(context.Background(), "s3://mybucket/meta.info", modifyContent)
    if err != nil {
        log.Fatal(err)	
    }
    
    defer reader.Close()
    content, err := ioutil.ReadAll(reader)
    if err != nil {
        log.Fatal(err)	
    }
    fmt.Printf("content: %s\n", content)
	
}
```


## Options


* **[Page](option/page.go)**

To control number and position of listed resources you can yse page option.

* **[Timeout](option/timeout.go)**

Provider specific timeout.

* **[BasicAuth](option/cred.go)** 

Provides user/password auth.

* **Source & Dest Options**

Groups options by source or destination options. This options work with Copy or Move operations.

```go

func main() {
	
    service := afs.New()
    secretPath :=  path.Join(os.Getenv("HOME"), ".secret","gcp.json")
    jwtConfig, err := gs.NewJwtConfig(option.NewLocation(secretPath))
    if err != nil {
    	log.Fatal(err)
    }
    sourceOptions := option.NewSource(jwtConfig)
    authConfig, err := s3.NewAuthConfig(option.NewLocation("aws.json"))
    if err != nil {
        log.Fatal(err)
    }
    destOptions := option.NewDest(authConfig)
	err = service.Copy(ctx, "gs://mybucket/data", "s3://mybucket/data",  sourceOptions, destOptions)
}

```

Check out [storage manager](#storage-managers) for additional options. 

## Storage Implementations

- [File](file/README.md)
- [In Memory](mem/README.md)
- [SSH - SCP](scp/README.md)
- [HTTP](http/README.md)
- [Tar](tar/README.md)
- [Zip](zip/README.md)
- [GCP - GS](https://github.com/viant/afsc/tree/master/gs)
- [AWS - S3](https://github.com/viant/afsc/tree/master/s3)

## Testing service

To unit test all storage operation all in memory you can use faker service.

In addition you can use error options to test exception handling.

- **DownloadError**
```go

func mian() {
	service := afs.NewFaker()
	ctx := context.Background()
	err := service.Upload(ctx, "gs://myBucket/folder/asset.txt", 0, strings.NewReader("some data"), option.NewUploadError(io.EOF))
	if err != nil {
		log.Fatalf("expect upload error: %v", err)
	}
}
```

- **ReaderError**
```go

func mian() {
    service := afs.NewFaker()
	ctx := context.Background()
	err := service.Upload(ctx, "gs://myBucket/folder/asset.txt", 0, strings.NewReader("some data"), option.NewDownloadError(io.EOF))
	if err != nil {
		log.Fatal(err)
	}
	_, err = service.DownloadWithURL(ctx, "gs://myBucket/folder/asset.txt")
	if err != nil {
		log.Fatalf("expect download error: %v", err)
	}
}
```

- **UploadError**
```go

func mian() {
    service := afs.NewFaker()
    ctx := context.Background()
    err := service.Upload(ctx, "gs://myBucket/folder/asset.txt", 0, strings.NewReader("some data"), option.NewUploadError(io.EOF))
    if err != nil {
        log.Fatalf("expect upload error: %v", err)
    }
}
```


## Test setup utilities

Package [asset](asset) defines basic utilities to quickly manage asset related unit tests.

```go

func Test_XXX(t *testing.T) {

    var useCases = []struct {
		description string
		location    string
		options     []storage.Option
		assets      []*asset.Resource
	}{

	}

	ctx := context.Background()
	for _, useCase := range useCases {
		service := afs.New()
		mgr, err := afs.Manager(useCase.location, useCase.options...)
		if err != nil {
			log.Fatal(err)
		}
		err = asset.Create(mgr, useCase.location, useCase.assets)
		if err != nil {
			log.Fatal(err)
		}
		
		//... actual app logic

		actuals, err := asset.Load(mgr, useCase.location)
		if err != nil {
			log.Fatal(err)
		}
        for _, expect := range useCase.assets {
            actual, ok := actuals[expect.Name]
            if !assert.True(t, ok, useCase.description+": "+expect.Name+fmt.Sprintf(" - actuals: %v", actuals)) {
                continue
            }
            assert.EqualValues(t, expect.Name, actual.Name, useCase.description+" "+expect.Name)
            assert.EqualValues(t, expect.Mode, actual.Mode, useCase.description+" "+expect.Name)
            assert.EqualValues(t, expect.Dir, actual.Dir, useCase.description+" "+expect.Name)
            assert.EqualValues(t, expect.Data, actual.Data, useCase.description+" "+expect.Name)
        }

		_ = asset.Cleanup(mgr, useCase.location)

	}
}

```

## GoCover

[![GoCover](https://gocover.io/github.com/viant/afs)](https://gocover.io/github.com/viant/afs)


<a name="License"></a>
## License

The source code is made available under the terms of the Apache License, Version 2, as stated in the file `LICENSE`.

Individual files may be made available under their own specific license,
all compatible with Apache License, Version 2. Please see individual files for details.

<a name="Credits-and-Acknowledgements"></a>

## Credits and Acknowledgements

**Library Author:** Adrian Witas

