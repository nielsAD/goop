// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package main

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Errors
var (
	ErrUnknownKey   = errors.New("structs: Unknown key")
	ErrTypeMismatch = errors.New("structs: Type mismatch")
)

// ParentKey of key
func ParentKey(key string) (string, string) {
	if strings.HasSuffix(key, "[]") {
		return key[0 : len(key)-2], "[]"
	}

	var idx = strings.LastIndexByte(key, '/')
	if idx == -1 {
		return "", ""
	}

	return key[0:idx], key[idx+1 : len(key)]
}

// DeleteEqual entries where dst[k] == src[k] (recursively)
func DeleteEqual(dst map[string]interface{}, src map[string]interface{}) {
	for k := range src {
		if reflect.DeepEqual(src[k], dst[k]) {
			delete(dst, k)
			continue
		}
		var v, ok = dst[k].(map[string]interface{})
		if ok {
			DeleteEqual(v, src[k].(map[string]interface{}))
		}
	}
}

// Map val to map[string]interface{} equivalent
func Map(val interface{}) interface{} {
	var v = reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Interface:
		fallthrough
	case reflect.Ptr:
		if v.IsNil() {
			return nil
		}
		return Map(v.Elem().Interface())
	case reflect.Map:
		var m = make(map[string]interface{})
		for _, key := range v.MapKeys() {
			m[fmt.Sprintf("%v", key.Interface())] = Map(v.MapIndex(key).Interface())
		}
		return m
	case reflect.Slice, reflect.Array:
		var r = make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			r[i] = Map(v.Index(i).Interface())
		}
		return r
	case reflect.Struct:
		var m = make(map[string]interface{})
		for i := 0; i < v.NumField(); i++ {
			var f = v.Type().Field(i)
			if f.Name == "" {
				continue
			}

			var x = Map(v.Field(i).Interface())
			if xx, ok := x.(map[string]interface{}); f.Anonymous && ok {
				for k, v := range xx {
					m[k] = v
				}
			} else {
				m[f.Name] = x
			}
		}
		return m
	default:
		return v.Interface()
	}
}

func empty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return empty(v.Elem())
	case reflect.Func:
		return v.IsNil()
	case reflect.Invalid:
		return true
	}
	return false
}

func mergeMap(dst reflect.Value, key reflect.Value, src reflect.Value, overwrite bool) error {
	if dst.Kind() != reflect.Map {
		return ErrTypeMismatch
	}

	if dst.IsNil() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}

	var idx = reflect.New(dst.Type().Elem()).Elem()
	var old = dst.MapIndex(key)
	if old.IsValid() {
		if err := Assign(idx, old); err != nil {
			return err
		}
	}
	if err := merge(idx, src, overwrite); err != nil {
		return err
	}

	dst.SetMapIndex(key, idx)
	return nil
}

func merge(dst reflect.Value, src reflect.Value, overwrite bool) error {
	switch src.Kind() {
	case reflect.Interface:
		fallthrough
	case reflect.Ptr:
		if src.IsNil() {
			if overwrite {
				return Assign(dst, src)
			}
			return nil
		}
		return merge(dst, src.Elem(), overwrite)
	case reflect.Map:
		if dst.Kind() == reflect.Ptr && dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
			return merge(dst.Elem(), src, overwrite)
		}

		for _, key := range src.MapKeys() {
			k := find(dst, []string{fmt.Sprintf("%v", key.Interface())})

			if k == nil || !k.CanSet() {
				if err := mergeMap(dst, key, src.MapIndex(key), overwrite); err != nil {
					return err
				}
				continue
			}

			if err := merge(*k, src.MapIndex(key), overwrite); err != nil {
				return err
			}
		}
		return nil
	case reflect.Struct:
		if dst.Kind() == reflect.Ptr && dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
			return merge(dst.Elem(), src, overwrite)
		}

		for i := 0; i < src.NumField(); i++ {
			var v = src.Field(i)
			if empty(v) && !overwrite {
				continue
			}

			var f = src.Type().Field(i)
			if f.Name == "" {
				continue
			}

			k := find(dst, []string{f.Name})

			if k == nil && f.Anonymous {
				k = &dst
			}
			if k == nil || !k.CanSet() {
				if err := mergeMap(dst, reflect.ValueOf(f.Name), v, overwrite); err != nil {
					return err
				}
				continue
			}

			if err := merge(*k, v, overwrite); err != nil {
				return err
			}
		}
		return nil
	default:
		if empty(dst) || overwrite {
			return Assign(dst, src)
		}
		return nil
	}
}

