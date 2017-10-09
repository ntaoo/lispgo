package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

import (
	"github.com/ntaoo/lispgo/core"
	. "github.com/ntaoo/lispgo/env"
	"github.com/ntaoo/lispgo/printer"
	"github.com/ntaoo/lispgo/reader"
	"github.com/ntaoo/lispgo/readline"
	. "github.com/ntaoo/lispgo/types"
)

// read
func Read(str string) (Top, error) {
	return reader.Read_str(str)
}

// eval
func isPair(x Top) bool {
	slc, e := GetSlice(x)
	if e != nil {
		return false
	}
	return len(slc) > 0
}

func quasiquote(ast Top) Top {
	if !isPair(ast) {
		return List{Val: []Top{Symbol{Val: "quote"}, ast}, Meta: nil}
	} else {
		slc, _ := GetSlice(ast)
		a0 := slc[0]
		if IsSymbol(a0) && (a0.(Symbol).Val == "unquote") {
			return slc[1]
		} else if isPair(a0) {
			slc0, _ := GetSlice(a0)
			a00 := slc0[0]
			if IsSymbol(a00) && (a00.(Symbol).Val == "splice-unquote") {
				return List{Val: []Top{Symbol{Val: "concat"},
					slc0[1],
					quasiquote(List{Val: slc[1:], Meta: nil})}, Meta: nil}
			}
		}
		return List{Val: []Top{Symbol{Val: "cons"}, quasiquote(a0), quasiquote(List{Val: slc[1:], Meta: nil})}, Meta: nil}
	}
}

func isMacroCall(ast Top, env EnvType) bool {
	if IsList(ast) {
		slc, _ := GetSlice(ast)
		if len(slc) == 0 {
			return false
		}
		a0 := slc[0]
		if IsSymbol(a0) && env.Find(a0.(Symbol)) != nil {
			mac, e := env.Get(a0.(Symbol))
			if e != nil {
				return false
			}
			if MalFunc_Q(mac) {
				return mac.(MalFunc).GetMacro()
			}
		}
	}
	return false
}

func macroExpand(ast Top, env EnvType) (Top, error) {
	var mac Top
	var e error
	for isMacroCall(ast, env) {
		slc, _ := GetSlice(ast)
		a0 := slc[0]
		mac, e = env.Get(a0.(Symbol))
		if e != nil {
			return nil, e
		}
		fn := mac.(MalFunc)
		ast, e = Apply(fn, slc[1:])
		if e != nil {
			return nil, e
		}
	}
	return ast, nil
}

func evalAST(ast Top, env EnvType) (Top, error) {
	//fmt.Printf("evalAST: %#v\n", ast)
	if IsSymbol(ast) {
		return env.Get(ast.(Symbol))
	} else if IsList(ast) {
		lst := []Top{}
		for _, a := range ast.(List).Val {
			exp, e := Eval(a, env)
			if e != nil {
				return nil, e
			}
			lst = append(lst, exp)
		}
		return List{lst, nil}, nil
	} else if IsVector(ast) {
		lst := []Top{}
		for _, a := range ast.(Vector).Val {
			exp, e := Eval(a, env)
			if e != nil {
				return nil, e
			}
			lst = append(lst, exp)
		}
		return Vector{lst, nil}, nil
	} else if IsHashMap(ast) {
		m := ast.(HashMap)
		new_hm := HashMap{map[string]Top{}, nil}
		for k, v := range m.Val {
			ke, e1 := Eval(k, env)
			if e1 != nil {
				return nil, e1
			}
			if _, ok := ke.(string); !ok {
				return nil, errors.New("non string hash-map key")
			}
			kv, e2 := Eval(v, env)
			if e2 != nil {
				return nil, e2
			}
			new_hm.Val[ke.(string)] = kv
		}
		return new_hm, nil
	} else {
		return ast, nil
	}
}

