package errors

import (
	"errors"
	"testing"
)

func TestFieldError_Error(t *testing.T) {
	type testCase struct {
		name string
		err  error
		want string
	}

	leaf := errors.New("boom")

	cases := []testCase{
		{
			name: "simple field",
			err:  NewField("foo", leaf),
			want: "foo: boom",
		},
		{
			name: "nested fields",
			err:  NewField("foo", NewField("bar", leaf)),
			want: "foo.bar: boom",
		},
		{
			name: "deeply nested fields",
			err:  NewField("a", NewField("b", NewField("c", leaf))),
			want: "a.b.c: boom",
		},
		{
			name: "nil error returns nil",
			err:  NewField("should_not_exist", nil),
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil && tc.want != "" {
				t.Errorf("got nil, want %q", tc.want)
			} else if tc.err != nil {
				got := tc.err.Error()
				if got != tc.want {
					t.Errorf("got %q, want %q", got, tc.want)
				}
			}
		})
	}
}

func TestIndexError_Error(t *testing.T) {
	type testCase struct {
		name string
		err  error
		want string
	}

	leaf := errors.New("fail")

	cases := []testCase{
		{
			name: "simple index",
			err:  NewIndex(3, leaf),
			want: "[3]: fail",
		},
		{
			name: "nested index",
			err:  NewIndex(2, NewIndex(1, leaf)),
			want: "[2][1]: fail",
		},
		{
			name: "index wraps field",
			err:  NewIndex(1, NewField("foo", leaf)),
			want: "[1].foo: fail",
		},
		{
			name: "field wraps index",
			err:  NewField("bar", NewIndex(7, leaf)),
			want: "bar[7]: fail",
		},
		{
			name: "mixed deep nesting",
			err:  NewField("top", NewIndex(2, NewField("mid", NewIndex(4, leaf)))),
			want: "top[2].mid[4]: fail",
		},
		{
			name: "nil error returns nil",
			err:  NewIndex(99, nil),
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil && tc.want != "" {
				t.Errorf("got nil, want %q", tc.want)
			} else if tc.err != nil {
				got := tc.err.Error()
				if got != tc.want {
					t.Errorf("got %q, want %q", got, tc.want)
				}
			}
		})
	}
}

func TestComplexNesting(t *testing.T) {
	leaf := errors.New("bad")

	// field[3].subfield[2].leaf: bad
	err := NewField("field", NewIndex(3, NewField("subfield", NewIndex(2, NewField("leaf", leaf)))))
	want := "field[3].subfield[2].leaf: bad"
	if got := err.Error(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// [0][1][2]: bad
	err2 := NewIndex(0, NewIndex(1, NewIndex(2, leaf)))
	want2 := "[0][1][2]: bad"
	if got := err2.Error(); got != want2 {
		t.Errorf("got %q, want %q", got, want2)
	}
}

func TestErrorUnwrap(t *testing.T) {
	leaf := errors.New("root")
	err := NewField("foo", NewIndex(1, NewField("bar", leaf)))

	// Walk chain and check inner errors
	var found error
	un := err
	for un != nil {
		switch e := un.(type) {
		case *FieldError:
			un = e.Err
		case *IndexError:
			un = e.Err
		default:
			found = un
			un = nil
		}
	}
	if found != leaf {
		t.Errorf("did not unwrap to leaf, got %#v", found)
	}
}

func TestOptionError_Error(t *testing.T) {
	type testCase struct {
		name string
		err  error
		want string
	}

	leaf := errors.New("fail")

	cases := []testCase{
		{
			name: "simple option",
			err:  NewOption("maybeField", leaf),
			want: "?maybeField: fail",
		},
		{
			name: "option wrapping field",
			err:  NewOption("maybeField", NewField("foo", leaf)),
			want: "?maybeField.foo: fail",
		},
		{
			name: "option wrapping index",
			err:  NewOption("maybeField", NewIndex(2, leaf)),
			want: "?maybeField[2]: fail",
		},
		{
			name: "nested options",
			err:  NewOption("outer", NewOption("inner", leaf)),
			want: "?outer.?inner: fail",
		},
		{
			name: "option wraps field wraps index",
			err:  NewOption("maybeField", NewField("foo", NewIndex(1, leaf))),
			want: "?maybeField.foo[1]: fail",
		},
		{
			name: "nil error returns nil",
			err:  NewOption("maybe", nil),
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil && tc.want != "" {
				t.Errorf("got nil, want %q", tc.want)
			} else if tc.err != nil {
				got := tc.err.Error()
				if got != tc.want {
					t.Errorf("got %q, want %q", got, tc.want)
				}
			}
		})
	}
}

func TestComplexOptionNesting(t *testing.T) {
	leaf := errors.New("zero")
	// foo[4].?bar[7].baz: zero
	err := NewField("foo",
		NewIndex(4,
			NewOption("bar",
				NewIndex(7,
					NewField("baz", leaf),
				),
			),
		),
	)
	want := "foo[4].?bar[7].baz: zero"
	if got := err.Error(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
