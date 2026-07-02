package testutil

import (
	"testing"
	"time"
)

func TestAssertNoGoroutineLeakAllowsStoppedWorker(t *testing.T) {
	AssertNoGoroutineLeak(t, 0, func() {
		done := make(chan struct{})
		go func() {
			close(done)
		}()
		<-done
	})
}

func TestAssertNoGoroutineLeakReportsLeakedWorker(t *testing.T) {
	if testing.Short() {
		t.Skip("uses real timeout")
	}
	leak := make(chan struct{})
	defer close(leak)

	fake := &fakeTB{}
	AssertNoGoroutineLeak(fake, 0, func() {
		go func() { <-leak }()
	})
	if !fake.failed {
		t.Fatal("expected leaked worker to fail the fake test")
	}
}

type fakeTB struct {
	failed bool
}

func (f *fakeTB) Helper() {}
func (f *fakeTB) Fatalf(string, ...any) {
	f.failed = true
}
func (f *fakeTB) Cleanup(func())                          {}
func (f *fakeTB) Deadline() (deadline time.Time, ok bool) { return time.Time{}, false }
func (f *fakeTB) Error(args ...any)                       { f.failed = true }
func (f *fakeTB) Errorf(string, ...any)                   { f.failed = true }
func (f *fakeTB) Fail()                                   { f.failed = true }
func (f *fakeTB) FailNow()                                { f.failed = true }
func (f *fakeTB) Failed() bool                            { return f.failed }
func (f *fakeTB) Log(args ...any)                         {}
func (f *fakeTB) Logf(string, ...any)                     {}
func (f *fakeTB) Name() string                            { return "fake" }
func (f *fakeTB) Skip(args ...any)                        {}
func (f *fakeTB) SkipNow()                                {}
func (f *fakeTB) Skipf(string, ...any)                    {}
func (f *fakeTB) Skipped() bool                           { return false }
func (f *fakeTB) TempDir() string                         { return "" }
