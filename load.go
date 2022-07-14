package leiorm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

func cannotConvert(d reflect.Value, s interface{}) error {
	var sname string
	switch s.(type) {
	case string:
		sname = "Redis simple string"
	case redis.Error:
		sname = "Redis error"
	case int64:
		sname = "Redis integer"
	case []byte:
		sname = "Redis bulk string"
	case []interface{}:
		sname = "Redis array"
	case nil:
		sname = "Redis nil"
	default:
		sname = reflect.TypeOf(s).String()
	}
	return fmt.Errorf("cannot convert from %s to %s", sname, d.Type())
}

func convertAssignBool(d reflect.Value, s interface{}) bool {
	switch s := s.(type) {
	case []byte:
		d.SetBool(len(s) > 0)
	case int64:
		d.SetInt(s)
	case string:
		d.SetBool(len(s) > 0)
	case nil:
		d.SetBool(false)
	default:
		err := cannotConvert(d, s)
		fmt.Println(err)
		return false
	}
	return true
}

func convertAssignInt(d reflect.Value, s interface{}) bool {
	var sv int64
	switch s := s.(type) {
	case nil:
		d.SetInt(0)
	case []byte:
		if i, e := strconv.ParseInt(string(s), 10, 64); e != nil {
			fmt.Printf("warning can't to int:%v\n", s)
			d.SetInt(0)
		} else {
			d.SetInt(i)
			sv = i
		}
	case int64:
		d.SetInt(s)
		sv = s
	case string:
		if i, e := strconv.ParseInt(s, 10, 64); e != nil {
			fmt.Printf("warning can't to int:%v\n", s)
			d.SetInt(0)
		} else {
			d.SetInt(i)
			sv = i
		}
	default:
		err := cannotConvert(d, s)
		fmt.Println(err)
		return false
	}

	if d.Int() != sv {
		fmt.Println("Warning: error range") //strconv.ErrRange
	}

	return true
}

func convertAssignUint(d reflect.Value, s interface{}) bool {
	var sv uint64
	switch s := s.(type) {
	case nil:
		d.SetUint(0)
	case []byte:
		if i, e := strconv.ParseUint(string(s), 10, 64); e != nil {
			fmt.Printf("warning can't to uint:%v\n", s)
			d.SetUint(0)
		} else {
			d.SetUint(i)
			sv = i
		}
	case int64:
		sv = uint64(s)
		d.SetUint(sv)
	case string:
		if i, e := strconv.ParseUint(s, 10, 64); e != nil {
			fmt.Printf("warning can't to uint:%v\n", s)
			d.SetUint(0)
		} else {
			d.SetUint(i)
			sv = i
		}
	default:
		err := cannotConvert(d, s)
		fmt.Println(err)
		return false
	}

	if d.Uint() != sv {
		fmt.Println("Warning: error range") //strconv.ErrRange
	}

	return true
}

func convertAssignFloat(d reflect.Value, s interface{}) bool {
	var sv float64
	switch s := s.(type) {
	case nil:
		d.SetFloat(0)
	case []byte:
		if i, e := strconv.ParseFloat(string(s), 64); e != nil {
			fmt.Printf("warning can't to float:%v\n", s)
			d.SetFloat(0)
		} else {
			d.SetFloat(i)
			sv = i
		}
	case int64:
		sv = float64(s)
		d.SetFloat(sv)
	case string:
		if i, e := strconv.ParseFloat(s, 64); e != nil {
			fmt.Printf("warning can't to float:%v\n", s)
			d.SetFloat(0)
		} else {
			d.SetFloat(i)
			sv = i
		}
	default:
		err := cannotConvert(d, s)
		fmt.Println(err)
		return false
	}

	if d.Float() != sv {
		fmt.Println("Warning: error range") //strconv.ErrRange
	}

	return true
}

func convertAssignArray(d reflect.Value, s interface{}, sep string) bool {
	str, ok := s.(string)
	if !ok {
		fmt.Printf("convertAssignArray can't convert to string:%v\n", s)
		return false
	}

	elemType := d.Type().Elem()
	strs := strings.Split(str, sep)
	if d.Len() != len(strs) {
		fmt.Println("warning: convertAssignArray len didn't match")
	}

	if elemType.Kind() == reflect.Array {
		for i := 0; i < d.Len() && i < len(strs); i++ {
			elem := reflect.New(elemType)
			convertAssignArray(elem.Elem(), strs[i], arrJoinLevel2)
			d.Index(i).Set(elem.Elem())
		}
	} else if elemType.Kind() == reflect.Slice {
		for i := 0; i < d.Len() && i < len(strs); i++ {
			elem := reflect.New(elemType)
			convertAssignSlice(elem.Elem(), strs[i], arrJoinLevel2)
			d.Index(i).Set(elem.Elem())
		}
	} else {
		for i := 0; i < d.Len() && i < len(strs); i++ {
			d.Index(i).Set(fromString(reflect.New(d.Type().Elem()).Elem(), strs[i]))
		}
	}

	return true
}

