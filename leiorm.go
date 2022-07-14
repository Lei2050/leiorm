package leiorm

import (
	"reflect"
)

type saveType int

const (
	SaveTypeBase saveType = iota
	SaveTypeBaseArray
	SaveTypeBaseMap
	SaveTypeStruct
	SaveTypeNotSupported
)

const (
	PrimaryKeyTag = "leiormpri"
	FieldSaveTag  = "leiorm"
)

var (
	arrJoinLevel1 = ";"
	arrJoinLevel2 = "-"
	mapJoinLevel1 = "|"
	mapJoinLevel2 = ":"
)

// isBaseArray judges whether t is a array of base type.
func isBaseArray(t reflect.Type) bool {
	k := t.Kind()
	if k != reflect.Array && k != reflect.Slice {
		return false
	}
	t = t.Elem()
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String, reflect.Bool:
		return true
	}
	return false
}

// getSaveType decides whether t is a currently supported type.
func getSaveType(t reflect.Type) saveType {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String, reflect.Bool:
		return SaveTypeBase
	case reflect.Array, reflect.Slice:
		t = t.Elem()
		switch t.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.String, reflect.Bool:
			return SaveTypeBaseArray
		case reflect.Array, reflect.Slice:
			switch t.Elem().Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Float32, reflect.Float64, reflect.String, reflect.Bool:
				//only support 2d array of base type
				return SaveTypeBaseArray
			}
		case reflect.Ptr:
			if t.Elem().Kind() == reflect.Struct {
				//Pointer in a array can only point to structures
				return SaveTypeStruct
			}
		}
		return SaveTypeNotSupported
	case reflect.Ptr:
		t = t.Elem()
		if t.Kind() == reflect.Struct {
			//Pointer can only point to structures
			return SaveTypeStruct
		}
		//multi-level pointer is not allowed
		return SaveTypeNotSupported
	case reflect.Map:
		switch t.Key().Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.String, reflect.Bool:
		default:
			//the key of a map must be a base type.
			return SaveTypeNotSupported
		}

		et := t.Elem()
		switch et.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.String, reflect.Bool:
			return SaveTypeBaseMap
		case reflect.Ptr:
			t = et.Elem()
			if t.Kind() == reflect.Struct {
				//Pointer can only point to structures
				return SaveTypeStruct
			}
			//multi-level pointer is not allowed
			return SaveTypeNotSupported
		case reflect.Array, reflect.Slice:
			//support the type kind of map[int][]float
			if isBaseArray(et) {
				return SaveTypeBaseMap
			}
			return SaveTypeNotSupported
		default:
			//the key of a map must be a base type.
			return SaveTypeNotSupported
		}

	case reflect.Struct:
		return SaveTypeStruct
	}
	return SaveTypeNotSupported
}

func findFieldByTag(v reflect.Value, tag string) (int, reflect.StructField, reflect.Value) {
	typ := v.Type()
	num := v.NumField()
	for i := 0; i < num; i++ {
		tfield := typ.Field(i)
		if _, ok := tfield.Tag.Lookup(tag); ok {
			return i, tfield, v.Field(i)
		}
	}
	return -1, reflect.StructField{}, reflect.Value{}
}
