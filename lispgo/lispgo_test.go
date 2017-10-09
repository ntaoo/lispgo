package main

import (
	"testing"
)

type TestCode struct {
	title    string
	code     string
	expected string
}

// TODO: Implement cons, remainder, modulo, sqrt, sin, cos, tan, asin, acos, atan, car, cdr, etc...
func newSuccessCodeArray() []TestCode {
	t := make([]TestCode, 0, 0)
	t = append(t, TestCode{title: "()", code: "()", expected: "()"})

	// Calc
	t = append(t, TestCode{title: "+", code: "(+ 1 2)", expected: "3"})
	t = append(t, TestCode{title: "- 1", code: "(- 10 3)", expected: "7"})
	//t = append(t, TestCode{title: "- 2", code: "(- 10 3 5)", expected: "2"}) // FIXME
	t = append(t, TestCode{title: "* 1", code: "(* 2 3)", expected: "6"})
	//t = append(t, TestCode{title: "* 2", code: "(* 2 3 4)", expected: "24"}) // FIXME
	t = append(t, TestCode{title: "/", code: "(/ 9 6)", expected: "1"})
	t = append(t, TestCode{title: "nested calc", code: "(* (+ 2 3) (- 5 3))", expected: "10"})
	t = append(t, TestCode{title: "nested calc2", code: "(/ (+ 9 1) (+ 2 3))", expected: "2"})

	// list
	t = append(t, TestCode{title: "list 1", code: "(list)", expected: "()"})
	t = append(t, TestCode{title: "list 2", code: "(list 1)", expected: "(1)"})
	t = append(t, TestCode{title: "list 3", code: "(list '(1 2) '(3 4))", expected: "((1 2) (3 4))"})

	t = append(t, TestCode{title: "Quote", code: "(quote (testing 1 (2.0) -3.14e159))", expected: "(testing 1 (2.0) -3.14e159)"})
	t = append(t, TestCode{title: "If", code: "(if 1 2)", expected: "2"})
	t = append(t, TestCode{title: "If2", code: "(if (= 3 4) 2)", expected: "nil"})
	t = append(t, TestCode{title: "Else", code: "(if (= 3 4) 2 5)", expected: "5"})
	return t
}

func newErrorCodeArray() []TestCode {
	t := make([]TestCode, 0, 0)
	t = append(t, TestCode{title: "Syntax Error", code: "(1)"})
	return t
}

func TestSuccessCases(t *testing.T) {
	boot()
	for _, element := range newSuccessCodeArray() {
		actual, err := rep(element.code)
		if err != nil {
			t.Errorf("eval: %v, title: %v, to expect %v, but outputs the error unexpectedly: %v",
				element.code, element.title, element.expected, err)
		} else if actual != element.expected {
			t.Errorf("eval %v, title: %v, to expect %v, but actual: %v", element.code, element.title, element.expected, actual)
		}
	}
}

func TestErrorCases(t *testing.T) {
	boot()
	for _, element := range newErrorCodeArray() {
		actual, err := rep(element.code)
		if err == nil {
			t.Errorf("eval: %v, title: %v, but actual: %v", element.code, element.title, actual)
		}
	}
}

func TestDefineFunc(t *testing.T) {
	boot()
	var err error
	_, err = rep("(define inc (lambda (a) (+ a 1)))")
	actual, err := rep("(inc 1)")
	expected := "2"
	if err != nil {
		t.Errorf("define func has an error, %v", err)
	} else if actual != expected {
		t.Errorf("define func has an error. expected: %v, actual: %v", expected, actual)
	}
}

func TestLambda(t *testing.T) {
	boot()
	var err error
	_, err = rep("(define sum2 (lambda (a b) (+ a b)))")
	actual, err := rep("(sum2 1 2)")
	expected := "3"
	if err != nil {
		t.Errorf("define func has an error, %v", err)
	} else if actual != expected {
		t.Errorf("define func has an error. expected: %v, actual: %v", expected, actual)
	}
}
