package rop

import (
	"strconv"
	"testing"
)

type probeErr string

func (e probeErr) Error() string { return string(e) }

func foldLeftVal[E, T any](e Either[E, T]) (E, bool) {
	var err E
	var ok bool
	Fold(e, func(e E) { err = e; ok = true }, func(T) {})
	return err, ok
}

func foldRightVal[E, T any](e Either[E, T]) (T, bool) {
	var v T
	var ok bool
	Fold(e, func(E) {}, func(t T) { v = t; ok = true })
	return v, ok
}

func isLeft[E, T any](e Either[E, T]) bool {
	_, ok := foldLeftVal(e)
	return ok
}

func isRight[E, T any](e Either[E, T]) bool {
	_, ok := foldRightVal(e)
	return ok
}

func eitherEqual[E comparable, T comparable](a, b Either[E, T]) bool {
	ae, al := foldLeftVal(a)
	be, bl := foldLeftVal(b)
	if al != bl {
		return false
	}
	if al {
		return ae == be
	}
	av, _ := foldRightVal(a)
	bv, _ := foldRightVal(b)
	return av == bv
}

func TestTST_ROP_001(t *testing.T) {
	r := Right[probeErr, int](7)
	var leftHit, rightHit bool
	Fold(r, func(probeErr) { leftHit = true }, func(int) { rightHit = true })
	if leftHit || !rightHit {
		t.Fatal("Right must fold through the success branch only")
	}
	l := Left[probeErr, int](probeErr("boom"))
	leftHit, rightHit = false, false
	Fold(l, func(probeErr) { leftHit = true }, func(int) { rightHit = true })
	if !leftHit || rightHit {
		t.Fatal("Left must fold through the error branch only")
	}
	if m := Map(l, func(int) int { return 0 }); !isLeft(m) {
		t.Fatal("Map of Left must remain Left")
	}
	if m := Map(r, func(int) int { return 0 }); !isRight(m) {
		t.Fatal("Map of Right must remain Right")
	}
}

func TestTST_ROP_002(t *testing.T) {
	calls := 0
	l := Left[probeErr, int](probeErr("x"))
	out := FlatMap(l, func(int) Either[probeErr, string] {
		calls++
		return Right[probeErr, string]("ok")
	})
	if calls != 0 {
		t.Fatalf("FlatMap(Left) must not call f, got %d calls", calls)
	}
	if !isLeft(out) {
		t.Fatal("FlatMap(Left) must return Left")
	}
	gotErr, ok := foldLeftVal(out)
	if !ok || gotErr != probeErr("x") {
		t.Fatal("original Left must be returned unchanged")
	}
}

func TestTST_ROP_003(t *testing.T) {
	calls := 0
	l := Left[probeErr, int](probeErr("orig"))
	out := Map(l, func(int) string {
		calls++
		return "s"
	})
	if calls != 0 {
		t.Fatalf("Map(Left) must not call f, got %d calls", calls)
	}
	if !isLeft(out) {
		t.Fatal("Map(Left) must remain Left")
	}
	gotErr, ok := foldLeftVal(out)
	if !ok || gotErr != probeErr("orig") {
		t.Fatal("Map(Left) must preserve the original error")
	}
}

func TestTST_ROP_008(t *testing.T) {
	l := Left[probeErr, int](probeErr("e"))
	if got := Fold(l, func(e probeErr) string { return "l:" + string(e) }, func(int) string { return "r" }); got != "l:e" {
		t.Fatal("Fold must fold Left through the first function")
	}
	r := Right[probeErr, int](3)
	if got := Fold(r, func(probeErr) string { return "l" }, func(int) string { return "r" }); got != "r" {
		t.Fatal("Fold must fold Right through the second function")
	}
}

func TestTST_ROP_009(t *testing.T) {
	l := Left[probeErr, int](probeErr("e"))
	out := MapLeft(l, func(e probeErr) probeErr { return probeErr("wrapped:" + string(e)) })
	if !isLeft(out) {
		t.Fatal("MapLeft of Left must remain Left")
	}
	gotErr, ok := foldLeftVal(out)
	if !ok || gotErr != probeErr("wrapped:e") {
		t.Fatal("MapLeft must transform the error")
	}
	r := Right[probeErr, int](5)
	out = MapLeft(r, func(probeErr) probeErr { return probeErr("x") })
	if !isRight(out) {
		t.Fatal("MapLeft of Right must pass through unchanged")
	}
	gotVal, ok := foldRightVal(out)
	if !ok || gotVal != 5 {
		t.Fatal("Right value must be preserved by MapLeft")
	}
}

func TestTST_ROP_010(t *testing.T) {
	f := func(i int) Either[probeErr, int] { return Right[probeErr, int](i + 1) }
	g := func(i int) Either[probeErr, string] { return Right[probeErr, string](strconv.Itoa(i)) }
	r := Right[probeErr, int](1)
	assoc := FlatMap(FlatMap(r, f), g)
	reassoc := FlatMap(r, func(i int) Either[probeErr, string] { return FlatMap(f(i), g) })
	if !eitherEqual[probeErr, string](assoc, reassoc) {
		t.Fatal("FlatMap must be associative")
	}
	bad := Left[probeErr, int](probeErr("err"))
	if !eitherEqual[probeErr, string](FlatMap(FlatMap(bad, f), g), FlatMap(bad, func(i int) Either[probeErr, string] { return FlatMap(f(i), g) })) {
		t.Fatal("FlatMap associativity must hold on Left")
	}
}

func TestTST_ROP_011(t *testing.T) {
	f := func(i int) int { return i * 10 }
	r := Right[probeErr, int](3)
	direct := Map(r, f)
	viaFlatMap := FlatMap(r, func(i int) Either[probeErr, int] { return Right[probeErr, int](f(i)) })
	if !eitherEqual[probeErr, int](direct, viaFlatMap) {
		t.Fatal("Map through FlatMap must be consistent with direct Map")
	}
	l := Left[probeErr, int](probeErr("err"))
	if !eitherEqual[probeErr, int](Map(l, f), FlatMap(l, func(i int) Either[probeErr, int] { return Right[probeErr, int](f(i)) })) {
		t.Fatal("Map through FlatMap on Left must preserve the error")
	}
}
