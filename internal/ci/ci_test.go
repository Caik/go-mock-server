package ci

import (
	"testing"
)

// Test interfaces and implementations for testing
type TestInterface interface {
	GetValue() string
}

type TestImpl struct {
	value string
}

func (t *TestImpl) GetValue() string {
	return t.value
}

// Another test interface for different tests
type TestInterface2 interface {
	GetName() string
}

type TestImpl2 struct {
	name string
}

func (t *TestImpl2) GetName() string {
	return t.name
}

// Types for circular dependency testing
type CircularA struct {
	B *CircularB
}

type CircularB struct {
	A *CircularA
}

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

func TestAddFunctionHandlesDuplicateRegistrations(t *testing.T) {
	type TestStruct struct {
		Value string
	}

	// First registration should succeed
	err := Add(func() *TestStruct {
		return &TestStruct{Value: "first"}
	})

	if err != nil {
		t.Fatalf("expected no error on first registration, got %v", err)
	}

	// Second registration should fail (duplicate)
	err = Add(func() *TestStruct {
		return &TestStruct{Value: "second"}
	})

	if err == nil {
		t.Fatal("expected error on duplicate registration, got none")
	}
}

func TestAddFunctionWithOptions(t *testing.T) {
	// Test adding with dig options
	err := Add(func() TestInterface {
		return &TestImpl{value: "test"}
	})

	if err != nil {
		t.Fatalf("expected no error when adding with interface, got %v", err)
	}

	var result TestInterface
	err = Get(&result)

	if err != nil {
		t.Fatalf("expected no error when retrieving interface, got %v", err)
	}

	if result == nil {
		t.Fatal("expected result to be non-nil")
	}

	if result.GetValue() != "test" {
		t.Errorf("expected value 'test', got '%s'", result.GetValue())
	}
}

func TestInvokeFunctionWithDependencies(t *testing.T) {
	type Dependency struct {
		Name string
	}

	type Service struct {
		Dep *Dependency
	}

	// Add dependency first
	err := Add(func() *Dependency {
		return &Dependency{Name: "test-dependency"}
	})

	if err != nil {
		t.Fatalf("expected no error when adding dependency, got %v", err)
	}

	// Add service that depends on dependency
	err = Add(func(dep *Dependency) *Service {
		return &Service{Dep: dep}
	})

	if err != nil {
		t.Fatalf("expected no error when adding service, got %v", err)
	}

	// Invoke function that uses both
	err = Invoke(func(service *Service, dep *Dependency) {
		if service == nil {
			t.Fatal("expected service to be non-nil")
		}

		if dep == nil {
			t.Fatal("expected dependency to be non-nil")
		}

		if service.Dep != dep {
			t.Error("expected service to have the same dependency instance")
		}

		if dep.Name != "test-dependency" {
			t.Errorf("expected dependency name 'test-dependency', got '%s'", dep.Name)
		}
	})

	if err != nil {
		t.Fatalf("expected no error when invoking with dependencies, got %v", err)
	}
}

func TestInvokeFunctionHandlesMissingDependency(t *testing.T) {
	type MissingDependency struct{}

	err := Invoke(func(missing *MissingDependency) {
		t.Fatal("this should not be called due to missing dependency")
	})

	if err == nil {
		t.Fatal("expected error when invoking with missing dependency, got none")
	}
}

func TestInvokeFunctionHandlesNilFunction(t *testing.T) {
	err := Invoke(nil)

	if err == nil {
		t.Fatal("expected error when invoking nil function, got none")
	}
}

