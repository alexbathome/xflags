package xflags

import (
	"fmt"
	"strconv"
	"time"
)

// Value is the interface to the dynamic value stored in a flag.
// (The default value is represented as a string.)
//
// Set is called once, in command line order, for each flag present.
type Value interface {
	// String() string
	Set(s string) error
}

// BoolValue is an optional interface to indicate boolean flags that can be
// supplied without a "=value" argument.
type BoolValue interface {
	Value
	IsBoolFlag() bool
}

func isBoolValue(v Value) bool {
	if bv, ok := v.(BoolValue); ok {
		return bv.IsBoolFlag()
	}
	return false
}

var _ BoolValue = (*genericValue[bool])(nil)

// ValidateFunc is a function that validates an argument before it is parsed.
type ValidateFunc = func(arg string) error

type funcValue func(string) error

func (f funcValue) Set(s string) error { return f(s) }

type cliGenericTypes interface {
	string | uint | uint64 | int | int64 | bool | time.Duration | float64
}

// genericSliceValue is a generic struct that describes a
// slice of generic type T
type genericSliceValue[T cliGenericTypes] struct {
	p   *[]T
	hot bool
}

func newGenericSlice[T cliGenericTypes](val []T, p *[]T) *genericSliceValue[T] {
	*p = val
	return &genericSliceValue[T]{p: p}
}

func (p *genericSliceValue[T]) String() string {
	return fmt.Sprintf("%v", *p.p)
}

func (p *genericSliceValue[T]) Get() interface{} { return *p.p }

func (p *genericSliceValue[T]) Set(s string) error {
	if !p.hot {
		*p.p = make([]T, 0, 1)
		p.hot = true
	}
	*p.p = append(*p.p, any(s).(T))
	return nil
}

type genericValue[T cliGenericTypes] struct {
	p      *T
	isBool bool
}

func newGeneric[T cliGenericTypes](val T, p *T, isbool bool) *genericValue[T] {
	*p = val
	return &genericValue[T]{
		p:      p,
		isBool: isbool,
	}
}

func (p *genericValue[T]) IsBoolFlag() bool { return p.isBool }

func (p *genericValue[T]) String() string {
	switch v := any(p.p).(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case int:
		return strconv.FormatInt(int64(v), 10)
	case bool:
		return strconv.FormatBool(v)
	case time.Duration:
		return fmt.Sprint(v)
	default:
		return ""
	}
}

func (p *genericValue[T]) Get() T {
	return *p.p
}

func (p *genericValue[T]) Set(s string) error {
	switch v := any(*p.p).(type) {
	case string:
		*p.p = any(s).(T)
	case float64:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		*p.p = any(float64(v)).(T)
	case uint64:
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		*p.p = any(uint(v)).(T)
	case uint:
		o, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return err
		}
		*p.p = any(uint(o)).(T)
	case int64:
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*p.p = any(int64(v)).(T)
	case int:
		o, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return err
		}
		*p.p = any(int(o)).(T)
	case bool:
		v, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		*p.p = any(bool(v)).(T)
	case time.Duration:
		v, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		*p.p = any(time.Duration(v)).(T)
	default:
		return fmt.Errorf("unsupported type: %T", *p.p)
	}
	return nil
}