func Eval(ast Top, env EnvType) (Top, error) {
	var e error
	for {

		//fmt.Printf("Eval: %v\n", printer.PrintString(ast, true))
		switch ast.(type) {
		case List: // continue
		default:
			return evalAST(ast, env)
		}

		// apply list
		ast, e = macroExpand(ast, env)
		if e != nil {
			return nil, e
		}
		if !IsList(ast) {
			return evalAST(ast, env)
		}
		if len(ast.(List).Val) == 0 {
			return ast, nil
		}

		a0 := ast.(List).Val[0]
		var a1 Top = nil
		var a2 Top = nil
		switch len(ast.(List).Val) {
		case 1:
			a1 = nil
			a2 = nil
		case 2:
			a1 = ast.(List).Val[1]
			a2 = nil
		default:
			a1 = ast.(List).Val[1]
			a2 = ast.(List).Val[2]
		}
		a0sym := "__<*lambda>__"
		if IsSymbol(a0) {
			a0sym = a0.(Symbol).Val
		}
		switch a0sym {
		case "define":
			res, e := Eval(a2, env)
			if e != nil {
				return nil, e
			}
			return env.Set(a1.(Symbol), res), nil
		case "let*":
			let_env, e := NewEnv(env, nil, nil)
			if e != nil {
				return nil, e
			}
			arr1, e := GetSlice(a1)
			if e != nil {
				return nil, e
			}
			for i := 0; i < len(arr1); i += 2 {
				if !IsSymbol(arr1[i]) {
					return nil, errors.New("non-symbol bind value")
				}
				exp, e := Eval(arr1[i+1], let_env)
				if e != nil {
					return nil, e
				}
				let_env.Set(arr1[i].(Symbol), exp)
			}
			ast = a2
			env = let_env
		case "quote":
			return a1, nil
		case "quasiquote":
			ast = quasiquote(a1)
		case "defmacro!":
			fn, e := Eval(a2, env)
			fn = fn.(MalFunc).SetMacro()
			if e != nil {
				return nil, e
			}
			return env.Set(a1.(Symbol), fn), nil
		case "macroExpand":
			return macroExpand(a1, env)
		case "try*":
			var exc Top
			exp, e := Eval(a1, env)
			if e == nil {
				return exp, nil
			} else {
				if a2 != nil && IsList(a2) {
					a2s, _ := GetSlice(a2)
					if IsSymbol(a2s[0]) && (a2s[0].(Symbol).Val == "catch*") {
						switch e.(type) {
						case LGError:
							exc = e.(LGError).Obj
						default:
							exc = e.Error()
						}
						binds := NewList(a2s[1])
						new_env, e := NewEnv(env, binds, NewList(exc))
						if e != nil {
							return nil, e
						}
						exp, e = Eval(a2s[2], new_env)
						if e == nil {
							return exp, nil
						}
					}
				}
				return nil, e
			}
		case "do":
			lst := ast.(List).Val
			_, e := evalAST(List{lst[1 : len(lst)-1], nil}, env)
			if e != nil {
				return nil, e
			}
			if len(lst) == 1 {
				return nil, nil
			}
			ast = lst[len(lst)-1]
		case "if":
			cond, e := Eval(a1, env)
			if e != nil {
				return nil, e
			}
			if cond == nil || cond == false {
				if len(ast.(List).Val) >= 4 {
					ast = ast.(List).Val[3]
				} else {
					return nil, nil
				}
			} else {
				ast = a2
			}
		case "lambda":
			fn := MalFunc{Eval, a2, env, a1, false, NewEnv, nil}
			return fn, nil
		default:
			el, e := evalAST(ast, env)
			if e != nil {
				return nil, e
			}
			f := el.(List).Val[0]
			if MalFunc_Q(f) {
				fn := f.(MalFunc)
				ast = fn.Exp
				env, e = NewEnv(fn.Env, fn.Params, List{el.(List).Val[1:], nil})
				if e != nil {
					return nil, e
				}
			} else {
				fn, ok := f.(Func)
				if !ok {
					return nil, errors.New("attempt to call non-function")
				}
				return fn.Fn(el.(List).Val[1:])
			}
		}

	} // TCO loop
}

// print
func Print(exp Top) (string, error) {
	return printer.PrintString(exp, true), nil
}

var replEnv, _ = NewEnv(nil, nil, nil)

// repl
func rep(str string) (Top, error) {
	var exp Top
	var res string
	var e error
	if exp, e = Read(str); e != nil {
		return nil, e
	}
	if exp, e = Eval(exp, replEnv); e != nil {
		return nil, e
	}
	if res, e = Print(exp); e != nil {
		return nil, e
	}
	return res, nil
}

func boot() {
	// core.go: defined using go
	for k, v := range core.GlobalFunctions {
		replEnv.Set(Symbol{k}, Func{v.(func([]Top) (Top, error)), nil})
	}
	replEnv.Set(Symbol{"eval"}, Func{func(a []Top) (Top, error) {
		return Eval(a[0], replEnv)
	}, nil})
	replEnv.Set(Symbol{"*ARGV*"}, List{})

	// core.mal: defined using the language itself
	rep("(define *host-language* \"go\")")
	rep("(define not (lambda (a) (if a false true)))")
	rep("(define load-file (lambda (f) (eval (read-string (str \"(do \" (slurp f) \")\")))))")
	rep("(defmacro! cond (lambda (& xs) (if (> (count xs) 0) (list 'if (first xs) (if (> (count xs) 1) (nth xs 1) (throw \"odd number of forms to cond\")) (cons 'cond (rest (rest xs)))))))")
	rep("(define *gensym-counter* (atom 0))")
	rep("(define gensym (lambda [] (symbol (str \"G__\" (swap! *gensym-counter* (lambda [x] (+ 1 x)))))))")
	rep("(defmacro! or (lambda (& xs) (if (empty? xs) nil (if (= 1 (count xs)) (first xs) (let* (condvar (gensym)) `(let* (~condvar ~(first xs)) (if ~condvar ~condvar (or ~@(rest xs)))))))))")

	// called with mal script to load and eval
	if len(os.Args) > 1 {
		args := make([]Top, 0, len(os.Args)-2)
		for _, a := range os.Args[2:] {
			args = append(args, a)
		}
		replEnv.Set(Symbol{"*ARGV*"}, List{args, nil})
		if _, e := rep("(load-file \"" + os.Args[1] + "\")"); e != nil {
			fmt.Printf("Error: %v\n", e)
			os.Exit(1)
		}
		os.Exit(0)
	}
}

func main() {
	boot()

	// repl loop
	rep("(println (str \"Mal [\" *host-language* \"]\"))")
	for {
		text, err := readline.Readline("lisp> ")
		text = strings.TrimRight(text, "\n")
		if err != nil {
			return
		}

		var out Top
		var e error
		if out, e = rep(text); e != nil {
			if e.Error() == "<empty line>" {
				continue
			}
			fmt.Printf("Error: %v\n", e)
			continue
		}
		fmt.Printf("%v\n", out)
	}
}
