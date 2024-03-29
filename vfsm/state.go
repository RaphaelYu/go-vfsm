package vfsm

import (
	"errors"
	"io"

	"github.com/Knetic/govaluate"
)

type (
	Transition[T StatefulObjectIf] struct {
		next string
		cond func(T, map[string]any) (bool, error)
	}

	ExceptionTransition[T StatefulObjectIf] struct {
		next string
		cond func(T, map[string]any, error, string) (bool, error)
	}

	State[T StatefulObjectIf] struct {
		enter           func(T) error
		id              string
		onGood          []*Transition[T]
		exceptionHandle *ExceptionTransition[T]
		exit            func(T) error
	}
)

func (tr ExceptionTransition[T]) Transit(self T, params map[string]any, e error, st string) (string, error) {
	hit, err := tr.cond(self, params, e, st)
	if err != nil {
		return "", err
	}
	if hit {
		return tr.next, err
	}
	return "", err
}

func (tr Transition[T]) Transit(self T, params map[string]any) (string, error) {
	hit, err := tr.cond(self, params)
	if err != nil {
		return "", err
	}
	if hit {
		return tr.next, err
	}
	return "", err
}

func BindHitThen[T StatefulObjectIf](cond func(T, map[string]any) (bool, error), next string) *Transition[T] {
	return &Transition[T]{
		next, cond,
	}
}

func (st *State[T]) Resolves(t ...*Transition[T]) {
	st.onGood = t

}

func (st *State[T]) Next(obj T, params map[string]any) (next string, err error) {
	if len(st.onGood) == 0 {
		return "", io.EOF
	}
	for _, f := range st.onGood {
		if next, err = f.Transit(obj, params); err != nil {
			next, err = st.exceptionHandle.Transit(obj, params, err, st.id)
			return

		}
		if len(next) > 0 {
			return
		}
	}
	return
}

func (st *State[T]) Exit(obj T) error {
	return st.exit(obj)
}

type ExpressionCondtion[T StatefulObjectIf] struct {
	expr      *govaluate.EvaluableExpression
	converter func(any) (bool, error)
}

func (expr *ExpressionCondtion[T]) Parse(v string) error {
	var err error
	expr.expr, err = govaluate.NewEvaluableExpression(v)
	return err
}

func (expr *ExpressionCondtion[T]) SetAdvance(f func(any) (bool, error)) {
	expr.converter = f
}

const statusField = "object:state"

var errWrongConditionExpression = errors.New("wrong condition expression")

func (expr *ExpressionCondtion[T]) Eval(obj T, params map[string]any) (bool, error) {
	params[statusField] = obj.Current()
	ret, err := expr.expr.Evaluate(params)
	if err != nil {
		return false, err
	}
	if expr.converter != nil {
		return expr.converter(ret)
	}
	retv, matched := ret.(bool)
	if !matched {
		return false, errWrongConditionExpression
	}
	return retv, nil
}