func TestGetFunctionRetrievesValue(t *testing.T) {
	type TestStruct struct {
		ID   int
		Name string
	}

	// Add constructor
	err := Add(func() *TestStruct {
		return &TestStruct{ID: 123, Name: "test"}
	})

	if err != nil {
		t.Fatalf("expected no error when adding constructor, got %v", err)
	}

	// Get value using Get function
	var result *TestStruct
	err = Get(&result)

	if err != nil {
		t.Fatalf("expected no error when getting value, got %v", err)
	}

	if result == nil {
		t.Fatal("expected result to be non-nil")
	}

	if result.ID != 123 {
		t.Errorf("expected ID 123, got %d", result.ID)
	}

	if result.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", result.Name)
	}
}

func TestGetFunctionHandlesMissingDependency(t *testing.T) {
	type MissingStruct struct{}

	var result *MissingStruct
	err := Get(&result)

	if err == nil {
		t.Fatal("expected error when getting missing dependency, got none")
	}

	if result != nil {
		t.Error("expected result to remain nil when error occurs")
	}
}

func TestGetFunctionHandlesPointerFallback(t *testing.T) {
	type TestStruct struct {
		Value string
	}

	// Add constructor that returns a pointer
	err := Add(func() *TestStruct {
		return &TestStruct{Value: "pointer-test"}
	})

	if err != nil {
		t.Fatalf("expected no error when adding pointer constructor, got %v", err)
	}

	// Get value - should work with pointer fallback
	var result *TestStruct
	err = Get(&result)

	if err != nil {
		t.Fatalf("expected no error when getting pointer value, got %v", err)
	}

	if result == nil {
		t.Fatal("expected result to be non-nil")
	}

	if result.Value != "pointer-test" {
		t.Errorf("expected value 'pointer-test', got '%s'", result.Value)
	}
}

func TestGetFunctionHandlesValueTypes(t *testing.T) {
	type TestStruct struct {
		Value string
	}

	// Add constructor that returns a value (not pointer)
	err := Add(func() TestStruct {
		return TestStruct{Value: "value-test"}
	})

	if err != nil {
		t.Fatalf("expected no error when adding value constructor, got %v", err)
	}

	// Get value - should work with value types
	var result TestStruct
	err = Get(&result)

	if err != nil {
		t.Fatalf("expected no error when getting value type, got %v", err)
	}

	if result.Value != "value-test" {
		t.Errorf("expected value 'value-test', got '%s'", result.Value)
	}
}

func TestGetFunctionHandlesInterfaces(t *testing.T) {
	// Add constructor that returns interface implementation
	err := Add(func() TestInterface2 {
		return &TestImpl2{name: "interface-test"}
	})

	if err != nil {
		t.Fatalf("expected no error when adding interface constructor, got %v", err)
	}

	// Get interface value
	var result TestInterface2
	err = Get(&result)

	if err != nil {
		t.Fatalf("expected no error when getting interface value, got %v", err)
	}

	if result == nil {
		t.Fatal("expected result to be non-nil")
	}

	if result.GetName() != "interface-test" {
		t.Errorf("expected name 'interface-test', got '%s'", result.GetName())
	}
}

func TestErrorHandlingAndVisualization(t *testing.T) {
	// Add first constructor
	err := Add(func(b *CircularB) *CircularA {
		return &CircularA{B: b}
	})

	if err != nil {
		t.Fatalf("expected no error when adding CircularA constructor, got %v", err)
	}

	// Try to add second constructor that creates circular dependency
	// This should fail at registration time with dig
	err = Add(func(a *CircularA) *CircularB {
		return &CircularB{A: a}
	})

	if err == nil {
		t.Fatal("expected error when adding circular dependency, got none")
	}

	// The error should be a dig error indicating a cycle
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}

	// Test that we can still use the container for other things
	type SimpleStruct struct {
		Value string
	}

	err = Add(func() *SimpleStruct {
		return &SimpleStruct{Value: "test"}
	})

	if err != nil {
		t.Fatalf("expected no error when adding simple constructor after cycle error, got %v", err)
	}

	var result *SimpleStruct
	err = Get(&result)

	if err != nil {
		t.Fatalf("expected no error when getting simple struct, got %v", err)
	}

	if result == nil || result.Value != "test" {
		t.Error("container should still work after cycle detection")
	}
}

