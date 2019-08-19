package asset_test

import (
	"context"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/storage"
	"log"
)

//Example_Create crete test assets example
func Example_Create() {

	var useCases = []struct {
		description string
		location    string
		options     []storage.Option
		assets      []*asset.Resource
	}{}

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
		_, err = service.Exists(ctx, useCase.location)
		if err != nil {
			log.Fatal(err)
		}

		actuals, err := asset.Load(mgr, useCase.location)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("actuals: %v\n", actuals)
		_ = asset.Cleanup(mgr, useCase.location)

	}
}
