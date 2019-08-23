package option

import (
	"github.com/viant/afs/storage"
	"reflect"
)

//Assign assign supplied option, if returns un assign options and true if assign at least one
func Assign(options []storage.Option, supported ...interface{}) ([]storage.Option, bool) {
	return assign(options, supported)
}

//Assign assign supplied option
func assign(options []storage.Option, supported []interface{}) ([]storage.Option, bool) {
	var unfiltered = make([]storage.Option, 0)
	if len(options) == 0 {
		return options, false
	}
	if len(supported) == 0 {
		return options, false
	}

	var index = make(map[reflect.Type]interface{})
	for i := range supported {
		index[reflect.TypeOf(supported[i]).Elem()] = supported[i]
	}
	assigned := false
	for i := range options {
		option := options[i]
		if option == nil {
			continue
		}
		optionValue := reflect.ValueOf(option)
		target, ok := index[optionValue.Type()]
		if !ok {
			for k, v := range index {
				if optionValue.Type().AssignableTo(k) {
					target = v
					ok = true
					break
				}
			}
		}
		if !ok {
			unfiltered = append(unfiltered, options[i])
			continue
		}
		assigned = true
		reflect.ValueOf(target).Elem().Set(optionValue)
	}
	return unfiltered, assigned
}