func convertAssignSlice(d reflect.Value, s interface{}, sep string) bool {
	str, ok := s.(string)
	if !ok {
		fmt.Printf("convertAssignSlice can't convert to string:%v\n", s)
		return false
	}

	elemType := d.Type().Elem()
	strs := strings.Split(str, sep)
	slc := d
	if elemType.Kind() == reflect.Array {
		for _, v := range strs {
			elem := reflect.New(elemType)
			convertAssignArray(elem.Elem(), v, arrJoinLevel2)
			slc = reflect.Append(slc, elem.Elem())
		}
	} else if elemType.Kind() == reflect.Slice {
		for _, v := range strs {
			elem := reflect.New(elemType)
			convertAssignSlice(elem.Elem(), v, arrJoinLevel2)
			slc = reflect.Append(slc, elem.Elem())
		}
	} else {
		for _, v := range strs {
			slc = reflect.Append(slc, fromString(reflect.New(d.Type().Elem()).Elem(), v))
		}
	}
	d.Set(slc)

	return true
}

func convertAssignString(d reflect.Value, s interface{}) bool {
	switch s := s.(type) {
	case nil:
		d.SetString("")
	case []byte:
		d.SetString(string(s))
	case string:
		d.SetString(s)
	default:
		err := cannotConvert(d, s)
		fmt.Println(err)
		return false
	}

	return true
}

func fromString(d reflect.Value, s string) reflect.Value {
	switch d.Type().Kind() {
	case reflect.Bool:
		b := false
		if len(s) > 0 && s != "0" {
			b = true
		}
		d.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ii, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			fmt.Printf("convertAssignArrayInt: can't convert:%s\n", s)
			return d
		}
		d.SetInt(ii)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ii, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			fmt.Printf("convertAssignArrayUint: can't convert:%s\n", s)
			return d
		}
		d.SetUint(ii)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			fmt.Printf("convertAssignArrayFloat: can't convert:%s\n", s)
			return d
		}
		d.SetFloat(f)
	case reflect.String:
		d.SetString(s)
	case reflect.Ptr:
		fromString(d.Elem(), s)
	case reflect.Array:
		//convertAssignArray(d)
	case reflect.Slice:
	case reflect.Map:
	case reflect.Struct:
	default: //其他类型不予支持
	}
	return d
}

func convertAssignMap(d reflect.Value, s interface{}) bool {
	str, ok := s.(string)
	if !ok {
		fmt.Printf("convertAssignMap can't convert to string:%v\n", s)
		return false
	}

	elemType := d.Type().Elem()
	strs := strings.Split(str, mapJoinLevel1)
	m := reflect.MakeMap(d.Type())
	if elemType.Kind() == reflect.Array {
		for i := range strs {
			ss := strings.Split(strs[i], mapJoinLevel2)
			if len(ss) != 2 {
				fmt.Printf("warning: != 2, %s\n", strs[i])
				continue
			}
			kv := reflect.New(m.Type().Key())
			vv := reflect.New(m.Type().Elem())
			fromString(kv, ss[0])
			convertAssignArray(vv.Elem(), ss[1], arrJoinLevel1)
			m.SetMapIndex(kv.Elem(), vv.Elem())
		}
	} else if elemType.Kind() == reflect.Slice {
		for i := range strs {
			ss := strings.Split(strs[i], mapJoinLevel2)
			if len(ss) != 2 {
				fmt.Printf("warning: != 2, %s\n", strs[i])
				continue
			}
			kv := reflect.New(m.Type().Key())
			vv := reflect.New(m.Type().Elem())
			fromString(kv, ss[0])
			convertAssignSlice(vv.Elem(), ss[1], arrJoinLevel1)
			m.SetMapIndex(kv.Elem(), vv.Elem())
		}
	} else {
		for i := range strs {
			ss := strings.Split(strs[i], ":")
			if len(ss) != 2 {
				fmt.Printf("warning: != 2, %s\n", strs[i])
				continue
			}
			kv := reflect.New(m.Type().Key())
			vv := reflect.New(m.Type().Elem())
			fromString(kv, ss[0])
			fromString(vv, ss[1])
			m.SetMapIndex(kv.Elem(), vv.Elem())
		}
	}
	d.Set(m)

	return true
}

func loadProcessArrayBase(rd RedisClienter, v reflect.Value, key string) bool {
	str, err := redis.String(rd.Do("GET", key))
	if err != nil {
		fmt.Printf("GET %v failed: %v\n", key, err)
		return false
	}
	if len(str) <= 0 {
		return true
	}

	return convertAssignArray(v, str, arrJoinLevel1)
}

