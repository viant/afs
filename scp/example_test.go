package scp_test

import (
	"context"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/afs/option"
	"github.com/viant/afs/scp"
	"io/ioutil"
	"log"
	"strings"
)

//Example_Storager storager usage example (uses files rather then URLs)
func Example_Storager() {

	//make sure that ~/.ssh/authorized_keys is configured
	auth, err := scp.LocalhostKeyAuth("")
	if err != nil {
		log.Fatal(err)
	}
	provider := scp.NewAuthProvider(auth, nil)
	config, err := provider.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}
	service, err := scp.NewStorager("127.0.0.1:22", 15000, config)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	location := "/tmp/myfile"
	err = service.Upload(ctx, location, 0644, strings.NewReader("somedata"))
	if err != nil {
		log.Fatal(err)
	}
	reader, err := service.Open(ctx, location)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("data: %s\n", data)

	has, _ := service.Exists(ctx, location)
	fmt.Printf("%v %v", location, has)

	files, err := service.List(ctx, location, 0, 3)
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

//Example_Service storager usage example
func Example_Service() {
	auth, err := scp.LocalhostKeyAuth("")
	if err != nil {
		log.Fatal(err)
	}
	service := afs.New()
	ctx := context.Background()
	reader, err := service.OpenURL(ctx, "scp://127.0.0.1:22/etc/hosts", auth, option.NewTimeout(2000))
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("data: %s\n", data)
}

//Example_BasicAuth storager with basic auth example
func Example_LocalhostKeyAuth() {
	auth := option.NewBasicAuth("myuser", "nypass")
	service := afs.New()
	ctx := context.Background()
	reader, err := service.OpenURL(ctx, "scp://127.0.0.1:22/etc/hosts", auth, option.NewTimeout(2000))
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("data: %s\n", data)
}

//Example_BasicAuth storager with basic auth example
func Example_BasicAuth() {
	auth, err := scp.LocalhostKeyAuth("pathToMyAuthorizedKeyContainer_id_rsa")
	if err != nil {
		log.Fatal(err)
	}
	service := afs.New()
	ctx := context.Background()
	reader, err := service.OpenURL(ctx, "scp://127.0.0.1:22/etc/hosts", auth, option.NewTimeout(2000))
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("data: %s\n", data)
}
