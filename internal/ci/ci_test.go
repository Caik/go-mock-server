package ci

import (
	"testing"
)

func TestAddFunctionRegistersConstructor(t *testing.T) {
	type TestStruct struct{}

	err := Add(func() *TestStruct {
		return &TestStruct{}
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var result *TestStruct
	err = Get(&result)

	if err != nil {
		t.Fatalf("expected no error when retrieving, got %v", err)
	}

	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
}

func TestInvokeFunctionExecutesConstructor(t *testing.T) {
	type TestStruct struct{}

	err := Add(func() *TestStruct {
		return &TestStruct{}
	})

	if err != nil {
		t.Fatalf("expected no error when adding constructor, got %v", err)
	}

	err = Invoke(func(ts *TestStruct) {
		if ts == nil {
			t.Fatal("expected TestStruct to be non-nil")
		}
	})

	if err != nil {
		t.Fatalf("expected no error when invoking constructor, got %v", err)
	}
}

func TestAddFunctionHandlesNilConstructor(t *testing.T) {
	err := Add(nil)

	if err == nil {
		t.Fatal("expected an error when adding a nil constructor, got none")
	}
}
