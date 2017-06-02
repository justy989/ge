package main

import (
	"regexp"
)

type BoundFunction func()
type KeyBinder interface {
	Bind(expr string, fn BoundFunction) error
	Handle(key byte) error
}

type KeyBinding struct {
	Expr     string
	Function BoundFunction
}

type regexBinder struct {
	bufferedKeys []byte
	bindings     []KeyBinding
}

func NewBinder() KeyBinder {
	return &regexBinder{}
}

func (binder *regexBinder) Bind(expr string, fn BoundFunction) error {
	binder.bindings = append(binder.bindings, KeyBinding{expr, fn})
	return nil
}

func (binder *regexBinder) Handle(key byte) error {
	var matchedBinding *KeyBinding
	binder.bufferedKeys = append(binder.bufferedKeys, key)
	for i, binding := range binder.bindings {
		for i := range binding.Expr {
			matched, _ := regexp.Match(binding.Expr, binder.bufferedKeys)
			if matched && matchedBinding != nil {
				// we matched two bindings
				return nil
			} else if matched {
				matchedBinding = &binder.bindings[i]
			}
		}
	}

	if matchedBinding != nil {
		// call our matched binding
		binder.bufferedKeys = []byte{}
		matchedBinding.Function()
		return nil
	}

	return nil
}
