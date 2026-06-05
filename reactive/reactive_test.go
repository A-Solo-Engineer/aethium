package reactive_test

import (
	"testing"

	"github.com/A-Solo-Engineer/aethium/reactive"
)

func BenchmarkSignalNotify(b *testing.B) {
	b.ReportAllocs()
	sig := reactive.NewSignal(0)
	reactive.Subscribe(func() {})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sig.Set(i)
	}
}

func BenchmarkSignalGet(b *testing.B) {
	b.ReportAllocs()
	sig := reactive.NewSignal(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sig.Get()
	}
}

func BenchmarkSignalGetTracked(b *testing.B) {
	b.ReportAllocs()
	sig := reactive.NewSignal(0)
	tracker := &mockTracker{}
	reactive.PushDependencyTracker(tracker)
	defer reactive.PopDependencyTracker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sig.Get()
	}
}

func BenchmarkComputedGet(b *testing.B) {
	b.ReportAllocs()
	computed := reactive.NewComputed(
		func() int { return 0 },
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = computed.Get()
	}
}

func BenchmarkComputedGetStale(b *testing.B) {
	b.ReportAllocs()
	computed := reactive.NewComputed(
		func() int { return 0 },
	)

	// Mark stale once before benchmark
	computed.Invalidate()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = computed.Get()
	}
}

type mockTracker struct{}

func (m *mockTracker) Track(id reactive.SignalID) {}
