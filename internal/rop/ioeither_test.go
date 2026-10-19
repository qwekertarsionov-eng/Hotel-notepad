package rop

import (
	"context"
	"testing"
	"time"
)

func TestTST_ROP_004(t *testing.T) {
	calls := 0
	eff := FromIO[probeErr, int](func() (int, error) {
		calls++
		return 5, nil
	})
	if calls != 0 {
		t.Fatalf("effect must not run before Run, got %d calls", calls)
	}
	got := eff.Run(context.Background())
	if calls != 1 {
		t.Fatalf("effect must run exactly once on Run, got %d calls", calls)
	}
	gotVal, ok := foldRightVal(got)
	if !ok || gotVal != 5 {
		t.Fatal("Run must return Right with the effect result")
	}
}

func TestTST_ROP_005(t *testing.T) {
	cause := probeErr("io failure")
	eff := FromIO[probeErr, int](func() (int, error) { return 0, cause })
	got := eff.Run(context.Background())
	gotErr, ok := foldLeftVal(got)
	if !ok || gotErr != cause {
		t.Fatal("FromIO must return Left with the same cause without losing it")
	}
}

func TestTST_ROP_006(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	eff := FromIO[probeErr, int](func() (int, error) {
		calls++
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(time.Hour):
			return 0, nil
		}
	})
	got := eff.Run(ctx)
	if calls != 1 {
		t.Fatalf("effect must run exactly once with cancelled ctx, got %d calls", calls)
	}
	_, ok := foldLeftVal(got)
	if !ok {
		t.Fatal("Run with cancelled ctx must return Left")
	}
}

func TestTST_ROP_007(t *testing.T) {
	calls := 0
	eff := FromIO[probeErr, int](func() (int, error) {
		calls++
		return calls, nil
	})
	first := eff.Run(context.Background())
	second := eff.Run(context.Background())
	if calls != 2 {
		t.Fatalf("two Runs must execute the effect twice, got %d calls", calls)
	}
	_, ok1 := foldRightVal(first)
	_, ok2 := foldRightVal(second)
	if !ok1 || !ok2 {
		t.Fatal("both Runs must return Right")
	}
}

func TestTST_ROP_012(t *testing.T) {
	calls := 0
	base := FromIO[probeErr, int](func() (int, error) {
		calls++
		return 1, nil
	})
	step := base.MapIO(func(i int) int { return i + 1 }).FlatMapIO(func(i int) IOEither[probeErr, int] {
		return FromIO[probeErr, int](func() (int, error) {
			calls++
			return i, nil
		})
	})
	if calls != 0 {
		t.Fatalf("MapIO/FlatMapIO must be lazy, got %d calls before Run", calls)
	}
	got := step.Run(context.Background())
	if calls != 2 {
		t.Fatalf("Run must execute both effects, got %d calls", calls)
	}
	gotVal, ok := foldRightVal(got)
	if !ok || gotVal != 2 {
		t.Fatal("Run must return Right with the composed result")
	}
}

func TestTST_ROP_013(t *testing.T) {
	got := FromIO[probeErr, int](func() (int, error) { return 42, nil }).Run(context.Background())
	gotVal, ok := foldRightVal(got)
	if !ok || gotVal != 42 {
		t.Fatal("FromIO of a successful value must return Right")
	}
}

func TestTST_ROP_014(t *testing.T) {
	cause := probeErr("panics")
	got := FromIO[probeErr, int](func() (int, error) { panic(cause) }).Run(context.Background())
	gotErr, ok := foldLeftVal(got)
	if !ok || gotErr != cause {
		t.Fatal("FromIO panic must become Left with the cause without loss")
	}
	got = FromIO[probeErr, int](func() (int, error) { return 0, cause }).Run(context.Background())
	gotErr, ok = foldLeftVal(got)
	if !ok || gotErr != cause {
		t.Fatal("FromIO error must become Left with the same cause")
	}
}
