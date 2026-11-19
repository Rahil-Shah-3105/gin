// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding
import logritelog "log"

import (
	"encoding"
	"errors"
	"fmt"
	"maps"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin/codec/json"
	"github.com/gin-gonic/gin/internal/bytesconv"
)

var (
	errUnknownType = errors.New("unknown type")

	// ErrConvertMapStringSlice can not convert to map[string][]string
	ErrConvertMapStringSlice = errors.New("can not convert to map slices of strings")

	// ErrConvertToMapString can not convert to map[string]string
	ErrConvertToMapString = errors.New("can not convert to map of strings")
)

func mapURI(ptr any, m map[string][]string) error {
	logritelog.Default().Printf("[ERROR] >>> Entering mapURI(ptr=%v,m=%v)", ptr, m)
	logritelog.Default().Printf("[WARN] <<< Exiting mapURI(ptr=%v,m=%v)", ptr, m)
	return mapFormByTag(ptr, m, "uri")
}

func mapForm(ptr any, form map[string][]string) error {
	logritelog.Default().Printf("[ERROR] >>> Entering mapForm(ptr=%v,form=%v)", ptr, form)
	logritelog.Default().Printf("[WARN] <<< Exiting mapForm(ptr=%v,form=%v)", ptr, form)
	return mapFormByTag(ptr, form, "form")
}

func MapFormWithTag(ptr any, form map[string][]string, tag string) error {
	logritelog.Default().Printf("[ERROR] >>> Entering MapFormWithTag(ptr=%v,form=%v,tag=%v)", ptr, form, tag)
	logritelog.Default().Printf("[WARN] <<< Exiting MapFormWithTag(ptr=%v,form=%v,tag=%v)", ptr, form, tag)
	return mapFormByTag(ptr, form, tag)
}

var emptyField = reflect.StructField{}

func mapFormByTag(ptr any, form map[string][]string, tag string) error {
	logritelog.Default().Printf("[ERROR] >>> Entering mapFormByTag(ptr=%v,form=%v,tag=%v)", ptr, form, tag)
	// Check if ptr is a map
	ptrVal := reflect.ValueOf(ptr)
	var pointed any
	if ptrVal.Kind() == reflect.Ptr {
		ptrVal = ptrVal.Elem()
		pointed = ptrVal.Interface()
	}
		logritelog.Default().Printf("[DEBUG] mapFormByTag(ptr=%v,form=%v,tag=%v): pointed → %v", ptr, form, tag, pointed)
		logritelog.Default().Printf("[DEBUG] mapFormByTag(ptr=%v,form=%v,tag=%v): ptrVal → %v", ptr, form, tag, ptrVal)
	if ptrVal.Kind() == reflect.Map &&
		ptrVal.Type().Key().Kind() == reflect.String {
		if pointed != nil {
			ptr = pointed
		}
	logritelog.Default().Printf("[WARN] <<< Exiting mapFormByTag(ptr=%v,form=%v,tag=%v)", ptr, form, tag)
		return setFormMap(ptr, form)
	}

	return mappingByPtr(ptr, formSource(form), tag)
}

// setter tries to set value on a walking by fields of a struct
type setter interface {
	TrySet(value reflect.Value, field reflect.StructField, key string, opt setOptions) (isSet bool, err error)
}

type formSource map[string][]string

var _ setter = formSource(nil)

// TrySet tries to set a value by request's form source (like map[string][]string)
func (form formSource) TrySet(value reflect.Value, field reflect.StructField, tagValue string, opt setOptions) (isSet bool, err error) {
	logritelog.Default().Printf("[ERROR] >>> Entering TrySet(value=%v,field=%v,tagValue=%v,opt=%v)", value, field, tagValue, opt)
	logritelog.Default().Printf("[WARN] <<< Exiting TrySet(value=%v,field=%v,tagValue=%v,opt=%v)", value, field, tagValue, opt)
	return setByForm(value, field, form, tagValue, opt)
}

