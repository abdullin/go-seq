package seq

import (
	"strings"
)

// Immutable, append-only structure that tracks property path in object
// It is implemented as a linked list.
type Path struct {
	Value  string
	Parent *Path
}

func NewPath(elements ...string) *Path {
	var p *Path
	for _, e := range elements {
		p = p.Extend(e)
	}
	return p

}

func (p *Path) Extend(s string) *Path {
	return &Path{
		Value:  s,
		Parent: p,
	}
}
func reverse(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

func (p *Path) String() string {
	// cheap way to build path
	var list []string

	current := p
	for current != nil {
		list = append(list, current.Value)
		current = current.Parent
	}
	reverse(list)

	joined := strings.Join(list, ".")
	clean := strings.Replace(joined, ".[", "[", -1)
	return clean

}
