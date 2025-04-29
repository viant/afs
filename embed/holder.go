package embed

import (
	"context"
	"crypto/sha256"
	"embed"
	"github.com/viant/afs/url"
	"github.com/viant/xunsafe"
	"io"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

var filesField = xunsafe.FieldByIndex(reflect.TypeOf(fs{}), 0)
var aFileType = filesField.Type.Elem().Elem()
var fileNameField = xunsafe.FieldByName(aFileType, "name")
var fileDataField = xunsafe.FieldByName(aFileType, "data")
var fileHashField = xunsafe.FieldByName(aFileType, "hash")

type fs embed.FS

// Holder represents embed.fs holder
type Holder struct {
	fs
	*xunsafe.Slice
	*xunsafe.Appender
	refSlicePtr reflect.Value
	slicePtr    interface{}
	entries     map[string]string
	needSorting bool
}

// AddFs adds file to embed.FS
func (r *Holder) AddFs(fs *embed.FS, path string) {
	mgr := newManager(fs, path)
	r.append(mgr, path, fs)
}

func (r *Holder) append(mgr *manager, URL string, embedFs *embed.FS) {
	objects, _ := mgr.List(context.Background(), URL, embedFs)
	parent := strings.Trim(url.Path(URL), "/")
	if len(objects) > 0 {
		for _, object := range objects {
			if object.IsDir() {
				if url.IsSchemeEquals(URL, object.URL()) {
					continue
				}
				r.append(mgr, object.URL(), embedFs)
				continue
			}
			reader, err := mgr.OpenURL(context.Background(), object.URL(), embedFs)
			if err != nil {
				continue
			}
			data, err := io.ReadAll(reader)
			r.Add(path.Join(parent, object.Name()), string(data))
			_ = reader.Close()
		}
	}
}

// NewHolder create a fs holder
func NewHolder() *Holder {
	reflSlice := reflect.MakeSlice(filesField.Type.Elem(), 0, 0)
	reflSlicePtr := reflect.New(filesField.Type.Elem())
	reflSlicePtr.Elem().Set(reflSlice)
	ret := &Holder{entries: map[string]string{}}
	ret.Slice = xunsafe.NewSlice(filesField.Type.Elem())
	slicePtr := xunsafe.AsPointer(reflSlicePtr.Interface())
	ret.refSlicePtr = reflSlicePtr
	ret.slicePtr = reflSlicePtr.Interface()
	ret.Appender = ret.Slice.Appender(slicePtr)
	ret.syncValues()
	return ret
}

func (r *Holder) syncValues() {
	holderPtr := unsafe.Pointer(&r.fs)
	filesField.SetValue(holderPtr, r.slicePtr)
}

// EmbedFs returns *embed.FS
func (r *Holder) EmbedFs() *embed.FS {
	r.sortFiles()
	ret := embed.FS(r.fs)
	return &ret
}

// Add adds file to embed.FS
func (r *Holder) Add(name string, data string) {
	r.ensureParent(name)
	aFile := r.newFile(name, data)
	r.Appender.Append(aFile)
	r.syncValues()
	r.needSorting = true

}

func (r *Holder) sortFiles() {
	if !r.needSorting {
		return
	}
	r.needSorting = false
	aSlice := r.refSlicePtr.Elem()
	sort.Slice(aSlice.Interface(), func(i, j int) bool {
		// grab struct values
		prev := r.Slice.ValueAt(xunsafe.AsPointer(r.slicePtr), i)
		next := r.Slice.ValueAt(xunsafe.AsPointer(r.slicePtr), j)
		prevName := fileNameField.String(xunsafe.AsPointer(prev))
		nextName := fileNameField.String(xunsafe.AsPointer(next))

		trim := func(s string) string { return strings.TrimSuffix(s, "/") }

		// depth == number of “/” after trimming a trailing “/”
		prevDepth := strings.Count(trim(prevName), "/")
		nextDepth := strings.Count(trim(nextName), "/")

		prevIsDir := strings.HasSuffix(prevName, "/")
		nextIsDir := strings.HasSuffix(nextName, "/")

		//  shallower (root-level) path wins
		if prevDepth != nextDepth {
			return prevDepth < nextDepth
		}
		//  at same depth, directory wins over file
		if prevIsDir != nextIsDir {
			return prevIsDir
		}
		//  finally, lexicographic order
		return prevName < nextName
	})

}

func (r *Holder) newFile(name string, data string) interface{} {
	aFile := reflect.New(aFileType).Interface()
	aFilePtr := xunsafe.AsPointer(aFile)
	fileNameField.Set(aFilePtr, name)
	fileDataField.Set(aFilePtr, data)
	value := fileHashField.Addr(aFilePtr)
	if hash, ok := value.(*[16]uint8); ok {
		h := sha256.New()
		h.Write([]byte(data))
		bs := h.Sum(nil)
		for i := range *hash {
			(*hash)[i] = bs[i]
		}
	}

	return aFile
}

func (r *Holder) ensureParent(name string) {
	parent, _ := filepath.Split(name)
	if parent == "" {
		return
	}
	if _, ok := r.entries[parent]; ok {
		return
	}
	r.entries[parent] = ""
	aParent := r.newFile(parent, "")
	r.Appender.Append(aParent)
	if strings.Count(parent, "/") > 1 {
		r.ensureParent(parent)
	}
}
