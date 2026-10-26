package collision

// Collision fixture (regression for the cross-file call-site bug).
//
// A GoCallsite stores only (line, column). The b.e.t.a call in Summarize below
// sits at line 18, column 9 — identical to the a.l.p.h.a call in shapes.go. A
// package-wide positional scan (the old behavior) would match shapes.go first
// and resolve Summarize -> alpha, a false edge. Scoping the scan to the
// callable's own file (other.go) resolves it correctly to beta.
//
// Second half of the trigger: Summarize is a method of Shape, whose type is
// declared in shapes.go, so the symbol table files Summarize under shapes.go,
// not other.go. The resolver must therefore key the scan on the callable's own
// Path (other.go), not its symbol-table bucket (shapes.go).
//
// Keep the beta() call on line 18, column 9.
func (s Shape) Summarize() int {
	return beta()
}