// Merge maps map[string]interface{} back to struct
func Merge(dst interface{}, src interface{}, overwrite bool) error {
	return merge(reflect.ValueOf(dst), reflect.ValueOf(src), overwrite)
}

func flatMap(prf string, val reflect.Value, dst map[string]interface{}) {
	switch val.Kind() {
	case reflect.Interface:
		fallthrough
	case reflect.Ptr:
		if !val.IsNil() {
			flatMap(prf, val.Elem(), dst)
		}
		dst[strings.ToLower(prf)] = val.Interface()
	case reflect.Map:
		dst[strings.ToLower(prf)] = val
		for _, key := range val.MapKeys() {
			var pre string
			if prf == "" {
				pre = fmt.Sprintf("%v", key.Interface())
			} else {
				pre = fmt.Sprintf("%s/%v", prf, key.Interface())
			}
			flatMap(pre, val.MapIndex(key), dst)
		}
	case reflect.Slice:
		dst[strings.ToLower(prf)] = val.Interface()
		fallthrough
	case reflect.Array:
		for i := 0; i < val.Len(); i++ {
			var pre string
			if prf == "" {
				pre = fmt.Sprintf("%d", i)
			} else {
				pre = fmt.Sprintf("%s/%d", prf, i)
			}
			flatMap(pre, val.Index(i), dst)
		}
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			var f = val.Type().Field(i)
			if f.Name == "" {
				continue
			}

			var pre = f.Name
			if f.Anonymous {
				pre = prf
			} else if prf != "" {
				pre = fmt.Sprintf("%s/%v", prf, f.Name)
			}
			flatMap(pre, val.Field(i), dst)
		}
	default:
		dst[strings.ToLower(prf)] = val.Interface()
	}
}

// FlatMap maps all the (nested) keys to their reflect.Value
func FlatMap(val interface{}) map[string]interface{} {
	var f = make(map[string]interface{})
	flatMap("", reflect.ValueOf(val), f)
	return f
}

func find(val reflect.Value, key []string) *reflect.Value {
	if len(key) == 0 {
		return &val
	}

	switch val.Kind() {
	case reflect.Interface:
		fallthrough
	case reflect.Ptr:
		if val.IsNil() {
			return nil
		}
		return find(val.Elem(), key)
	case reflect.Map:
		for _, k := range val.MapKeys() {
			if strings.EqualFold(key[0], fmt.Sprintf("%v", k.Interface())) {
				return find(val.MapIndex(k), key[1:])
			}
		}
	case reflect.Slice, reflect.Array:
		n, err := strconv.ParseInt(key[0], 10, 64)
		if err != nil || n < 0 || n >= int64(val.Len()) {
			return nil
		}
		return find(val.Index(int(n)), key[1:])
	case reflect.Struct:
		var anon *reflect.Value
		for i := 0; i < val.NumField(); i++ {
			var f = val.Type().Field(i)
			if f.Anonymous {
				if v := find(val.Field(i), key); v != nil {
					anon = v
				}
			}
			if strings.EqualFold(key[0], f.Name) {
				return find(val.Field(i), key[1:])
			}
		}
		return anon
	}

	return nil
}

// Find tries to find the (nested) reflect.Value for key
func Find(val interface{}, key string) *reflect.Value {
	return find(reflect.ValueOf(val), strings.Split(key, "/"))
}

// Assign src to dst
func Assign(dst, src reflect.Value) error {
	if !src.Type().AssignableTo(dst.Type()) {
		if src.Type().ConvertibleTo(dst.Type()) {
			src = src.Convert(dst.Type())
		} else if dst.Kind() == reflect.Ptr {
			if dst.IsNil() {
				dst.Set(reflect.New(dst.Type().Elem()))
			}
			return Assign(dst.Elem(), src)
		} else {
			return ErrTypeMismatch
		}
	}

	dst.Set(src)
	return nil
}

