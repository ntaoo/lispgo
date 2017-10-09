package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

import (
	"github.com/ntaoo/lispgo/printer"
	"github.com/ntaoo/lispgo/reader"
	"github.com/ntaoo/lispgo/readline"
	. "github.com/ntaoo/lispgo/types"
)

// Errors/Exceptions
func throw(a []Top) (Top, error) {
	return nil, LGError{a[0]}
}

func printStr(a []Top) (Top, error) {
	return printer.PrintList(a, true, "", "", " "), nil
}

func str(a []Top) (Top, error) {
	return printer.PrintList(a, false, "", "", ""), nil
}

func prn(a []Top) (Top, error) {
	fmt.Println(printer.PrintList(a, true, "", "", " "))
	return nil, nil
}

func printLine(a []Top) (Top, error) {
	fmt.Println(printer.PrintList(a, false, "", "", " "))
	return nil, nil
}

func slurp(a []Top) (Top, error) {
	b, e := ioutil.ReadFile(a[0].(string))
	if e != nil {
		return nil, e
	}
	return string(b), nil
}

// Number functions
func time_ms(a []Top) (Top, error) {
	return int(time.Now().UnixNano() / int64(time.Millisecond)), nil
}

func copyHashMap(hm HashMap) HashMap {
	new_hm := HashMap{map[string]Top{}, nil}
	for k, v := range hm.Val {
		new_hm.Val[k] = v
	}
	return new_hm
}

func assoc(a []Top) (Top, error) {
	if len(a) < 3 {
		return nil, errors.New("assoc requires at least 3 arguments")
	}
	if len(a)%2 != 1 {
		return nil, errors.New("assoc requires odd number of arguments")
	}
	if !IsHashMap(a[0]) {
		return nil, errors.New("assoc called on non-hash map")
	}
	newHM := copyHashMap(a[0].(HashMap))
	for i := 1; i < len(a); i += 2 {
		key := a[i]
		if !IsString(key) {
			return nil, errors.New("assoc called with non-string key")
		}
		newHM.Val[key.(string)] = a[i+1]
	}
	return newHM, nil
}

func dissoc(a []Top) (Top, error) {
	if len(a) < 2 {
		return nil, errors.New("dissoc requires at least 3 arguments")
	}
	if !IsHashMap(a[0]) {
		return nil, errors.New("dissoc called on non-hash map")
	}
	newHM := copyHashMap(a[0].(HashMap))
	for i := 1; i < len(a); i += 1 {
		key := a[i]
		if !IsString(key) {
			return nil, errors.New("dissoc called with non-string key")
		}
		delete(newHM.Val, key.(string))
	}
	return newHM, nil
}

func get(a []Top) (Top, error) {
	if len(a) != 2 {
		return nil, errors.New("get requires 2 arguments")
	}
	if IsNil(a[0]) {
		return nil, nil
	}
	if !IsHashMap(a[0]) {
		return nil, errors.New("get called on non-hash map")
	}
	if !IsString(a[1]) {
		return nil, errors.New("get called with non-string key")
	}
	return a[0].(HashMap).Val[a[1].(string)], nil
}

func contains_Q(hm Top, key Top) (Top, error) {
	if IsNil(hm) {
		return false, nil
	}
	if !IsHashMap(hm) {
		return nil, errors.New("get called on non-hash map")
	}
	if !IsString(key) {
		return nil, errors.New("get called with non-string key")
	}
	_, ok := hm.(HashMap).Val[key.(string)]
	return ok, nil
}

func keys(a []Top) (Top, error) {
	if !IsHashMap(a[0]) {
		return nil, errors.New("keys called on non-hash map")
	}
	slc := []Top{}
	for k := range a[0].(HashMap).Val {
		slc = append(slc, k)
	}
	return List{Val: slc, Meta: nil}, nil
}
func vals(a []Top) (Top, error) {
	if !IsHashMap(a[0]) {
		return nil, errors.New("keys called on non-hash map")
	}
	slc := []Top{}
	for _, v := range a[0].(HashMap).Val {
		slc = append(slc, v)
	}
	return List{slc, nil}, nil
}

func cons(a []Top) (Top, error) {
	val := a[0]
	lst, e := GetSlice(a[1])
	if e != nil {
		return nil, e
	}

	return List{Val: append([]Top{val}, lst...), Meta: nil}, nil
}

func concat(a []Top) (Top, error) {
	if len(a) == 0 {
		return List{}, nil
	}
	slc1, e := GetSlice(a[0])
	if e != nil {
		return nil, e
	}
	for i := 1; i < len(a); i += 1 {
		slc2, e := GetSlice(a[i])
		if e != nil {
			return nil, e
		}
		slc1 = append(slc1, slc2...)
	}
	return List{Val: slc1, Meta: nil}, nil
}

func nth(a []Top) (Top, error) {
	slc, e := GetSlice(a[0])
	if e != nil {
		return nil, e
	}
	idx := a[1].(int)
	if idx < len(slc) {
		return slc[idx], nil
	} else {
		return nil, errors.New("nth: index out of range")
	}
}

