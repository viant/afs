package mem

import (
	"fmt"
	"github.com/viant/afs/file"
	"github.com/viant/afs/object"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	noSuchFileOrDirectoryErrorMessage = "no such file or directory"
)

//Folder represents memory folder
type Folder struct {
	storage.Object
	mutex   *sync.RWMutex
	files   map[string]*File
	folders map[string]*Folder
}

//Objects returns folder objects
func (f *Folder) Objects() []storage.Object {
	var result = make([]storage.Object, 0)
	result = append(result, f.Object)

	f.mutex.Lock()
	defer f.mutex.Unlock()

	for i := range f.folders {
		result = append(result, f.folders[i].Object)
	}
	for i := range f.files {
		result = append(result, f.files[i].Object)
	}
	return result
}

func (f *Folder) putFolder(object storage.Object) error {
	folder := &Folder{}
	if err := object.Unwrap(&folder); err != nil {
		return err
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if _, ok := f.files[folder.Name()]; ok {
		return fmt.Errorf("%v is file", object.URL())
	}
	f.folders[folder.Name()] = folder
	return nil
}

func (f *Folder) putFile(object storage.Object) error {
	objFile := &File{}
	if err := object.Unwrap(&objFile); err != nil {
		return err
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if _, ok := f.folders[objFile.Name()]; ok {
		return fmt.Errorf("%v is directory", object.URL())
	}
	f.files[objFile.Name()] = objFile
	return objFile.uploadError
}

//File returns file or downloadErr
func (f *Folder) file(name string) (*File, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	result, ok := f.files[name]
	if !ok {
		return nil, fmt.Errorf("%v: "+noSuchFileOrDirectoryErrorMessage, url.Join(f.URL(), name))
	}
	return result, nil
}

//File returns file or downloadErr
func (f *Folder) folder(name string) (*Folder, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	result, ok := f.folders[name]
	if !ok {
		keys := make([]string, 0)
		for k := range f.folders {
			keys = append(keys, k)
		}
		return nil, fmt.Errorf("%v: "+noSuchFileOrDirectoryErrorMessage+", %v%v", url.Join(f.URL(), name), f.Name(), keys)
	}
	return result, nil
}

//File returns a file for supplied location
func (f *Folder) File(URL string) (*File, error) {
	object, err := f.Lookup(URL, 0)
	if err != nil {
		return nil, err
	}
	if object.IsDir() {
		return nil, fmt.Errorf("%v is directory", URL)
	}
	result := &File{}
	return result, object.Unwrap(&result)
}

//Folder returns a folder for supplied URL, when mkdirMode is non zero it will create missing folders with supplied mode
func (f *Folder) Folder(URL string, mkdirMode os.FileMode) (*Folder, error) {
	object, err := f.Lookup(URL, mkdirMode)
	if err != nil {
		return nil, err
	}
	result := &Folder{}
	return result, object.Unwrap(&result)
}

//Lookup lookup path, when mkdirMode is non zero it will create missing folders with supplied mode
func (f *Folder) Lookup(URL string, mkdirMode os.FileMode) (storage.Object, error) {
	_, URLPath := url.Base(URL, Scheme)

	if URLPath == "" {
		return f.Object, nil
	}
	var elements = SplitPath(URLPath)
	if len(elements) == 0 {
		return f.Object, nil
	}
	isLast := len(elements) == 1
	if isLast {
		if matched, err := f.file(elements[0]); err == nil {
			return matched.Object, nil
		}
	}
	childName := elements[0]
	child, err := f.folder(elements[0])
	if err != nil {
		if mkdirMode == 0 {
			return nil, err
		}
		//Error f.URL get missing first /
		childURL := url.Join(f.URL(), childName)
		child = NewFolder(childURL, mkdirMode)
		if err = f.putFolder(child); err != nil {
			return nil, err
		}
	}
	if isLast {
		return child.Object, nil
	}
	subPath := strings.Join(elements[1:], "/")
	return child.Lookup(subPath, mkdirMode)
}

//Delete deletes object or return error
func (f *Folder) Delete(name string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	_, hasFolder := f.folders[name]
	if hasFolder {
		delete(f.folders, name)
	}
	_, hasFile := f.files[name]
	if hasFile {
		delete(f.files, name)
	}
	if hasFile || hasFolder {
		return nil
	}
	return fmt.Errorf("%v: no such file or directory", url.Join(f.URL(), name))
}

//Put adds object to this folder
func (f *Folder) Put(object storage.Object) error {
	if object.IsDir() {
		return f.putFolder(object)
	}
	return f.putFile(object)
}

//NewFolder returns a folder for supplied URL
func NewFolder(URL string, mode os.FileMode) *Folder {
	baseURL, URLPath := Split(URL)
	URL = url.Join(baseURL, URLPath)
	name := URLPath
	if strings.Count(URLPath, "/") > 0 {
		_, name = path.Split(URLPath)
	}
	info := file.NewInfo(name, 0, mode, time.Now(), true)
	folder := &Folder{
		mutex:   &sync.RWMutex{},
		files:   make(map[string]*File),
		folders: make(map[string]*Folder),
	}
	folder.Object = object.New(URL, info, folder)
	return folder
}