// AssignString src to dst
func AssignString(dst reflect.Value, src string) error {
	if i, ok := dst.Interface().(encoding.TextUnmarshaler); ok {
		return i.UnmarshalText([]byte(src))
	}

	switch dst.Kind() {
	case reflect.Ptr:
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return AssignString(dst.Elem(), src)
	case reflect.String:
		dst.SetString(src)
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(src)
		if err != nil {
			return err
		}
		dst.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(src, 10, 64)
		if err != nil {
			return err
		}
		dst.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(src, 10, 64)
		if err != nil {
			return err
		}
		dst.SetUint(n)
		return nil
	default:
		return ErrTypeMismatch
	}
}

// Get config value via flat index string
func Get(src interface{}, key string) (interface{}, error) {
	var val = Find(src, key)
	if val == nil {
		return nil, ErrUnknownKey
	}
	return val.Interface(), nil
}

// Set config value via flat index string
func Set(dst interface{}, key string, val interface{}) error {
	var f = Find(dst, key)
	if f != nil && f.CanSet() {
		return Assign(*f, reflect.ValueOf(val))
	}

	parent, key := ParentKey(key)
	if parent == "" || key == "" {
		return ErrUnknownKey
	}

	f = Find(dst, parent)
	if f == nil {
		return ErrUnknownKey
	}

	switch f.Kind() {
	case reflect.Map:
		if f.IsNil() {
			f.Set(reflect.MakeMap(f.Type()))
		}

		var idx = reflect.New(f.Type().Key()).Elem()
		if err := AssignString(idx, key); err != nil {
			return err
		}

		var tmp = reflect.New(f.Type().Elem()).Elem()
		if err := Assign(tmp, reflect.ValueOf(val)); err != nil {
			return err
		}

		f.SetMapIndex(idx, tmp)
		return nil
	case reflect.Slice:
		if key != "[]" {
			return ErrUnknownKey
		}

		var tmp = reflect.New(f.Type().Elem()).Elem()
		if err := Assign(tmp, reflect.ValueOf(val)); err != nil {
			return err
		}

		f.Set(reflect.Append(*f, tmp))
		return nil
	default:
		return ErrUnknownKey
	}
}

// Unset config value via flat index string
func Unset(dst interface{}, key string) (err error) {
	var f = Find(dst, key)
	if f == nil {
		return ErrUnknownKey
	}
	if f.CanSet() {
		err = Assign(*f, reflect.Zero(f.Type()))
	}

	parent, key := ParentKey(key)
	f = Find(dst, parent)

	if f != nil {
		switch f.Kind() {
		case reflect.Map:
			var idx = reflect.New(f.Type().Key()).Elem()
			if err := AssignString(idx, key); err != nil {
				return err
			}

			f.SetMapIndex(idx, reflect.Value{})
			return nil
		case reflect.Slice:
			idx, err := strconv.Atoi(key)
			if err != nil {
				return err
			}

			var len = f.Len()
			f.Set(reflect.AppendSlice(f.Slice(0, idx), f.Slice(idx+1, len)))
			return nil
		}
	}
	return err
}

// GetString config value via flat index string
func GetString(src interface{}, key string) (string, error) {
	val, err := Get(src, key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", val), nil
}

// SetString config value via flat index string
func SetString(dst interface{}, key string, val string) error {
	var f = Find(dst, key)
	if f != nil && f.CanSet() {
		return AssignString(*f, val)
	}

	parent, key := ParentKey(key)
	if parent == "" || key == "" {
		return ErrUnknownKey
	}

	f = Find(dst, parent)
	if f == nil {
		return ErrUnknownKey
	}

	switch f.Kind() {
	case reflect.Map:
		if f.IsNil() {
			f.Set(reflect.MakeMap(f.Type()))
		}

		var idx = reflect.New(f.Type().Key()).Elem()
		if err := AssignString(idx, key); err != nil {
			return err
		}

		var tmp = reflect.New(f.Type().Elem()).Elem()
		if err := AssignString(tmp, val); err != nil {
			return err
		}

		f.SetMapIndex(idx, tmp)
		return nil
	case reflect.Slice:
		if key != "[]" {
			return ErrUnknownKey
		}

		var tmp = reflect.New(f.Type().Elem()).Elem()
		if err := AssignString(tmp, val); err != nil {
			return err
		}

		f.Set(reflect.Append(*f, tmp))
		return nil
	default:
		return ErrUnknownKey
	}
}
