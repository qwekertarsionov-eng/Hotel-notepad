package rop

import "reflect"

// Either[E, T] represents a value that is either a Left (error) or a Right (success).
type Either[E, T any] struct {
	kind uint8 // 0 = left, 1 = right
	err  E
	val  T
}

// Left constructs a Left Either carrying an error.
func Left[E, T any](e E) Either[E, T] {
	return Either[E, T]{kind: 0, err: e}
}

// Right constructs a Right Either carrying a value.
func Right[E, T any](v T) Either[E, T] {
	return Either[E, T]{kind: 1, val: v}
}

// IsLeft reports whether e is a Left.
func IsLeft[E, T any](e Either[E, T]) bool {
	return e.kind == 0
}

// IsRight reports whether e is a Right.
func IsRight[E, T any](e Either[E, T]) bool {
	return e.kind == 1
}

// Fold applies onLeft to the error or onRight to the value and returns the result.
// Both callbacks may return a value or nothing; the result is returned as any.
func Fold[E, T any](e Either[E, T], onLeft, onRight interface{}) any {
	var fn reflect.Value
	var arg reflect.Value
	if e.kind == 0 {
		fn = reflect.ValueOf(onLeft)
		arg = reflect.ValueOf(e.err)
	} else {
		fn = reflect.ValueOf(onRight)
		arg = reflect.ValueOf(e.val)
	}
	outs := fn.Call([]reflect.Value{arg})
	if len(outs) == 0 {
		return nil
	}
	return outs[0].Interface()
}

// Map transforms the value of a Right, leaving a Left untouched.
func Map[E, T1, T2 any](e Either[E, T1], f func(T1) T2) Either[E, T2] {
	if e.kind == 0 {
		return Left[E, T2](e.err)
	}
	return Right[E, T2](f(e.val))
}

// FlatMap chains a function returning an Either over a Right, leaving a Left untouched.
func FlatMap[E, T1, T2 any](e Either[E, T1], f func(T1) Either[E, T2]) Either[E, T2] {
	if e.kind == 0 {
		return Left[E, T2](e.err)
	}
	return f(e.val)
}

// MapLeft transforms the error of a Left, leaving a Right untouched.
func MapLeft[E1, E2, T any](e Either[E1, T], f func(E1) E2) Either[E2, T] {
	if e.kind == 0 {
		return Left[E2, T](f(e.err))
	}
	return Right[E2, T](e.val)
}

// EitherOf returns a Right carrying v.
func EitherOf[E, T any](v T) Either[E, T] {
	return Right[E, T](v)
}