func loadProcessArray(rd RedisClienter, v reflect.Value, key string) bool {
	elemIsPointer := false
	elemType := v.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem() //so, multi-level pointer is not allowed !!!
		elemIsPointer = true
	}

	kkey := key + "s"

	ids, err := redis.Strings(rd.Do("SMEMBERS", kkey))
	if err != nil {
		fmt.Println("smembers skey and Ints failed.")
		return false
	}

	for i := 0; i < v.Len() && i < len(ids); i++ {
		elem := reflect.New(elemType)
		loadProcessStruct(rd, elem.Elem(), fmt.Sprintf("%s:%s", key, ids[i]))

		if elemIsPointer {
			v.Index(i).Set(elem)
		} else {
			v.Index(i).Set(elem.Elem())
		}
	}

	return true
}

func loadProcessSliceBase(rd RedisClienter, v reflect.Value, key string) bool {
	str, err := redis.String(rd.Do("GET", key))
	if err != nil {
		fmt.Printf("GET %v failed: %v\n", key, err)
		return false
	}
	if len(str) <= 0 {
		return true
	}

	return convertAssignSlice(v, str, arrJoinLevel1)
}

func loadProcessSlice(rd RedisClienter, v reflect.Value, key string) bool {
	elemIsPointer := false
	elemType := v.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem() //so, multi-level pointer is not allowed !!!
		elemIsPointer = true
	}

	kkey := key + "s"

	ids, err := redis.Strings(rd.Do("SMEMBERS", kkey))
	if err != nil {
		fmt.Println("smembers skey and Ints failed.")
		return false
	}

	slc := v
	for _, id := range ids {
		elem := reflect.New(elemType)
		loadProcessStruct(rd, elem.Elem(), fmt.Sprintf("%s:%s", key, id))

		if elemIsPointer {
			slc = reflect.Append(slc, elem)
		} else {
			slc = reflect.Append(slc, elem.Elem())
		}
	}
	v.Set(slc)

	return true
}

func loadProcessMapBase(rd RedisClienter, v reflect.Value, key string) bool {
	str, err := redis.String(rd.Do("GET", key))
	if err != nil {
		fmt.Printf("GET %v failed: %v\n", key, err)
		return false
	}
	if len(str) <= 0 {
		fmt.Println("loadProcessMapBase empty")
		return true
	}

	if v.IsZero() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	return convertAssignMap(v, str)
}

func loadProcessMap(rd RedisClienter, v reflect.Value, key string) bool {
	elemIsPointer := false
	elemType := v.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem() //so, multi-level pointer is not allowed !!!
		elemIsPointer = true
	}

	kkey := key + "s"

	ids, err := redis.Strings(rd.Do("SMEMBERS", kkey))
	if err != nil {
		fmt.Println("smembers skey and Ints failed.")
		return false
	}

	if v.IsZero() {
		v.Set(reflect.MakeMap(v.Type()))
	}
	for _, id := range ids {
		elem := reflect.New(elemType)
		loadProcessStruct(rd, elem.Elem(), fmt.Sprintf("%s:%s", key, id))

		kv := reflect.New(v.Type().Key())
		fromString(kv, id)
		if elemIsPointer {
			v.SetMapIndex(kv.Elem(), elem)
		} else {
			v.SetMapIndex(kv.Elem(), elem.Elem())
		}
	}

	return true
}

