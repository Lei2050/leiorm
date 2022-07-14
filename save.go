package leiorm

import (
	"bytes"
	"fmt"
	"reflect"
)

func arrayToStr(v reflect.Value, args ...interface{}) string {
	var strbuffer bytes.Buffer

	joinChar1 := arrJoinLevel1
	if len(args) >= 1 {
		joinChar1 = args[0].(string)
	}

	joinChar2 := arrJoinLevel2
	if len(args) >= 2 {
		joinChar2 = args[1].(string)
	}

	elemKind := v.Type().Elem().Kind()
	arr2d := elemKind == reflect.Array || elemKind == reflect.Slice

	l := v.Len()
	for i := 0; i < l; i++ {
		if i != 0 {
			strbuffer.WriteString(joinChar1)
		}
		val := v.Index(i)
		if arr2d {
			strbuffer.WriteString(arrayToStr(val, joinChar2))
		} else {
			strbuffer.WriteString(fmt.Sprintf("%v", val.Interface()))
		}
	}
	return strbuffer.String()
}

func mapToStr(v reflect.Value) string {
	var strbuffer bytes.Buffer
	iter := v.MapRange()
	arr2d := isBaseArray(v.Type().Elem())
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()
		if arr2d {
			strbuffer.WriteString(fmt.Sprintf("%v%s%s%s",
				key.Interface(), mapJoinLevel2, arrayToStr(val), mapJoinLevel1))
		} else {
			strbuffer.WriteString(fmt.Sprintf("%v%s%v%s",
				key.Interface(), mapJoinLevel2, val.Interface(), mapJoinLevel1))
		}
	}
	if strbuffer.Len() <= 0 {
		return ""
	}
	return string(strbuffer.Bytes()[:strbuffer.Len()-1])
}

func saveProcessArray(v reflect.Value, key string) (rcs RedisCommands) {
	kkey := key + "s"

	rdArgs := RedisArgs{kkey}
	var autoId int
	l := v.Len()
	for i := 0; i < l; i++ {
		value := v.Index(i)
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		_, _, subIdField := findFieldByTag(value, PrimaryKeyTag)
		var tkey interface{}
		if subIdField.IsValid() {
			tkey = subIdField.Interface()
		} else {
			autoId++
			tkey = autoId
		}
		rdArgs = rdArgs.Add(tkey)

		rcs = rcs.Add(saveProcessStruct(value, fmt.Sprintf("%s:%v", key, tkey))...)
	}

	rcs = rcs.Add(&RedisCommand{"DEL", RedisArgs{kkey}})
	if len(rdArgs) > 1 {
		rcs = rcs.Add(&RedisCommand{"SADD", rdArgs})
	}

	return
}

func saveProcessMap(v reflect.Value, key string) (rcs RedisCommands) {
	kkey := key + "s"

	rdArgs := RedisArgs{kkey}
	iter := v.MapRange()
	for iter.Next() {
		ikey := iter.Key()
		ival := iter.Value()
		if ival.Kind() == reflect.Ptr {
			ival = ival.Elem()
		}

		rdArgs = rdArgs.Add(ikey.Interface())

		rcs = rcs.Add(saveProcessStruct(ival, fmt.Sprintf("%s:%v", key, ikey.Interface()))...)
	}

	rcs = rcs.Add(&RedisCommand{"DEL", RedisArgs{kkey}})
	if len(rdArgs) > 1 {
		rcs = rcs.Add(&RedisCommand{"SADD", rdArgs})
	}

	return rcs
}

