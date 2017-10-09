package types

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Errors/Exceptions
type LGError struct {
	Obj Top
}

func (e LGError) Error() string {
	return fmt.Sprintf("%#v", e.Obj)
}

// General types
type Top interface {
}

type EnvType interface {
	Find(key Symbol) EnvType
	Set(key Symbol, value Top) Top
	Get(key Symbol) (Top, error)
}

func IsNil(obj Top) bool {
	return obj == nil
}

func IsTrue(obj Top) bool {
	b, ok := obj.(bool)
	return ok && b == true
}

func IsFalse(obj Top) bool {
	b, ok := obj.(bool)
	return ok && b == false
}

// Symbols
type Symbol struct {
	Val string
}

func IsSymbol(obj Top) bool {
	_, ok := obj.(Symbol)
	return ok
}

func NewKeyword(s string) (Top, error) {
	return "\u029e" + s, nil
}

func IsKeyword(obj Top) bool {
	s, ok := obj.(string)
	return ok && strings.HasPrefix(s, "\u029e")
}

func IsString(obj Top) bool {
	_, ok := obj.(string)
	return ok
}

type Func struct {
	Fn   func([]Top) (Top, error)
	Meta Top
}

func IsFunc(obj Top) bool {
	_, ok := obj.(Func)
	return ok
}

type MalFunc struct {
	Eval    func(Top, EnvType) (Top, error)
	Exp     Top
	Env     EnvType
	Params  Top
	IsMacro bool
	GenEnv  func(EnvType, Top, Top) (EnvType, error)
	Meta    Top
}

func MalFunc_Q(obj Top) bool {
	_, ok := obj.(MalFunc)
	return ok
}

func (f MalFunc) SetMacro() Top {
	f.IsMacro = true
	return f
}

func (f MalFunc) GetMacro() bool {
	return f.IsMacro
}

// Take either a MalFunc or regular function and apply it to the
// arguments
func Apply(f Top, a []Top) (Top, error) {
	switch f := f.(type) {
	case MalFunc:
		env, e := f.GenEnv(f.Env, f.Params, List{a, nil})
		if e != nil {
			return nil, e
		}
		return f.Eval(f.Exp, env)
	case Func:
		return f.Fn(a)
	case func([]Top) (Top, error):
		return f(a)
	default:
		return nil, errors.New("Invalid function to Apply")
	}
}

type List struct {
	Val  []Top
	Meta Top
}

func NewList(a ...Top) Top {
	return List{Val: a, Meta: nil}
}

func IsList(obj Top) bool {
	_, ok := obj.(List)
	return ok
}

// Vectors
type Vector struct {
	Val  []Top
	Meta Top
}

func IsVector(obj Top) bool {
	_, ok := obj.(Vector)
	return ok
}

func GetSlice(seq Top) ([]Top, error) {
	switch obj := seq.(type) {
	case List:
		return obj.Val, nil
	case Vector:
		return obj.Val, nil
	default:
		return nil, errors.New("GetSlice called on non-sequence")
	}
}

// Hash Maps
type HashMap struct {
	Val  map[string]Top
	Meta Top
}

func NewHashMap(seq Top) (Top, error) {
	lst, e := GetSlice(seq)
	if e != nil {
		return nil, e
	}
	if len(lst)%2 == 1 {
		return nil, errors.New("Odd number of arguments to NewHashMap")
	}
	m := map[string]Top{}
	for i := 0; i < len(lst); i += 2 {
		str, ok := lst[i].(string)
		if !ok {
			return nil, errors.New("expected hash-map key string")
		}
		m[str] = lst[i+1]
	}
	return HashMap{m, nil}, nil
}

func IsHashMap(obj Top) bool {
	_, ok := obj.(HashMap)
	return ok
}

// Atoms
type Atom struct {
	Val  Top
	Meta Top
}

func (a *Atom) Set(val Top) Top {
	a.Val = val
	return a
}

func IsAtom(obj Top) bool {
	_, ok := obj.(*Atom)
	return ok
}

// General functions

func objType(obj Top) string {
	if obj == nil {
		return "nil"
	}
	return reflect.TypeOf(obj).Name()
}

func IsSeq(seq Top) bool {
	if seq == nil {
		return false
	}
	return (reflect.TypeOf(seq).Name() == "List") ||
		(reflect.TypeOf(seq).Name() == "Vector")
}

func Eq(a Top, b Top) bool {
	ota := reflect.TypeOf(a)
	otb := reflect.TypeOf(b)
	if !((ota == otb) || (IsSeq(a) && IsSeq(b))) {
		return false
	}
	switch a.(type) {
	case Symbol:
		return a.(Symbol).Val == b.(Symbol).Val
	case List:
		as, _ := GetSlice(a)
		bs, _ := GetSlice(b)
		if len(as) != len(bs) {
			return false
		}
		for i := 0; i < len(as); i += 1 {
			if !Eq(as[i], bs[i]) {
				return false
			}
		}
		return true
	case Vector:
		as, _ := GetSlice(a)
		bs, _ := GetSlice(b)
		if len(as) != len(bs) {
			return false
		}
		for i := 0; i < len(as); i += 1 {
			if !Eq(as[i], bs[i]) {
				return false
			}
		}
		return true
	case HashMap:
		am := a.(HashMap).Val
		bm := b.(HashMap).Val
		if len(am) != len(bm) {
			return false
		}
		for k, v := range am {
			if !Eq(v, bm[k]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
