package errors

import "strconv"

type FieldError struct {
	Field string
	Err   error
}

func (e *FieldError) Error() string {
	if nested, ok := e.Err.(*FieldError); ok {
		return e.Field + "." + nested.Error()
	}
	if nested, ok := e.Err.(*IndexError); ok {
		return e.Field + nested.Error()
	}
	if nested, ok := e.Err.(*OptionError); ok {
		return e.Field + "." + nested.Error()
	}
	return e.Field + ": " + e.Err.Error()
}

// and utility:
func NewField(field string, err error) error {
	if err == nil {
		return nil
	}
	return &FieldError{Field: field, Err: err}
}

type IndexError struct {
	Index int
	Err   error
}

func (e *IndexError) Error() string {
	// Recursively assemble the path
	if nested, ok := e.Err.(*FieldError); ok {
		return "[" + strconv.Itoa(e.Index) + "]." + nested.Error()
	}
	if nested, ok := e.Err.(*IndexError); ok {
		return "[" + strconv.Itoa(e.Index) + "]" + nested.Error()
	}
	if nested, ok := e.Err.(*OptionError); ok {
		return "[" + strconv.Itoa(e.Index) + "]." + nested.Error()
	}
	return "[" + strconv.Itoa(e.Index) + "]: " + e.Err.Error()
}

func NewIndex(idx int, err error) error {
	if err == nil {
		return nil
	}
	return &IndexError{Index: idx, Err: err}
}

type OptionError struct {
	Field string
	Err   error
}

func (e *OptionError) Error() string {
	if nested, ok := e.Err.(*FieldError); ok {
		return "?" + e.Field + "." + nested.Error()
	}
	if nested, ok := e.Err.(*IndexError); ok {
		return "?" + e.Field + nested.Error()
	}
	if nested, ok := e.Err.(*OptionError); ok {
		return "?" + e.Field + "." + nested.Error()
	}
	return "?" + e.Field + ": " + e.Err.Error()
}

func NewOption(field string, err error) error {
	if err == nil {
		return nil
	}
	return &OptionError{Field: field, Err: err}
}
