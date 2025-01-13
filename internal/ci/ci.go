package ci

import (
	"errors"
	"go.uber.org/dig"
	"os"
)

var container = dig.New()

func Add(constructor interface{}, opts ...dig.ProvideOption) error {
	return container.Provide(constructor, opts...)
}

func Invoke(constructor interface{}, opts ...dig.InvokeOption) error {
	err := container.Invoke(constructor, opts...)

	if dig.CanVisualizeError(err) {
		_ = dig.Visualize(container, os.Stdout)

		return err
	}

	return err
}

func Get[T any](value *T) error {
	err := container.Invoke(func(other T) {
		*value = other
	})

	if err == nil {
		return nil
	}

	rootCause := dig.RootCause(err)
	var de dig.Error

	if !errors.As(rootCause, &de) {
		return err
	}

	// it's a dig error (probably a missing dependency), so we try to find the dependency via pointer
	return container.Invoke(func(other *T) {
		*value = *other
	})
}
