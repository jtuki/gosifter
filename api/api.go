package api

import (
	"encoding/json"
	"fmt"
	gosifter "github.com/jtuki/gosifter/src"
	"reflect"
)

// api function
//
// @param
//  s - 需要执行筛选/脱敏的结构体对象（或者其指针）
//  clevel - 最高允许的安全等级（高于此等级的将被筛除）
func SiftStruct(s interface{}, clevel int) (map[string]interface{}, error) {
	isPtr := false // s是否是指针类型
	rt := reflect.TypeOf(s)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		isPtr = true
	}

	if rt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("invalid param type %v", rt.Kind())
	}

	if cs, err := gosifter.GetSifter(rt); err != nil {
		return nil, err
	} else {
		if !isPtr {
			return cs.SiftStruct(s, clevel)
		} else { // pointer type to struct
			return cs.SiftStruct(reflect.ValueOf(s).Elem().Interface(), clevel)
		}
	}
}

// api function
//
// 封装的序列化操作，返回序列化之后的结果和可能的错误。
func Marshal(s interface{}, clevel int) ([]byte, error) {
	if clevel == CONFIDENTIAL_LEVEL_MAX {
		return json.Marshal(s)
	}
	m, err := SiftStruct(s, clevel)
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}
