package option

import (
	"fmt"
	"github.com/viant/afs/storage"
	"reflect"
)

//Mode assign mode
type FilterMode struct {
	Strict bool
}

//Assign assign supplied option or throw error is unsupported
func Assign(options []storage.Option, supported ...interface{}) ([]storage.Option, error) {
	mode := &FilterMode{}
	_, _ = assign(options, false, append(supported, &mode))
	return assign(options, mode.Strict, supported)
}

//Assign assign supplied option or throw error is unsupported
func assign(options []storage.Option, strictMode bool, supported []interface{}) ([]storage.Option, error) {
	var unfiltered = make([]storage.Option, 0)
	if len(options) == 0 {
		return options, nil
	}
	if len(supported) == 0 {
		if !strictMode {
			return unfiltered, nil
		}
		return nil, fmt.Errorf("unsupported option %T", options[0])
	}
	var index = make(map[reflect.Type]interface{})
	for i := range supported {
		index[reflect.TypeOf(supported[i]).Elem()] = supported[i]
	}
	var err error
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
			if strictMode {
				err = fmt.Errorf("unsupported option %T", options[i])
			} else {
				unfiltered = append(unfiltered, options[i])
			}
			continue
		}
		reflect.ValueOf(target).Elem().Set(optionValue)
	}
	return unfiltered, err
}