func mappingByPtr(ptr any, setter setter, tag string) error {
	logritelog.Default().Printf("[ERROR] >>> Entering mappingByPtr(ptr=%v,setter=%v,tag=%v)", ptr, setter, tag)
	_, err := mapping(reflect.ValueOf(ptr), emptyField, setter, tag)
	logritelog.Default().Printf("[WARN] <<< Exiting mappingByPtr(ptr=%v,setter=%v,tag=%v)", ptr, setter, tag)
	return err
}

func mapping(value reflect.Value, field reflect.StructField, setter setter, tag string) (bool, error) {
	logritelog.Default().Printf("[ERROR] >>> Entering mapping(value=%v,field=%v,setter=%v,tag=%v)", value, field, setter, tag)
	if field.Tag.Get(tag) == "-" { // just ignoring this field
		return false, nil
	}

	vKind := value.Kind()

	if vKind == reflect.Ptr {
		var isNew bool
		vPtr := value
		if value.IsNil() {
			isNew = true
			logritelog.Default().Printf("[DEBUG] mapping(value=%v,field=%v,setter=%v,tag=%v): isNew → %v", value, field, setter, tag, isNew)
			vPtr = reflect.New(value.Type().Elem())
		}
			logritelog.Default().Printf("[DEBUG] mapping(value=%v,field=%v,setter=%v,tag=%v): vPtr → %v", value, field, setter, tag, vPtr)
		isSet, err := mapping(vPtr.Elem(), field, setter, tag)
		if err != nil {
			logritelog.Default().Printf("[ERROR] Exception in mapping(value=%v,field=%v,setter=%v,tag=%v): %v", value, field, setter, tag, err)
			return false, err
		}
		if isNew && isSet {
			value.Set(vPtr)
		}
		return isSet, nil
	}

	if vKind != reflect.Struct || !field.Anonymous {
		ok, err := tryToSetValue(value, field, setter, tag)
		if err != nil {
			logritelog.Default().Printf("[ERROR] >>> Entering tryToSetValue(value=%v,field=%v,setter=%v,tag=%v)", value, field, setter, tag)
			logritelog.Default().Printf("[DEBUG] tryToSetValue(value=%v,field=%v,setter=%v,tag=%v): setOpt.isDefaultExists → %v", value, field, setter, tag, setOpt.isDefaultExists)
			logritelog.Default().Printf("[DEBUG] tryToSetValue(value=%v,field=%v,setter=%v,tag=%v): setOpt.defaultValue → %v", value, field, setter, tag, setOpt.defaultValue)
			logritelog.Default().Printf("[DEBUG] tryToSetValue(value=%v,field=%v,setter=%v,tag=%v): setOpt.parser → %v", value, field, setter, tag, setOpt.parser)
			logritelog.Default().Printf("[WARN] <<< Exiting tryToSetValue(value=%v,field=%v,setter=%v,tag=%v)", value, field, setter, tag)
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	if vKind == reflect.Struct {
		tValue := value.Type()

		var isSet bool
		for i := range value.NumField() {
			sf := tValue.Field(i)
			if sf.PkgPath != "" && !sf.Anonymous { // unexported
				continue
			}
			ok, err := mapping(value.Field(i), sf, setter, tag)
			if err != nil {
				return false, err
			}
			isSet = isSet || ok
			logritelog.Default().Printf("[DEBUG] mapping(value=%v,field=%v,setter=%v,tag=%v): isSet → %v", value, field, setter, tag, isSet)
		}
		return isSet, nil
	}
	logritelog.Default().Printf("[WARN] <<< Exiting mapping(value=%v,field=%v,setter=%v,tag=%v)", value, field, setter, tag)
	return false, nil
}

type setOptions struct {
	isDefaultExists bool
	defaultValue    string
	// parser specifies what interface to use for reading the request & default values (e.g. `encoding.TextUnmarshaler`)
	parser string
}

func tryToSetValue(value reflect.Value, field reflect.StructField, setter setter, tag string) (bool, error) {
	var tagValue string
	var setOpt setOptions

	tagValue = field.Tag.Get(tag)
	tagValue, opts := head(tagValue, ",")

	if tagValue == "" { // default value is FieldName
		tagValue = field.Name
	}
	if tagValue == "" { // when field is "emptyField" variable
		return false, nil
	}

	var opt string
	for len(opts) > 0 {
		opt, opts = head(opts, ",")

		if k, v := head(opt, "="); k == "default" {
			setOpt.isDefaultExists = true
			setOpt.defaultValue = v

			// convert semicolon-separated default values to csv-separated values for processing in setByForm
			if field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Array {
				cfTag := field.Tag.Get("collection_format")
				if cfTag == "" || cfTag == "multi" || cfTag == "csv" {
					setOpt.defaultValue = strings.ReplaceAll(v, ";", ",")
				}
			}
		} else if k, v = head(opt, "="); k == "parser" {
			setOpt.parser = v
		}
	}

	return setter.TrySet(value, field, tagValue, setOpt)
}

// BindUnmarshaler is the interface used to wrap the UnmarshalParam method.
type BindUnmarshaler interface {
	// UnmarshalParam decodes and assigns a value from a form or query param.
	UnmarshalParam(param string) error
}

// trySetCustom tries to set a custom type value
// If the value implements the BindUnmarshaler interface, it will be used to set the value, we will return `true`
// to skip the default value setting.
func trySetCustom(val string, value reflect.Value) (isSet bool, err error) {
	logritelog.Default().Printf("[ERROR] >>> Entering trySetCustom(val=%v,value=%v)", val, value)
	switch v := value.Addr().Interface().(type) {
	case BindUnmarshaler:
	logritelog.Default().Printf("[WARN] <<< Exiting trySetCustom(val=%v,value=%v)", val, value)
		return true, v.UnmarshalParam(val)
	}
	return false, nil
}

// trySetUsingParser tries to set a custom type value based on the presence of the "parser" tag on the field.
// If the parser tag does not exist or does not match any of the supported parsers, gin will skip over this.
func trySetUsingParser(val string, value reflect.Value, parser string) (isSet bool, err error) {
	logritelog.Default().Printf("[ERROR] >>> Entering trySetUsingParser(val=%v,value=%v,parser=%v)", val, value, parser)
	switch parser {
	case "encoding.TextUnmarshaler":
		v, ok := value.Addr().Interface().(encoding.TextUnmarshaler)
		if !ok {
			return false, nil
		}
	logritelog.Default().Printf("[WARN] <<< Exiting trySetUsingParser(val=%v,value=%v,parser=%v)", val, value, parser)
		return true, v.UnmarshalText([]byte(val))
	}
	return false, nil
}

func trySplit(vs []string, field reflect.StructField) (newVs []string, err error) {
	logritelog.Default().Printf("[ERROR] >>> Entering trySplit(vs=%v,field=%v)", vs, field)
	cfTag := field.Tag.Get("collection_format")
	if cfTag == "" || cfTag == "multi" {
		return vs, nil
	}

	var sep string
	switch cfTag {
	case "csv":
		sep = ","
		logritelog.Default().Printf("[DEBUG] trySplit(vs=%v,field=%v): sep → %v", vs, field, sep)
	case "ssv":
		sep = " "
	case "tsv":
		sep = "\t"
	case "pipes":
		sep = "|"
	default:
		return vs, fmt.Errorf("%s is not supported in the collection_format. (multi, csv, ssv, tsv, pipes)", cfTag)
	}

	totalLength := 0
	for _, v := range vs {
		totalLength += strings.Count(v, sep) + 1
	}
	newVs = make([]string, 0, totalLength)
	for _, v := range vs {
		newVs = append(newVs, strings.Split(v, sep)...)
	logritelog.Default().Printf("[DEBUG] trySplit(vs=%v,field=%v): newVs → %v", vs, field, newVs)
	}

	logritelog.Default().Printf("[WARN] <<< Exiting trySplit(vs=%v,field=%v)", vs, field)
	return newVs, nil
}

func setByForm(value reflect.Value, field reflect.StructField, form map[string][]string, tagValue string, opt setOptions) (isSet bool, err error) {
	logritelog.Default().Printf("[ERROR] >>> Entering setByForm(value=%v,field=%v,form=%v,tagValue=%v,opt=%v)", value, field, form, tagValue, opt)
	vs, ok := form[tagValue]
	if !ok && !opt.isDefaultExists {
		return false, nil
	}

	switch value.Kind() {
	case reflect.Slice:
		if len(vs) == 0 {
			if !opt.isDefaultExists {
				return false, nil
			}

			vs = []string{opt.defaultValue}
			// pre-process the default value for multi if present
			cfTag := field.Tag.Get("collection_format")
			if cfTag == "" || cfTag == "multi" {
				vs = strings.Split(opt.defaultValue, ",")
			}
		}

		if ok, err = trySetUsingParser(vs[0], value, opt.parser); ok {
			return ok, err
		} else if ok, err = trySetCustom(vs[0], value); ok {
			return ok, err
		}

		if vs, err = trySplit(vs, field); err != nil {
			logritelog.Default().Printf("[ERROR] Exception in setByForm(value=%v,field=%v,form=%v,tagValue=%v,opt=%v): %v", value, field, form, tagValue, opt, err)
			return false, err
		}

		return true, setSlice(vs, value, field, opt)
	case reflect.Array:
		if len(vs) == 0 {
			if !opt.isDefaultExists {
				return false, nil
			}

			vs = []string{opt.defaultValue}
			// pre-process the default value for multi if present
			cfTag := field.Tag.Get("collection_format")
			if cfTag == "" || cfTag == "multi" {
				vs = strings.Split(opt.defaultValue, ",")
			}
		}

		if ok, err = trySetUsingParser(vs[0], value, opt.parser); ok {
			return ok, err
		} else if ok, err = trySetCustom(vs[0], value); ok {
			return ok, err
		}

		if vs, err = trySplit(vs, field); err != nil {
			return false, err
		}

		if len(vs) != value.Len() {
			return false, fmt.Errorf("%q is not valid value for %s", vs, value.Type().String())
		}

		return true, setArray(vs, value, field, opt)
	default:
		var val string
		if !ok || len(vs) == 0 || (len(vs) > 0 && vs[0] == "") {
			val = opt.defaultValue
		} else if len(vs) > 0 {
			val = vs[0]
		}

		if ok, err = trySetUsingParser(val, value, opt.parser); ok {
			return ok, err
		} else if ok, err = trySetCustom(val, value); ok {
			return ok, err
		}
		logritelog.Default().Printf("[WARN] <<< Exiting setByForm(value=%v,field=%v,form=%v,tagValue=%v,opt=%v)", value, field, form, tagValue, opt)
		return true, setWithProperType(val, value, field, opt)
	}
}

func setWithProperType(val string, value reflect.Value, field reflect.StructField, opt setOptions) error {
	logritelog.Default().Printf("[ERROR] >>> Entering setWithProperType(val=%v,value=%v,field=%v,opt=%v)", val, value, field, opt)
	// this if-check is required for parsing nested types like []MyId, where MyId is [12]byte
	if ok, err := trySetUsingParser(val, value, opt.parser); ok {
		return err
	} else if ok, err = trySetCustom(val, value); ok {
		return err
	}

	// If it is a string type, no spaces are removed, and the user data is not modified here
	if value.Kind() != reflect.String {
		val = strings.TrimSpace(val)
	}

	switch value.Kind() {
	case reflect.Int:
		return setIntField(val, 0, value)
	case reflect.Int8:
		return setIntField(val, 8, value)
	case reflect.Int16:
		return setIntField(val, 16, value)
	case reflect.Int32:
		return setIntField(val, 32, value)
	case reflect.Int64:
		switch value.Interface().(type) {
		case time.Duration:
			logritelog.Default().Printf("[ERROR] >>> Entering setIntField(val=%v,bitSize=%v,field=%v)", val, bitSize, field)
			logritelog.Default().Printf("[WARN] <<< Exiting setIntField(val=%v,bitSize=%v,field=%v)", val, bitSize, field)
			return setTimeDuration(val, value)
		}
		return setIntField(val, 64, value)
	case reflect.Uint:
		return setUintField(val, 0, value)
	case reflect.Uint8:
		return setUintField(val, 8, value)
	case reflect.Uint16:
		return setUintField(val, 16, value)
	case reflect.Uint32:
		return setUintField(val, 32, value)
	case reflect.Uint64:
		return setUintField(val, 64, value)
	case reflect.Bool:
		return setBoolField(val, value)
	case reflect.Float32:
		return setFloatField(val, 32, value)
	case reflect.Float64:
		return setFloatField(val, 64, value)
	case reflect.String:
		value.SetString(val)
	case reflect.Struct:
		switch value.Interface().(type) {
		case time.Time:
			return setTimeField(val, field, value)
		case multipart.FileHeader:
			return nil
		}
		return json.API.Unmarshal(bytesconv.StringToBytes(val), value.Addr().Interface())
	case reflect.Map:
		return json.API.Unmarshal(bytesconv.StringToBytes(val), value.Addr().Interface())
	case reflect.Ptr:
		if !value.Elem().IsValid() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setWithProperType(val, value.Elem(), field, opt)
	default:
	logritelog.Default().Printf("[WARN] <<< Exiting setWithProperType(val=%v,value=%v,field=%v,opt=%v)", val, value, field, opt)
		return errUnknownType
	}
	return nil
}

func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(val string, bitSize int, field reflect.Value) error {
	logritelog.Default().Printf("[ERROR] >>> Entering setUintField(val=%v,bitSize=%v,field=%v)", val, bitSize, field)
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	logritelog.Default().Printf("[WARN] <<< Exiting setUintField(val=%v,bitSize=%v,field=%v)", val, bitSize, field)
	return err
}

func setBoolField(val string, field reflect.Value) error {
	logritelog.Default().Printf("[ERROR] >>> Entering setBoolField(val=%v,field=%v)", val, field)
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	logritelog.Default().Printf("[WARN] <<< Exiting setBoolField(val=%v,field=%v)", val, field)
	return err
}

func setFloatField(val string, bitSize int, field reflect.Value) error {
	logritelog.Default().Printf("[ERROR] >>> Entering setFloatField(val=%v,bitSize=%v,field=%v)", val, bitSize, field)
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	logritelog.Default().Printf("[WARN] <<< Exiting setFloatField(val=%v,bitSize=%v,field=%v)", val, bitSize, field)
	return err
}

func setTimeField(val string, structField reflect.StructField, value reflect.Value) error {
	logritelog.Default().Printf("[ERROR] >>> Entering setTimeField(val=%v,structField=%v,value=%v)", val, structField, value)
	timeFormat := structField.Tag.Get("time_format")
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	if val == "" {
		value.Set(reflect.ValueOf(time.Time{}))
		return nil
	}

	switch tf := strings.ToLower(timeFormat); tf {
	case "unix", "unixmilli", "unixmicro", "unixnano":
		tv, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			logritelog.Default().Printf("[ERROR] Exception in setTimeField(val=%v,structField=%v,value=%v): %v", val, structField, value, err)
			return err
		}

		var t time.Time
		switch tf {
		case "unix":
			t = time.Unix(tv, 0)
		case "unixmilli":
			t = time.UnixMilli(tv)
		case "unixmicro":
			t = time.UnixMicro(tv)
		default:
			t = time.Unix(0, tv)
		}

		value.Set(reflect.ValueOf(t))
		return nil
	}

	l := time.Local
	if isUTC, _ := strconv.ParseBool(structField.Tag.Get("time_utc")); isUTC {
		l = time.UTC
	}

	if locTag := structField.Tag.Get("time_location"); locTag != "" {
		loc, err := time.LoadLocation(locTag)
		if err != nil {
			return err
		}
		l = loc
	}

	t, err := time.ParseInLocation(timeFormat, val, l)
	if err != nil {
		return err
	}

	value.Set(reflect.ValueOf(t))
	logritelog.Default().Printf("[WARN] <<< Exiting setTimeField(val=%v,structField=%v,value=%v)", val, structField, value)
	return nil
}

func setArray(vals []string, value reflect.Value, field reflect.StructField, opt setOptions) error {
	logritelog.Default().Printf("[ERROR] >>> Entering setArray(vals=%v,value=%v,field=%v,opt=%v)", vals, value, field, opt)
	for i, s := range vals {
		err := setWithProperType(s, value.Index(i), field, opt)
		if err != nil {
			logritelog.Default().Printf("[ERROR] Exception in setArray(vals=%v,value=%v,field=%v,opt=%v): %v", vals, value, field, opt, err)
			return err
		}
	}
	logritelog.Default().Printf("[WARN] <<< Exiting setArray(vals=%v,value=%v,field=%v,opt=%v)", vals, value, field, opt)
	return nil
}

func setSlice(vals []string, value reflect.Value, field reflect.StructField, opt setOptions) error {
	logritelog.Default().Printf("[ERROR] >>> Entering setSlice(vals=%v,value=%v,field=%v,opt=%v)", vals, value, field, opt)
	slice := reflect.MakeSlice(value.Type(), len(vals), len(vals))
	err := setArray(vals, slice, field, opt)
	if err != nil {
		logritelog.Default().Printf("[ERROR] Exception in setSlice(vals=%v,value=%v,field=%v,opt=%v): %v", vals, value, field, opt, err)
		return err
	}
	value.Set(slice)
	logritelog.Default().Printf("[WARN] <<< Exiting setSlice(vals=%v,value=%v,field=%v,opt=%v)", vals, value, field, opt)
	return nil
}

func setTimeDuration(val string, value reflect.Value) error {
	logritelog.Default().Printf("[ERROR] >>> Entering setTimeDuration(val=%v,value=%v)", val, value)
	if val == "" {
		val = "0"
	}

	d, err := time.ParseDuration(val)
	if err != nil {
		logritelog.Default().Printf("[ERROR] Exception in setTimeDuration(val=%v,value=%v): %v", val, value, err)
		return err
	}
	value.Set(reflect.ValueOf(d))
	logritelog.Default().Printf("[WARN] <<< Exiting setTimeDuration(val=%v,value=%v)", val, value)
	return nil
}

func head(str, sep string) (head string, tail string) {
	logritelog.Default().Printf("[ERROR] >>> Entering head(str=%v,sep=%v)", str, sep)
	head, tail, _ = strings.Cut(str, sep)
	logritelog.Default().Printf("[WARN] <<< Exiting head(str=%v,sep=%v)", str, sep)
	return head, tail
}

func setFormMap(ptr any, form map[string][]string) error {
	logritelog.Default().Printf("[ERROR] >>> Entering setFormMap(ptr=%v,form=%v)", ptr, form)
	el := reflect.TypeOf(ptr).Elem()

	if el.Kind() == reflect.Slice {
		ptrMap, ok := ptr.(map[string][]string)
		if !ok {
			return ErrConvertMapStringSlice
		}
		maps.Copy(ptrMap, form)

		return nil
	}

	ptrMap, ok := ptr.(map[string]string)
	if !ok {
		return ErrConvertToMapString
	}
	for k, v := range form {
		ptrMap[k] = v[len(v)-1] // pick last
	}

	logritelog.Default().Printf("[WARN] <<< Exiting setFormMap(ptr=%v,form=%v)", ptr, form)
	return nil
}
