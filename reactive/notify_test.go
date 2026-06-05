package reactive_test

import (
	"testing"

	"github.com/A-Solo-Engineer/aethium/reactive"
)

func TestNotify_Selective(t *testing.T) {
	sig1 := reactive.NewSignal(10)
	sig2 := reactive.NewSignal(20)

	notified1 := 0
	notified2 := 0
	notifiedAll := 0

	reactive.SubscribeSignal(sig1.ID(), func(sid reactive.SignalID) {
		notified1++
	})

	reactive.SubscribeSignal(sig2.ID(), func(sid reactive.SignalID) {
		notified2++
	})

	reactive.SubscribeAll(func(sid reactive.SignalID) {
		notifiedAll++
	})

	// Set sig1
	sig1.Set(11)
	if notified1 != 1 {
		t.Errorf("expected sig1 listener to be notified 1 time, got %d", notified1)
	}
	if notified2 != 0 {
		t.Errorf("expected sig2 listener to be notified 0 times, got %d", notified2)
	}
	if notifiedAll != 1 {
		t.Errorf("expected SubscribeAll listener to be notified 1 time, got %d", notifiedAll)
	}

	// Set sig2
	sig2.Set(21)
	if notified1 != 1 {
		t.Errorf("expected sig1 listener to still be 1, got %d", notified1)
	}
	if notified2 != 1 {
		t.Errorf("expected sig2 listener to be notified 1 time, got %d", notified2)
	}
	if notifiedAll != 2 {
		t.Errorf("expected SubscribeAll listener to be notified 2 times, got %d", notifiedAll)
	}
}

func TestComputed_Efficiency(t *testing.T) {
	sig1 := reactive.NewSignal(1)
	sig2 := reactive.NewSignal(2)

	computeCount := 0
	computed := reactive.NewComputed(
		func() []reactive.SignalID { return []reactive.SignalID{sig1.ID()} },
		func() int {
			computeCount++
			return sig1.Get()
		},
	)

	// Initial compute
	_ = computed.Get()
	initialCount := computeCount

	// Set sig2 (which computed does NOT depend on)
	sig2.Set(3)
	
	// Check if computed was notified (it shouldn't be)
	// We can check this by seeing if Notify was called on computed.id
	// or more simply, check if computed is still clean.
	_ = computed.Get()
	if computeCount != initialCount {
		t.Errorf("expected computed NOT to recompute when unrelated signal changes, but computeCount went from %d to %d", initialCount, computeCount)
	}

	// Set sig1 (which computed DOES depend on)
	sig1.Set(2)
	_ = computed.Get()
	if computeCount <= initialCount {
		t.Error("expected computed to recompute when dependency changes")
	}
}
