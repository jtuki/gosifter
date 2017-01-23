package api

import (
	"testing"
	"fmt"
	"encoding/json"
)

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


	fmt.Println("=== S1 ===")
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

	// 原始的 json marshal 结果
	if originalJson, err := json.Marshal(s1); err != nil {
		t.Fatal(err)
	} else {
		fmt.Printf("original json s1: %s\n", originalJson)
	}

	// 筛过之后的 json marshal 结果
	if siftedS1, err := SiftStruct(s1); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println("siftedS1:", siftedS1)
		if jsonBytes, err := json.Marshal(siftedS1); err != nil {
			t.Fatal(err)
		} else {
			fmt.Printf("siftedS1 json: %s\n", jsonBytes)
		}
	}

	fmt.Println("=== S2 ===")
	s2 := S2{
		S21: "S21_value",
		S22: int32(1666),
		S23: float64(1666.666),
		S24: T(1666),
		S25: float64(2666.666),
	}
	s2.S1 = s1

	// 原始的 json marshal 结果
	if originalJson, err := json.Marshal(s2); err != nil {
		t.Fatal(err)
	} else {
		fmt.Printf("original json s2: %s\n", originalJson)
	}
	
	// 筛过之后的 json marshal 结果
	if siftedS2, err := SiftStruct(s2); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println("siftedS2:", siftedS2)
		if jsonBytes, err := json.Marshal(siftedS2); err != nil {
			t.Fatal(err)
		} else {
			fmt.Printf("siftedS2 json: %s\n", jsonBytes)
		}
	}
}
