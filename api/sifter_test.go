package api

import (
	"testing"
	"fmt"
	"encoding/json"
)

// @param
//  name - name separator 用于输出
//  s - 原始数据存在的对象
//  us1 us2 - 用于 unmarshal 操作的对象
func compareMarshallerResult(t *testing.T, name string, s interface{}, us1, us2 interface{}) {
	fmt.Printf("=== %s ===\n", name)
	
	// 原始的 json marshal 结果
	var jsonBytes, siftBytes []byte
	var siftedMap map[string]interface{}
	var err error
	if jsonBytes, err = json.Marshal(s); err != nil {
		t.Fatal(err)
	} else {
		fmt.Printf("original json s1: %s\n", jsonBytes)
	}

	// 筛过之后的 json marshal 结果
	if siftedMap, err = SiftStruct(s); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println("siftedS1:", siftedMap)
		if siftBytes, err = json.Marshal(siftedMap); err != nil {
			t.Fatal(err)
		} else {
			fmt.Printf("siftedS1 json: %s\n", siftBytes)
		}
	}
	
	if err = json.Unmarshal(jsonBytes, us1); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(siftBytes, us2); err != nil {
		t.Fatal(err)
	}
}

func TestSiftStruct(t *testing.T) {
	type T uint64
	type S1 struct {
		// 基于结构体域的可见性原则，如下字段应该被忽略
		s11_lowercase  string `json:"s_11_lowercase"`
		s11_lowercase2 string `json:",omitempty"`
		s11_lowercase3 string `json:"s_11_lowercase,omitempty"`
		s11_lowercase4 string

		S11 string  `json:"s_11"`
		S12 int32   `json:"s_12,omitempty"`
		S13 float64 `json:"-"`
		S14 T       `json:",omitempty"`
	}
	
	type S2 struct {
		S1 `json:"s1_struct"`

		S21 string  `json:"s_21"`
		S22 int32   `json:"s_22,omitempty"`
		S23 float64 `json:"-"`
		S24 T       `json:",omitempty"`
		S25 float64 `json:"s_25"`
	}

	s1 := S1{
		s11_lowercase: "s11_lowercase",
		s11_lowercase2: "s11_lowercase2",
		s11_lowercase3: "s11_lowercase3",
		s11_lowercase4: "s11_lowercase4",
		
		S11: "S11_value",
		S12: int32(666),
		S13: float64(666.666),
		S14: T(666),
	}

	func() {
		var us1, us2 S1
		compareMarshallerResult(t, "s1", s1, &us1, &us2)
		if us1 != us2 {
			t.Fatalf("unmarshal: not the same")
		} else {
			fmt.Printf("unmarshal: us1[%+v]\n", us1)
			fmt.Printf("unmarshal: us2[%+v]\n", us2)
		}
	}()

	s2 := S2{
		S21: "S21_value",
		S22: int32(1666),
		S23: float64(1666.666),
		S24: T(1666),
		S25: float64(2666.666),
	}
	s2.S1 = s1

	func() {
		var us1, us2 S2
		compareMarshallerResult(t, "s2", s2, &us1, &us2)
		if us1 != us2 {
			t.Fatalf("unmarshal: not the same")
		} else {
			fmt.Printf("unmarshal: us1[%+v]\n", us1)
			fmt.Printf("unmarshal: us2[%+v]\n", us2)
		}
	}()
}
