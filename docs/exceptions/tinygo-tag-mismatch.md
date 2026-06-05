# Exception: TinyGo -tags flag mismatch

**File:** `internal/build/build.go`, `buildWasm` function
**Issue:** Passed `-tags=tinygo,js` but `js_stub.go` uses `//go:build tinygo`
**Resolution:** Fixed — changed to `-tags=tinygo`
**Status:** RESOLVED
**Resolved in:** v0.0.1 (pre-tag cleanup)