func loadProcessStruct(rd RedisClienter, v reflect.Value, key string) bool {
	//typ := reflect.TypeOf(v)
	//value := reflect.ValueOf(v)
	typ := v.Type()
	value := v
	num := value.NumField()

	src, err := redis.StringMap(rd.Do("HGETALL", key))
	if err != nil {
		fmt.Println(err)
		return false
	}

	for i := 0; i < num; i++ {
		tfield := typ.Field(i)
		vfield := value.Field(i)
		if !vfield.CanInterface() {
			continue
		}
		var fSaveTag string
		var ok bool
		if fSaveTag, ok = tfield.Tag.Lookup("leiorm"); !ok {
			if fSaveTag, ok = tfield.Tag.Lookup("leiormpri"); !ok {
				//both `leirom` and `leiormpri` are not found,
				//and the field will not be saved.
				continue
			}
		}

		if vfield.Kind() == reflect.Ptr {
			if !vfield.CanInterface() {
				fmt.Println("can't interface")
				return false
			}
			// Already a reflect.Ptr
			if vfield.IsNil() {
				vfield.Set(reflect.New(vfield.Type().Elem()))
			}
		}

		//if !d.CanSet() {
		//	fmt.Println("can't set")
		//	continue
		//}

		fv := src[fSaveTag]

		switch tfield.Type.Kind() {
		case reflect.Bool:
			convertAssignBool(vfield, fv)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			convertAssignInt(vfield, fv)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			convertAssignUint(vfield, fv)
		case reflect.Float32, reflect.Float64:
			convertAssignFloat(vfield, fv)
		case reflect.String:
			convertAssignString(vfield, fv)
		case reflect.Array:
			switch getSaveType(tfield.Type.Elem()) {
			case SaveTypeBase, SaveTypeBaseArray:
				convertAssignArray(vfield, fv, arrJoinLevel1)
			case SaveTypeStruct:
				loadProcessArray(rd, vfield, key+":"+fSaveTag)
			default:
				continue
			}
		case reflect.Slice:
			switch getSaveType(tfield.Type.Elem()) {
			case SaveTypeBase, SaveTypeBaseArray:
				convertAssignSlice(vfield, fv, arrJoinLevel1)
			case SaveTypeStruct:
				loadProcessSlice(rd, vfield, key+":"+fSaveTag)
			default:
				continue
			}
		case reflect.Map:
			switch getSaveType(tfield.Type) {
			case SaveTypeBaseMap:
				convertAssignMap(vfield, fv)
			case SaveTypeStruct:
				loadProcessMap(rd, vfield, key+":"+fSaveTag)
			default:
				continue
			}
		case reflect.Ptr:
			vfield = vfield.Elem()
			fallthrough
		case reflect.Struct:
			loadProcessStruct(rd, vfield, key+":"+fSaveTag)
		default: //
			continue
		}
	}

	return true
}

func loadProcessBool(rd RedisClienter, v reflect.Value, key interface{}) bool {
	b, err := redis.Bool(rd.Do("GET", key))
	if err != nil {
		fmt.Printf("GET %v failed: %+v\n", key, err)
		return false
	}
	v.SetBool(b)
	return true
}

func loadProcessInt(rd RedisClienter, v reflect.Value, key interface{}) bool {
	b, err := redis.Int64(rd.Do("GET", key))
	if err != nil {
		fmt.Printf("GET %v failed: %+v\n", key, err)
		return false
	}
	v.SetInt(b)
	return true
}

func loadProcessUint(rd RedisClienter, v reflect.Value, key interface{}) bool {
	b, err := redis.Uint64(rd.Do("GET", key))
	if err != nil {
		fmt.Printf("GET %v failed: %+v\n", key, err)
		return false
	}
	v.SetUint(b)
	return true
}

func loadProcessFloat(rd RedisClienter, v reflect.Value, key interface{}) bool {
	b, err := redis.Float64(rd.Do("GET", key))
	if err != nil {
		fmt.Printf("GET %v failed: %+v\n", key, err)
		return false
	}
	v.SetFloat(b)
	return true
}

func loadProcessString(rd RedisClienter, v reflect.Value, key interface{}) bool {
	b, err := redis.String(rd.Do("GET", key))
	if err != nil {
		fmt.Printf("GET %v failed: %+v\n", key, err)
		return false
	}
	v.SetString(b)
	return true
}

func LoadModel(rd RedisClienter, ider interface{}, skey interface{}) bool {
	value := reflect.ValueOf(ider)
	for value.Type().Kind() != reflect.Ptr {
		fmt.Println("must be a pointer of struct")
		return false
	}
	value = value.Elem()

	var key string
	if skey != nil {
		key = fmt.Sprintf("%v", skey)
	}

	switch value.Kind() {
	case reflect.Bool:
		return loadProcessBool(rd, value, key)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return loadProcessInt(rd, value, key)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return loadProcessUint(rd, value, key)
	case reflect.Float32, reflect.Float64:
		return loadProcessFloat(rd, value, key)
	case reflect.String:
		return loadProcessString(rd, value, key)
	case reflect.Struct:
		return loadProcessStruct(rd, value, key)
	case reflect.Array:
		switch getSaveType(value.Type().Elem()) {
		case SaveTypeBase, SaveTypeBaseArray:
			return loadProcessArrayBase(rd, value, key)
		case SaveTypeStruct:
			return loadProcessArray(rd, value, key)
		}
	case reflect.Slice:
		switch getSaveType(value.Type().Elem()) {
		case SaveTypeBase, SaveTypeBaseArray:
			return loadProcessSliceBase(rd, value, key)
		case SaveTypeStruct:
			return loadProcessSlice(rd, value, key)
		}
	case reflect.Map:
		switch getSaveType(value.Type()) {
		case SaveTypeBaseMap:
			return loadProcessMapBase(rd, value, key)
		case SaveTypeStruct:
			return loadProcessMap(rd, value, key)
		default:
			return false
		}

	////
	//case reflect.Ptr: // Already a reflect.Ptr

	//case reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Chan:
	//case reflect.Interface:
	//case reflect.UnsafePointer:
	default:
	}

	return false
}