func TestComplexDependencyGraph(t *testing.T) {
	type Database struct {
		ConnectionString string
	}

	type Logger struct {
		Level string
	}

	type UserService struct {
		DB     *Database
		Logger *Logger
	}

	type OrderService struct {
		DB         *Database
		Logger     *Logger
		UserSvc    *UserService
	}

	type APIController struct {
		UserSvc  *UserService
		OrderSvc *OrderService
		Logger   *Logger
	}

	// Add all dependencies in order
	err := Add(func() *Database {
		return &Database{ConnectionString: "test-db"}
	})
	if err != nil {
		t.Fatalf("expected no error adding Database, got %v", err)
	}

	err = Add(func() *Logger {
		return &Logger{Level: "INFO"}
	})
	if err != nil {
		t.Fatalf("expected no error adding Logger, got %v", err)
	}

	err = Add(func(db *Database, logger *Logger) *UserService {
		return &UserService{DB: db, Logger: logger}
	})
	if err != nil {
		t.Fatalf("expected no error adding UserService, got %v", err)
	}

	err = Add(func(db *Database, logger *Logger, userSvc *UserService) *OrderService {
		return &OrderService{DB: db, Logger: logger, UserSvc: userSvc}
	})
	if err != nil {
		t.Fatalf("expected no error adding OrderService, got %v", err)
	}

	err = Add(func(userSvc *UserService, orderSvc *OrderService, logger *Logger) *APIController {
		return &APIController{UserSvc: userSvc, OrderSvc: orderSvc, Logger: logger}
	})
	if err != nil {
		t.Fatalf("expected no error adding APIController, got %v", err)
	}

	// Test that complex dependency graph resolves correctly
	err = Invoke(func(controller *APIController) {
		if controller == nil {
			t.Fatal("expected controller to be non-nil")
		}

		if controller.UserSvc == nil {
			t.Fatal("expected UserService to be non-nil")
		}

		if controller.OrderSvc == nil {
			t.Fatal("expected OrderService to be non-nil")
		}

		if controller.Logger == nil {
			t.Fatal("expected Logger to be non-nil")
		}

		// Verify that the same instances are shared
		if controller.UserSvc.DB != controller.OrderSvc.DB {
			t.Error("expected same Database instance to be shared")
		}

		if controller.UserSvc.Logger != controller.OrderSvc.Logger {
			t.Error("expected same Logger instance to be shared")
		}

		if controller.OrderSvc.UserSvc != controller.UserSvc {
			t.Error("expected same UserService instance to be shared")
		}

		// Verify values
		if controller.UserSvc.DB.ConnectionString != "test-db" {
			t.Errorf("expected connection string 'test-db', got '%s'", controller.UserSvc.DB.ConnectionString)
		}

		if controller.Logger.Level != "INFO" {
			t.Errorf("expected log level 'INFO', got '%s'", controller.Logger.Level)
		}
	})

	if err != nil {
		t.Fatalf("expected no error when invoking complex dependency graph, got %v", err)
	}
}

func TestConcurrentAccess(t *testing.T) {
	type ConcurrentStruct struct {
		ID int
	}

	// Add constructor
	err := Add(func() *ConcurrentStruct {
		return &ConcurrentStruct{ID: 42}
	})

	if err != nil {
		t.Fatalf("expected no error when adding constructor, got %v", err)
	}

	// Test concurrent access to the container
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			var result *ConcurrentStruct
			err := Get(&result)

			if err != nil {
				t.Errorf("goroutine %d: expected no error, got %v", id, err)
				return
			}

			if result == nil {
				t.Errorf("goroutine %d: expected result to be non-nil", id)
				return
			}

			if result.ID != 42 {
				t.Errorf("goroutine %d: expected ID 42, got %d", id, result.ID)
				return
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
