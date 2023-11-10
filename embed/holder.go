package embed

import (
	"crypto/sha256"
	"embed"
	"github.com/viant/xunsafe"
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
	r.Append(aFile)
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
		prev := r.Slice.ValueAt(xunsafe.AsPointer(r.slicePtr), i)
		next := r.Slice.ValueAt(xunsafe.AsPointer(r.slicePtr), j)
		prevName := fileNameField.String(xunsafe.AsPointer(prev))
		nextName := fileNameField.String(xunsafe.AsPointer(next))
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
	r.Append(aParent)
	if strings.Count(parent, "/") > 1 {
		r.ensureParent(parent)
	}
}