func saveProcessStruct(v reflect.Value, key string) (rcs RedisCommands) {
	//typ := reflect.TypeOf(v)
	//value := reflect.ValueOf(v)
	typ := v.Type()
	value := v
	if value.Type().Kind() == reflect.Ptr {
		value = value.Elem()
	}

	hmsetArgs := RedisArgs{key}
	num := value.NumField()
	for i := 0; i < num; i++ {
		tfield := typ.Field(i)
		vfield := value.Field(i)
		var fSaveTag string
		var ok bool
		if fSaveTag, ok = tfield.Tag.Lookup("leiorm"); !ok {
			if fSaveTag, ok = tfield.Tag.Lookup("leiormpri"); !ok {
				//both `leirom` and `leiormpri` are not found,
				//and the field will not be saved.
				continue
			}
		}

		switch tfield.Type.Kind() {
		case reflect.Bool:
			if vfield.Bool() {
				hmsetArgs = hmsetArgs.Add(fSaveTag, int(1))
			} else {
				hmsetArgs = hmsetArgs.Add(fSaveTag, int(0))
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.String:
			hmsetArgs = hmsetArgs.Add(fSaveTag, vfield.Interface())
		case reflect.Array, reflect.Slice:
			switch getSaveType(tfield.Type.Elem()) {
			case SaveTypeBase, SaveTypeBaseArray:
				hmsetArgs = hmsetArgs.Add(fSaveTag, arrayToStr(vfield))
			case SaveTypeStruct:
				rcs = rcs.Add(saveProcessArray(vfield, key+":"+fSaveTag)...)
			default:
				continue
			}
		case reflect.Map:
			switch getSaveType(tfield.Type) {
			case SaveTypeBaseMap:
				hmsetArgs = hmsetArgs.Add(fSaveTag, mapToStr(vfield))
			case SaveTypeStruct:
				rcs = rcs.Add(saveProcessMap(vfield, key+":"+fSaveTag)...)
			default:
				continue
			}
		case reflect.Ptr:
			if getSaveType(tfield.Type) != SaveTypeStruct {
				continue
			}
			vfield = vfield.Elem()
			fallthrough
		case reflect.Struct:
			rcs = rcs.Add(saveProcessStruct(vfield, key+":"+fSaveTag)...)
			//_, _, subIdField := findFieldByTag(vfield, PrimaryKeyTag)
			//if subIdField.IsValid() {
			//	hmsetArgs = hmsetArgs.Add(fSaveTag, subIdField.Interface())
			//	rcs = rcs.Add(saveProcessStruct(vfield, fmt.Sprintf("%s:%s:%v", key, fSaveTag, subIdField.Interface()))...)
			//} else {
			//	hmsetArgs = hmsetArgs.Add(fSaveTag, key)
			//	rcs = rcs.Add(saveProcessStruct(vfield, key+":"+fSaveTag)...)
			//}
		default: //其他类型不予支持
			continue
		}
	}

	if len(hmsetArgs) > 1 {
		rcs = rcs.Add(&RedisCommand{"HMSET", hmsetArgs})
	}

	return rcs
}

func saveProcessBool(v bool, key interface{}) (rcs RedisCommands) {
	rdkey := fmt.Sprintf("%v", key)
	var rdval int
	if v {
		rdval = 1
	}
	rcs = rcs.Add(&RedisCommand{"SET", RedisArgs{rdkey, rdval}})
	return rcs
}

func saveProcessSimple(v interface{}, key interface{}) (rcs RedisCommands) {
	rdkey := fmt.Sprintf("%v", key)
	rcs = rcs.Add(&RedisCommand{"SET", RedisArgs{rdkey, v}})
	return rcs
}

/*
	Bool Int Int8 Int16 Int32 Int64 Uint Uint8 Uint16 Uint32 Uint64
	Uintptr Float32 Float64 Complex64 Complex128 Array Chan Func
	Interface Map Ptr Slice String Struct UnsafePointer
*/
func SaveModel(rd RedisClienter, ider interface{}, skey interface{}) error {
	if ider == nil {
		return fmt.Errorf("model is nil")
	}
	var key string
	if skey != nil {
		key = fmt.Sprintf("%v", skey)
	}

	value := reflect.ValueOf(ider)
	for value.Type().Kind() == reflect.Ptr {
		value = value.Elem()
	}

	var rcs RedisCommands

	switch value.Kind() {
	case reflect.Bool:
		if len(key) == 0 {
			return fmt.Errorf("must specify a key for saving value of bool type")
		}
		rcs = saveProcessBool(value.Bool(), key)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		if len(key) == 0 {
			return fmt.Errorf("must specify a key for saving value of number type or string")
		}
		rcs = saveProcessSimple(value.Interface(), key)
	case reflect.Struct:
		if len(key) == 0 {
			key = value.Type().Name()
			_, _, subIdField := findFieldByTag(value, PrimaryKeyTag)
			if subIdField.IsValid() {
				key += fmt.Sprintf(":%v", subIdField.Interface())
			}
		}
		rcs = saveProcessStruct(value, key)
	case reflect.Array, reflect.Slice:
		if len(key) == 0 {
			return fmt.Errorf("must specify a key for saving value of array or slice")
		}
		switch getSaveType(value.Type().Elem()) {
		case SaveTypeBase, SaveTypeBaseArray:
			rcs = saveProcessSimple(arrayToStr(value), key)
		case SaveTypeStruct:
			rcs = saveProcessArray(value, key)
		}
	case reflect.Map:
		if len(key) == 0 {
			return fmt.Errorf("must specify a key for saving value of map type")
		}
		switch getSaveType(value.Type()) {
		case SaveTypeBaseMap:
			rcs = saveProcessSimple(mapToStr(value), key)
		case SaveTypeStruct:
			rcs = saveProcessMap(value, key)
		default:
			return fmt.Errorf("unsupported type")
		}

	////
	//case reflect.Ptr: // Already a reflect.Ptr

	//case reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Chan:
	//case reflect.Interface:
	//case reflect.UnsafePointer:

	default:
		return fmt.Errorf("unsupported type")
	}

	for _, rc := range rcs {
		rd.Send(rc.Cmd, rc.Args...)
		//rd.Do(rc.Cmd, rc.Args...)
	}
	rd.Flush()

	return nil
}
