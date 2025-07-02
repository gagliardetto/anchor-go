package idl

import "encoding/json"

// Option marshals to null if the value is nil, otherwise it marshals to the value.
type Option[T any] struct {
	value *T
}

func Some[T any](value T) Option[T] {
	return Option[T]{value: &value}
}

func None[T any]() Option[T] {
	return Option[T]{value: nil}
}

func (o Option[T]) IsSome() bool {
	return o.value != nil
}

func (o Option[T]) IsNone() bool {
	return o.value == nil
}

func (o Option[T]) Unwrap() T {
	if o.value == nil {
		panic("called `Option.Unwrap()` on a `None` value")
	}
	return *o.value
}

func (o Option[T]) UnwrapOr(defaultValue T) T {
	if o.value == nil {
		return defaultValue
	}
	return *o.value
}

func (o Option[T]) UnwrapOrElse(f func() T) T {
	if o.value == nil {
		return f()
	}
	return *o.value
}

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(o.value)
}

func (o *Option[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		o.value = nil
		return nil
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	o.value = &value
	return nil
}

// OptionSkip is a variant of Option that skips serialization if the value is nil.
type OptionSkip[T any] struct {
	value *T
}

func SomeSkip[T any](value T) OptionSkip[T] {
	return OptionSkip[T]{value: &value}
}

func NoneSkip[T any]() OptionSkip[T] {
	return OptionSkip[T]{value: nil}
}

func (o OptionSkip[T]) IsSome() bool {
	return o.value != nil
}

func (o OptionSkip[T]) IsNone() bool {
	return o.value == nil
}

func (o OptionSkip[T]) Unwrap() T {
	if o.value == nil {
		panic("called `Option.Unwrap()` on a `None` value")
	}
	return *o.value
}

func (o OptionSkip[T]) UnwrapOr(defaultValue T) T {
	if o.value == nil {
		return defaultValue
	}
	return *o.value
}

func (o OptionSkip[T]) UnwrapOrElse(f func() T) T {
	if o.value == nil {
		return f()
	}
	return *o.value
}

func (o OptionSkip[T]) MarshalJSON() ([]byte, error) {
	if o.value == nil {
		return nil, nil // TODO: does this work?
	}
	return json.Marshal(o.value)
}