func first(a []Top) (Top, error) {
	if len(a) == 0 {
		return nil, nil
	}
	if a[0] == nil {
		return nil, nil
	}
	slc, e := GetSlice(a[0])
	if e != nil {
		return nil, e
	}
	if len(slc) == 0 {
		return nil, nil
	}
	return slc[0], nil
}

func rest(a []Top) (Top, error) {
	if a[0] == nil {
		return List{}, nil
	}
	slc, e := GetSlice(a[0])
	if e != nil {
		return nil, e
	}
	if len(slc) == 0 {
		return List{}, nil
	}
	return List{Val: slc[1:], Meta: nil}, nil
}

func isEmpty(a []Top) (Top, error) {
	switch obj := a[0].(type) {
	case List:
		return len(obj.Val) == 0, nil
	case Vector:
		return len(obj.Val) == 0, nil
	case nil:
		return true, nil
	default:
		return nil, errors.New("empty? called on non-sequence")
	}
}

func count(a []Top) (Top, error) {
	switch obj := a[0].(type) {
	case List:
		return len(obj.Val), nil
	case Vector:
		return len(obj.Val), nil
	case map[string]Top:
		return len(obj), nil
	case nil:
		return 0, nil
	default:
		return nil, errors.New("count called on non-sequence")
	}
}

func apply(a []Top) (Top, error) {
	if len(a) < 2 {
		return nil, errors.New("apply requires at least 2 args")
	}
	f := a[0]
	args := []Top{}
	for _, b := range a[1 : len(a)-1] {
		args = append(args, b)
	}
	last, e := GetSlice(a[len(a)-1])
	if e != nil {
		return nil, e
	}
	args = append(args, last...)
	return Apply(f, args)
}

func mapFunc(a []Top) (Top, error) {
	if len(a) != 2 {
		return nil, errors.New("map requires 2 args")
	}
	f := a[0]
	results := []Top{}
	args, e := GetSlice(a[1])
	if e != nil {
		return nil, e
	}
	for _, arg := range args {
		res, e := Apply(f, []Top{arg})
		results = append(results, res)
		if e != nil {
			return nil, e
		}
	}
	return List{Val: results, Meta: nil}, nil
}

func conj(a []Top) (Top, error) {
	if len(a) < 2 {
		return nil, errors.New("conj requires at least 2 arguments")
	}
	switch seq := a[0].(type) {
	case List:
		new_slc := []Top{}
		for i := len(a) - 1; i > 0; i -= 1 {
			new_slc = append(new_slc, a[i])
		}
		return List{append(new_slc, seq.Val...), nil}, nil
	case Vector:
		new_slc := seq.Val
		for _, x := range a[1:] {
			new_slc = append(new_slc, x)
		}
		return Vector{new_slc, nil}, nil
	}

	if !IsHashMap(a[0]) {
		return nil, errors.New("dissoc called on non-hash map")
	}
	new_hm := copyHashMap(a[0].(HashMap))
	for i := 1; i < len(a); i += 1 {
		key := a[i]
		if !IsString(key) {
			return nil, errors.New("dissoc called with non-string key")
		}
		delete(new_hm.Val, key.(string))
	}
	return new_hm, nil
}

func seq(a []Top) (Top, error) {
	if a[0] == nil {
		return nil, nil
	}
	switch arg := a[0].(type) {
	case List:
		if len(arg.Val) == 0 {
			return nil, nil
		}
		return arg, nil
	case Vector:
		if len(arg.Val) == 0 {
			return nil, nil
		}
		return List{Val: arg.Val, Meta: nil}, nil
	case string:
		if len(arg) == 0 {
			return nil, nil
		}
		newSlc := []Top{}
		for _, ch := range strings.Split(arg, "") {
			newSlc = append(newSlc, ch)
		}
		return List{Val: newSlc, Meta: nil}, nil
	}
	return nil, errors.New("seq requires string or list or vector or nil")
}

// Metadata functions
func with_meta(a []Top) (Top, error) {
	if len(a) != 2 {
		return nil, errors.New("with-meta requires 2 args")
	}
	obj := a[0]
	m := a[1]
	switch tobj := obj.(type) {
	case List:
		return List{tobj.Val, m}, nil
	case Vector:
		return Vector{tobj.Val, m}, nil
	case HashMap:
		return HashMap{tobj.Val, m}, nil
	case Func:
		return Func{tobj.Fn, m}, nil
	case MalFunc:
		fn := tobj
		fn.Meta = m
		return fn, nil
	default:
		return nil, errors.New("with-meta not supported on type")
	}
}

func meta(a []Top) (Top, error) {
	obj := a[0]
	switch tobj := obj.(type) {
	case List:
		return tobj.Meta, nil
	case Vector:
		return tobj.Meta, nil
	case HashMap:
		return tobj.Meta, nil
	case Func:
		return tobj.Meta, nil
	case MalFunc:
		return tobj.Meta, nil
	default:
		return nil, errors.New("meta not supported on type")
	}
}

