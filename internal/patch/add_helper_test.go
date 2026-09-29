// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package patch

import "testing"

func TestConvertValue(t *testing.T) {

	type Test2 struct {
		FieldA string `json:"fieldA"`
		FieldB int    `json:"fieldB"`
	}

	type Test1 struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
		Field3 Test2  `json:"field3"`
	}

	v2, err := ConvertValue[*Test2](nil)
	if err != nil {
		t.Fatalf("ConvertValue() failed: %v", err)
	}
	if v2 != nil {
		t.Errorf("ConvertValue() = %v, want nil", v2)
	}

	v2, err = ConvertValue[*Test2](map[string]any{
		"fieldA": "valueA",
		"fieldB": 42,
	})
	if err != nil {
		t.Fatalf("ConvertValue() failed: %v", err)
	}
	if v2 == nil || v2.FieldA != "valueA" || v2.FieldB != 42 {
		t.Errorf("ConvertValue() = %v, want &{valueA 42}", v2)
	}

	v2, err = ConvertValue[*Test2](&Test1{
		Field1: "value1",
		Field2: 100,
		Field3: Test2{
			FieldA: "valueA",
			FieldB: 42,
		},
	})
	if err != nil {
		t.Fatalf("ConvertValue() failed: %v", err)
	}
	if v2 == nil || v2.FieldA != "" || v2.FieldB != 0 {
		t.Errorf("ConvertValue() = %v, want &{valueA 42}", v2)
	}

	// tests := []struct {
	// 	name string // description of this test case
	// 	// Named input parameters for target function.
	// 	value   any
	// 	want    *T
	// 	wantErr bool
	// }{
	// 	// TODO: Add test cases.
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		got, gotErr := ConvertValuePtr(tt.value)
	// 		if gotErr != nil {
	// 			if !tt.wantErr {
	// 				t.Errorf("ConvertValuePtr() failed: %v", gotErr)
	// 			}
	// 			return
	// 		}
	// 		if tt.wantErr {
	// 			t.Fatal("ConvertValuePtr() succeeded unexpectedly")
	// 		}
	// 		// TODO: update the condition below to compare got with tt.want.
	// 		if true {
	// 			t.Errorf("ConvertValuePtr() = %v, want %v", got, tt.want)
	// 		}
	// 	})
	// }
}
