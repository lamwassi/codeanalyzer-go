package core_test

// Regression test for the cross-file call-site collision bug.
//
// A GoCallsite records only (line, column). Because Go line numbers restart per
// file, the same (line, column) occurs in multiple files of one package. The
// call-graph resolver used to scan the whole package and take the first
// positional match, so a call could be resolved against a colliding call
// expression in a sibling file — yielding a false edge (real callee, wrong
// caller). The fix scopes the scan to the callable's own source file
// (GoCallable.Path), which is authoritative even when the symbol table files a
// method under its receiver type's declaring file.
//
// Fixture testdata/callsite_collision:
//   shapes.go  line 18 col 9:  Shape.Describe   -> alpha()
//   other.go   line 18 col 9:  Shape.Summarize  -> beta()   (Shape declared in shapes.go)
// Both call sites share (18, 9). Correct edges: Describe->alpha, Summarize->beta.
// The bug produced Summarize->alpha (collision) and dropped Summarize->beta.

import "testing"

// hasCollisionEdge reports whether the shared collision app has src->tgt.
func hasCollisionEdge(src, tgt string) bool {
	for _, e := range sharedCollisionL2.CallGraph {
		if e.Source == src && e.Target == tgt {
			return true
		}
	}
	return false
}

func TestCallsiteCollision_ResolvesToOwnFile(t *testing.T) {
	const (
		describe  = "collision.Shape.Describe"
		summarize = "collision.Shape.Summarize"
		alpha     = "collision.alpha"
		beta      = "collision.beta"
	)

	// Correct edges must be present.
	if !hasCollisionEdge(describe, alpha) {
		t.Errorf("missing correct edge %s -> %s", describe, alpha)
	}
	if !hasCollisionEdge(summarize, beta) {
		t.Errorf("missing correct edge %s -> %s (the fix must resolve the call site "+
			"against the callable's own file, not a sibling file at the same line:col)", summarize, beta)
	}

	// The collision edges must NOT be present.
	if hasCollisionEdge(summarize, alpha) {
		t.Errorf("false edge present %s -> %s: call site resolved against the "+
			"wrong file at the same (line, column)", summarize, alpha)
	}
	if hasCollisionEdge(describe, beta) {
		t.Errorf("false edge present %s -> %s: call site resolved against the "+
			"wrong file at the same (line, column)", describe, beta)
	}
}
