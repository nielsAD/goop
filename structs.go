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
		if m, ok := dst[k].(map[string]interface{}); ok {
			DeleteEqual(m, src[k].(map[string]interface{}))
		}
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

// MergeOptions for Merge()
type MergeOptions struct {
	Overwrite bool
	Delete    bool
}

func mergeMap(dst reflect.Value, key reflect.Value, val reflect.Value, opt *MergeOptions) ([]string, error) {
	if dst.Kind() != reflect.Map {
		return nil, ErrTypeMismatch
	}

	if dst.IsNil() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}

	var idx = reflect.New(dst.Type().Elem()).Elem()
	var old = dst.MapIndex(key)
	if old.IsValid() {
		if err := Assign(idx, old); err != nil {
			return nil, err
		}
	}

	undecoded, err := merge(idx, val, opt)
	if err != nil {
		return nil, err
	}

	dst.SetMapIndex(key, idx)
	return undecoded, nil
}

func merge(dst reflect.Value, src reflect.Value, opt *MergeOptions) ([]string, error) {
	if src.Kind() != reflect.Ptr && src.Kind() != reflect.Interface && dst.Kind() == reflect.Ptr && dst.IsNil() {
		dst.Set(reflect.New(dst.Type().Elem()))
		return merge(dst.Elem(), src, opt)
	}

	switch src.Kind() {
	case reflect.Interface:
		fallthrough
	case reflect.Ptr:
		if src.IsNil() {
			if opt.Overwrite {
				return nil, Assign(dst, src)
			}
			return nil, nil
		}
		return merge(dst, src.Elem(), opt)
	case reflect.Slice, reflect.Array:
		switch dst.Kind() {
		case reflect.Array:
			if dst.Len() < src.Len() {
				return nil, ErrTypeMismatch
			}
		case reflect.Slice:
			if !empty(dst) && !opt.Overwrite {
				return nil, nil
			}
			if src.IsNil() {
				return nil, Assign(dst, src)
			}
			dst.Set(reflect.MakeSlice(dst.Type(), src.Len(), src.Len()))
		case reflect.Interface:
			if !src.Type().AssignableTo(dst.Type()) {
				return nil, ErrTypeMismatch
			}
			if !empty(dst) && !opt.Overwrite {
				return nil, nil
			}
			if src.IsNil() {
				return nil, Assign(dst, src)
			}
			dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Len()))
			dst = dst.Elem()
		default:
			return nil, ErrTypeMismatch
		}

		var undecoded = make([]string, 0)
		for i := 0; i < src.Len(); i++ {
			undec, err := merge(dst.Index(i), src.Index(i), opt)
			if err != nil {
				return nil, err
			}
			for _, u := range undec {
				undecoded = append(undecoded, fmt.Sprintf("%d.%s", i, u))
			}
		}
		return undecoded, nil
	case reflect.Map:
		var dstmap = dst.Kind() == reflect.Map
		if dst.Kind() == reflect.Interface {
			if !src.Type().AssignableTo(dst.Type()) {
				return nil, ErrTypeMismatch
			}
			if !empty(dst) && !opt.Overwrite {
				return nil, nil
			}
			if src.IsNil() {
				return nil, Assign(dst, src)
			}
			dst.Set(reflect.MakeMap(src.Type()))
			dst = dst.Elem()
			dstmap = true
		}

		var undecoded = make([]string, 0)
		for _, key := range src.MapKeys() {
			n := fmt.Sprintf("%v", key.Interface())
			k := find(dst, []string{n})

			var undec []string
			var err error

			if k != nil && k.CanSet() {
				undec, err = merge(*k, src.MapIndex(key), opt)
			} else if dstmap {
				undec, err = mergeMap(dst, key, src.MapIndex(key), opt)
			} else if k != nil {
				return nil, ErrTypeMismatch
			} else {
				undecoded = append(undecoded, n)
				continue
			}

			if err != nil {
				return nil, err
			}
			for _, u := range undec {
				undecoded = append(undecoded, fmt.Sprintf("%s.%s", n, u))
			}
		}
		if dstmap && opt.Delete {
			for _, key := range dst.MapKeys() {
				n := fmt.Sprintf("%v", key.Interface())
				if find(src, []string{n}) != nil {
					continue
				}

				dst.SetMapIndex(key, reflect.Value{})
			}
		}
		return undecoded, nil
	case reflect.Struct:
		var undecoded = make([]string, 0)
		for i := 0; i < src.NumField(); i++ {
			var f = src.Field(i)
			if !f.CanInterface() || (empty(f) && !opt.Overwrite) {
				continue
			}

			var t = src.Type().Field(i)
			var k = find(dst, []string{t.Name})
			if k == nil && t.Anonymous {
				k = &dst
			}

			var undec []string
			var err error

			if k != nil && k.CanSet() {
				undec, err = merge(*k, f, opt)
			} else {
				undec, err = mergeMap(dst, reflect.ValueOf(t.Name), f, opt)
			}
			if err != nil {
				return nil, err
			}
			for _, u := range undec {
				undecoded = append(undecoded, fmt.Sprintf("%s.%s", t.Name, u))
			}
		}
		return undecoded, nil
	default:
		if empty(dst) || opt.Overwrite {
			return nil, Assign(dst, src)
		}
		return nil, nil
	}
}