// Atom functions
func deref(a []Top) (Top, error) {
	if !IsAtom(a[0]) {
		return nil, errors.New("deref called with non-atom")
	}
	return a[0].(*Atom).Val, nil
}

func reset_BANG(a []Top) (Top, error) {
	if !IsAtom(a[0]) {
		return nil, errors.New("reset! called with non-atom")
	}
	a[0].(*Atom).Set(a[1])
	return a[1], nil
}

func swap_BANG(a []Top) (Top, error) {
	if !IsAtom(a[0]) {
		return nil, errors.New("swap! called with non-atom")
	}
	if len(a) < 2 {
		return nil, errors.New("swap! requires at least 2 args")
	}
	atm := a[0].(*Atom)
	args := []Top{atm.Val}
	f := a[1]
	args = append(args, a[2:]...)
	res, e := Apply(f, args)
	if e != nil {
		return nil, e
	}
	atm.Set(res)
	return res, nil
}

var GlobalFunctions = map[string]Top{
	"=": func(a []Top) (Top, error) {
		return Eq(a[0], a[1]), nil
	},
	"throw": throw,
	"nil?": func(a []Top) (Top, error) {
		return IsNil(a[0]), nil
	},
	"true?": func(a []Top) (Top, error) {
		return IsTrue(a[0]), nil
	},
	"false?": func(a []Top) (Top, error) {
		return IsFalse(a[0]), nil
	},
	"symbol": func(a []Top) (Top, error) {
		return Symbol{a[0].(string)}, nil
	},
	"symbol?": func(a []Top) (Top, error) {
		return IsSymbol(a[0]), nil
	},
	"string?": func(a []Top) (Top, error) {
		return (IsString(a[0]) && !IsKeyword(a[0])), nil
	},
	"keyword": func(a []Top) (Top, error) {
		if IsKeyword(a[0]) {
			return a[0], nil
		} else {
			return NewKeyword(a[0].(string))
		}
	},
	"keyword?": func(a []Top) (Top, error) {
		return IsKeyword(a[0]), nil
	},

	"pr-str":    func(a []Top) (Top, error) { return printStr(a) },
	"str":       func(a []Top) (Top, error) { return str(a) },
	"prn":       func(a []Top) (Top, error) { return prn(a) },
	"printLine": func(a []Top) (Top, error) { return printLine(a) },
	"read-string": func(a []Top) (Top, error) {
		return reader.Read_str(a[0].(string))
	},
	"slurp": slurp,
	"readline": func(a []Top) (Top, error) {
		return readline.Readline(a[0].(string))
	},

	"<": func(a []Top) (Top, error) {
		return a[0].(int) < a[1].(int), nil
	},
	"<=": func(a []Top) (Top, error) {
		return a[0].(int) <= a[1].(int), nil
	},
	">": func(a []Top) (Top, error) {
		return a[0].(int) > a[1].(int), nil
	},
	">=": func(a []Top) (Top, error) {
		return a[0].(int) >= a[1].(int), nil
	},
	"+": func(a []Top) (Top, error) {
		return a[0].(int) + a[1].(int), nil
	},
	"-": func(a []Top) (Top, error) {
		return a[0].(int) - a[1].(int), nil
	},
	"*": func(a []Top) (Top, error) {
		return a[0].(int) * a[1].(int), nil
	},
	"/": func(a []Top) (Top, error) {
		return a[0].(int) / a[1].(int), nil
	},
	"time-ms": time_ms,

	"list": func(a []Top) (Top, error) {
		return List{a, nil}, nil
	},
	"list?": func(a []Top) (Top, error) {
		return IsList(a[0]), nil
	},
	"vector": func(a []Top) (Top, error) {
		return Vector{a, nil}, nil
	},
	"vector?": func(a []Top) (Top, error) {
		return IsVector(a[0]), nil
	},
	"hash-map": func(a []Top) (Top, error) {
		return NewHashMap(List{a, nil})
	},
	"map?": func(a []Top) (Top, error) {
		return IsHashMap(a[0]), nil
	},
	"assoc":  assoc,
	"dissoc": dissoc,
	"get":    get,
	"contains?": func(a []Top) (Top, error) {
		return contains_Q(a[0], a[1])
	},
	"keys": keys,
	"vals": vals,

	"sequential?": func(a []Top) (Top, error) {
		return IsSeq(a[0]), nil
	},
	"cons":   cons,
	"concat": concat,
	"nth":    nth,
	"first":  first,
	"rest":   rest,
	"empty?": isEmpty,
	"count":  count,
	"apply":  apply,
	"map":    mapFunc,
	"conj":   conj,
	"seq":    seq,

	"with-meta": with_meta,
	"meta":      meta,
	"atom": func(a []Top) (Top, error) {
		return &Atom{a[0], nil}, nil
	},
	"atom?": func(a []Top) (Top, error) {
		return IsAtom(a[0]), nil
	},
	"deref":  deref,
	"reset!": reset_BANG,
	"swap!":  swap_BANG,
}
