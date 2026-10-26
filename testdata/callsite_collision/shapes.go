// Package collision reproduces the cross-file call-site collision bug: two files
// in one package with a call expression at the same (line, column), plus a method
// whose body lives in a different file than its receiver type's declaration.
package collision

// Shape is declared here; some of its methods live in other.go.
type Shape struct {
	sides int
}

func alpha() int { return 1 }

func beta() int { return 2 }

// Describe calls alpha() at line 18. The call expression starts at a column that
// is deliberately mirrored by a different call in other.go on the same line.
func (s Shape) Describe() int {
	return alpha()
}