// Merge maps map[string]interface{} back to struct
func Merge(dst interface{}, src interface{}, opt *MergeOptions) ([]string, error) {
	return merge(reflect.ValueOf(dst), reflect.ValueOf(src), opt)
}

// Map val to map[string]interface{} equivalent
func Map(val interface{}) interface{} {
	var v = reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Invalid:
		return nil
	case reflect.Interface:
		fallthrough
	case reflect.Ptr:
		if v.IsNil() {
			return v.Interface()
		}
		return Map(v.Elem().Interface())
	case reflect.Slice, reflect.Array:
		var r = make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			r[i] = Map(v.Index(i).Interface())
		}
		return r
	case reflect.Map:
		var m = make(map[string]interface{})
		for _, key := range v.MapKeys() {
			m[fmt.Sprintf("%v", key.Interface())] = Map(v.MapIndex(key).Interface())
		}
		return m
	case reflect.Struct:
		var m = make(map[string]interface{})
		for i := 0; i < v.NumField(); i++ {
			var f = v.Field(i)
			if !f.CanInterface() {
				continue
			}

			var t = v.Type().Field(i)
			var x = Map(f.Interface())
			if xx, ok := x.(map[string]interface{}); t.Anonymous && ok {
				for k, v := range xx {
					m[k] = v
				}
			} else {
				m[t.Name] = x
			}
		}
		return m
	default:
		return v.Interface()
	}
}

func flatMap(prf string, val reflect.Value, dst map[string]interface{}) {
	switch val.Kind() {
	case reflect.Invalid:
		dst[strings.ToLower(prf)] = nil
	case reflect.Interface:
		fallthrough
	case reflect.Ptr:
		if val.IsNil() {
			dst[strings.ToLower(prf)] = val.Interface()
		} else {
			flatMap(prf, val.Elem(), dst)
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
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			var f = val.Field(i)
			if !f.CanInterface() {
				continue
			}

			var t = val.Type().Field(i)
			var pre = t.Name
			if t.Anonymous {
				pre = prf
			} else if prf != "" {
				pre = fmt.Sprintf("%s/%v", prf, t.Name)
			}
			flatMap(pre, f, dst)
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
	case reflect.Slice, reflect.Array:
		n, err := strconv.ParseInt(key[0], 0, 64)
		if err != nil || n < 0 || n >= int64(val.Len()) {
			return nil
		}
		return find(val.Index(int(n)), key[1:])
	case reflect.Map:
		for _, k := range val.MapKeys() {
			if strings.EqualFold(key[0], fmt.Sprintf("%v", k.Interface())) {
				return find(val.MapIndex(k), key[1:])
			}
		}
	case reflect.Struct:
		var anon *reflect.Value
		for i := 0; i < val.NumField(); i++ {
			var f = val.Field(i)
			if !f.CanInterface() {
				continue
			}

			var t = val.Type().Field(i)
			if t.Anonymous {
				if v := find(f, key); v != nil {
					anon = v
				}
			}
			if strings.EqualFold(key[0], t.Name) {
				return find(f, key[1:])
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
		} else {
			if dst.Kind() == reflect.Ptr {
				if !dst.IsNil() {
					return Assign(dst.Elem(), src)
				}
				var tmp = reflect.New(dst.Type().Elem())
				if err := Assign(tmp.Elem(), src); err != nil {
					return err
				}
				dst.Set(tmp)
				return nil
			} else if src.Kind() == reflect.String && src.CanInterface() && dst.CanAddr() {
				if i, ok := dst.Addr().Interface().(encoding.TextUnmarshaler); ok {
					return i.UnmarshalText([]byte(src.Interface().(string)))
				}
			}

			return ErrTypeMismatch
		}
	}

	dst.Set(src)
	return nil
}

// AssignString src to dst
func AssignString(dst reflect.Value, src string) error {
	if dst.CanAddr() {
		if i, ok := dst.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return i.UnmarshalText([]byte(src))
		}
	}

	switch dst.Kind() {
	case reflect.Ptr:
		if !dst.IsNil() {
			return AssignString(dst.Elem(), src)
		}
		var tmp = reflect.New(dst.Type().Elem())
		if err := AssignString(tmp.Elem(), src); err != nil {
			return err
		}
		dst.Set(tmp)
		return nil
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
		n, err := strconv.ParseInt(src, 0, 64)
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
	if f == nil || !f.CanSet() {
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

	if f != nil && f.CanSet() {
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
	if f == nil || !f.CanSet() {
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
