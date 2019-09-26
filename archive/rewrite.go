package archive

import (
	"context"
	"fmt"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/file"
	"github.com/viant/afs/storage"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

//Modifier represents rewrite modifier
type Modifier func(resources []*asset.Resource) ([]*asset.Resource, error)

//Rewrite rewrites content
func Rewrite(ctx context.Context, walker storage.Walker, URL string, upload storage.Upload, handler Modifier) error {
	var resources = make([]*asset.Resource, 0)
	err := walker.Walk(ctx, URL, func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
		name := path.Join(parent, info.Name())
		var data []byte
		if !info.IsDir() {
			if data, err = ioutil.ReadAll(reader); err != nil {
				return false, err
			}
		}
		resource := asset.New(name, info.Mode(), info.IsDir(), "", data)
		resources = append(resources, resource)
		return true, nil
	})
	if err != nil {
		return err
	}
	resources, err = handler(resources)
	if err != nil {
		return err
	}
	resources = dedupe(resources)
	for i := range resources {
		resource := resources[i]
		parent, _ := path.Split(resource.Name)
		if err = upload(ctx, parent, resource.Info(), resource.Reader()); err != nil {
			return err
		}
	}
	return err
}

func dedupe(resources []*asset.Resource) []*asset.Resource {
	var result = make([]*asset.Resource, 0)
	var dedupe = make(map[string]bool)
	for i := range resources {
		if dedupe[resources[i].Name] {
			continue
		}
		dedupe[resources[i].Name] = true
		result = append(result, resources[i])
	}
	return result
}

//DeleteHandler represents on rewrite upload delete handler
func DeleteHandler(location string) func(resources []*asset.Resource) ([]*asset.Resource, error) {
	return func(resources []*asset.Resource) ([]*asset.Resource, error) {
		var filtered = make([]*asset.Resource, 0)
		deleted := false

		for i := range resources {
			if resources[i].Name == location || strings.HasPrefix(resources[i].Name+"/", location) {
				deleted = true
				continue
			}
			filtered = append(filtered, resources[i])
		}
		var err error
		if !deleted {
			err = fmt.Errorf("%v: not found", location)
		}
		return filtered, err
	}
}

func addResource(existing map[string]*asset.Resource, resource *asset.Resource, resources []*asset.Resource) ([]*asset.Resource, error) {
	location := resource.Name
	parent, _ := path.Split(location)
	var additions = make([]*asset.Resource, 0)
	parentDepth := strings.Count(parent, "/")
	additions = append(additions, resource)
	for i := 0; i < parentDepth; i++ {
		parent = strings.Trim(parent, "/")
		if _, ok := existing[parent]; ok || parent == "" {
			break
		}
		dir := asset.New(parent, file.DefaultDirOsMode, true, "", nil)
		existing[parent] = dir
		additions = append(additions, dir)
		parent, _ = path.Split(parent)
	}
	for i := 0; i < len(additions)/2; i++ {
		tmp := additions[i]
		swapIndex := len(additions) - 1 - i
		additions[i] = additions[swapIndex]
		additions[swapIndex] = tmp
	}
	if parent == "" {
		resources = append(resources, additions...)
		return resources, nil
	}

	for i, resource := range resources {
		if resource.Name == parent {
			additions = append(additions, resources[i+1:]...)
			resources = append(resources[:i+1], additions...)
			return resources, nil
		}
	}
	return nil, fmt.Errorf("unable merge parent:%v, loc %v", parent, location)
}

//CreateHandler represents on rewrite upload create handler
func CreateHandler(location string, mode os.FileMode, data []byte, isDir bool) func(resources []*asset.Resource) ([]*asset.Resource, error) {
	return func(resources []*asset.Resource) ([]*asset.Resource, error) {
		newResource := asset.New(location, mode, isDir, "", data)
		var dirs = make(map[string]*asset.Resource)
		for i, resource := range resources {
			if resources[i].Name == location {
				err := resources[i].MergeFrom(newResource)
				return resources, err
			}
			if resource.Dir {
				dirs[resource.Name] = resources[i]
			}
		}
		resource := asset.New(location, mode, isDir, "", data)
		return addResource(dirs, resource, resources)

	}
}

//UploadHandler represents on rewrite upload create handler
func UploadHandler(toUpload []*asset.Resource) func(resources []*asset.Resource) ([]*asset.Resource, error) {
	return func(resources []*asset.Resource) ([]*asset.Resource, error) {
		var existing = make(map[string]*asset.Resource)
		for i, resource := range resources {
			existing[strings.Trim(resource.Name, "/")] = resources[i]
		}
		var err error
		for _, resource := range toUpload {
			if existing, ok := existing[resource.Name]; ok {
				err = existing.MergeFrom(resource)
				if err != nil {
					return nil, err
				}
				continue
			}
			if resources, err = addResource(existing, resource, resources); err != nil {
				return nil, err
			}
		}
		return resources, nil
	}
}

//UpdateDestination updates resource with specified destination
func UpdateDestination(destination string, resources []*asset.Resource) []*asset.Resource {
	if strings.Trim(destination, "/") == "" {
		return resources
	}
	var result = make([]*asset.Resource, len(resources))
	for i := range resources {
		resource := *resources[i]
		resource.Name = path.Join(destination, resources[i].Name)
		result[i] = &resource
	}
	return result
}
