package rop

import (
	"context"
)

// IOEither[E, T] is a lazy effect that, when run, produces an Either[E, T].
type IOEither[E, T any] struct {
	run func(context.Context) Either[E, T]
}

// FromIO lifts an effectful function into IOEither. The effect runs only on Run.
func FromIO[E, T any](f func() (T, error)) IOEither[E, T] {
	return IOEither[E, T]{
		run: func(ctx context.Context) Either[E, T] {
			var (
				v        T
				err      error
				panicked bool
			)
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
						if e, ok := r.(error); ok {
							err = e
						}
					}
				}()
				v, err = f()
			}()
			if panicked || err != nil {
				if e, ok := any(err).(E); ok {
					return Left[E, T](e)
				}
				var zero E
				return Left[E, T](zero)
			}
			return Right[E, T](v)
		},
	}
}

// MapIO transforms the success value lazily.
func (io IOEither[E, T]) MapIO[T2 any](f func(T) T2) IOEither[E, T2] {
	return IOEither[E, T2]{
		run: func(ctx context.Context) Either[E, T2] {
			e := io.run(ctx)
			if e.kind == 0 {
				return Left[E, T2](e.err)
			}
			return Right[E, T2](f(e.val))
		},
	}
}

// FlatMapIO chains an effect lazily.
func (io IOEither[E, T]) FlatMapIO[T2 any](f func(T) IOEither[E, T2]) IOEither[E, T2] {
	return IOEither[E, T2]{
		run: func(ctx context.Context) Either[E, T2] {
			e := io.run(ctx)
			if e.kind == 0 {
				return Left[E, T2](e.err)
			}
			return f(e.val).run(ctx)
		},
	}
}

// Run executes the effect once and returns the Either result.
func (io IOEither[E, T]) Run(ctx context.Context) Either[E, T] {
	return io.run(ctx)
}
