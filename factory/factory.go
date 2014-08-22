package factory

import (
	"fmt"
	"reflect"
	"errors"
)

var factories = make(map[reflect.Type]map[string]interface{})

func checkParams(v reflect.Value, params map[string]interface{}) error {
	for field, value := range params {
		fieldV := v.FieldByName(field)
		if !fieldV.IsValid() {
			return errors.New(fmt.Sprintf("Invalid field %s.", field))
		}
		valueV := reflect.ValueOf(value)
		valueT := valueV.Type()
		if valueV.Kind() == reflect.Func {
			funcT := valueV.Type()
			if funcT.NumIn() != 0 {
				return errors.New("Function should not take any arguments.")
			}
			if funcT.NumOut() != 1 {
				return errors.New("Function should only return one value.")
			}
			valueT = funcT.Out(0)
		}
		if !valueT.AssignableTo(fieldV.Type()) {
			return errors.New(fmt.Sprintf("Value %+v for field %s is invalid.", value, field))
		}
	}
	return nil
}

func execParams(v reflect.Value, params map[string]interface{}) {
	for field, value := range params {
		fieldV := v.FieldByName(field)
		valueV := reflect.ValueOf(value)
		if valueV.Kind() == reflect.Func {
			evaluated := valueV.Call([]reflect.Value{})
			valueV = evaluated[0]
		}
		fieldV.Set(valueV)
	}
}

func parseArgs(i interface{}, options []interface{}) (reflect.Value, map[string]interface{}, error) {
	p := reflect.ValueOf(i)
	if p.Kind() != reflect.Ptr {
		return reflect.Value{}, nil, errors.New("Not a pointer.")
	}
	
	v := reflect.Indirect(p)
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, nil, errors.New("Does not point to a struct.")
	}
	
	if len(options) > 1 {
		return reflect.Value{}, nil, errors.New("Too many options.")
	}
	
	defaultParams := factories[v.Type()]
	if defaultParams == nil {
		defaultParams = make(map[string]interface{})
	}
	
	if len(options) == 1 {
		params, ok := options[0].(map[string]interface{})
		if !ok {
			return reflect.Value{}, nil, errors.New("Options are not map[string]interface{}.")
		}
		if err := checkParams(v, params); err != nil {
			return reflect.Value{}, nil, err
		}
		for key, val := range params {
			defaultParams[key] = val
		}
	}
	return v, defaultParams, nil
}

func Build(i interface{}, options ...interface{}) (interface{}, error) {
	v, params, err := parseArgs(i, options)
	if err != nil {
		return nil, err
	}
	execParams(v, params)
	return i, nil
}

func BuildMany(i interface{}, n int, options ...interface{}) ([]interface{}, error) {
	v, params, err := parseArgs(i, options)
	if err != nil {
		return nil, err
	}
	arr := make([]interface{}, n)
	for k, _ := range arr {
		arr[k] = reflect.New(v.Type()).Interface()
		execParams(reflect.Indirect(reflect.ValueOf(arr[k])), params)
	}
	return arr, nil
}

func MustBuildMany(i interface{}, n int, options ...interface{}) []interface{} {
	v, params, err := parseArgs(i, options)
	if err != nil {
		return []interface{}{}
	}
	arr := make([]interface{}, n)
	for k, _ := range arr {
		arr[k] = reflect.New(v.Type()).Interface()
		execParams(reflect.Indirect(reflect.ValueOf(arr[k])), params)
	}
	return arr
}

func MustBuild(i interface{}, options ...interface{}) interface{} {
	if v, params, err := parseArgs(i, options); err == nil {
		execParams(v, params)
	}
	return i
}

func Register(i interface{}, params map[string]interface{}) error {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Struct {
		return errors.New("Not a struct.")
	}
	if err := checkParams(v, params); err != nil {
		return err
	}
	factories[v.Type()] = params
	return nil
}
