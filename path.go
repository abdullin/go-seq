package seq

import "strings"

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

func (p *Path) String() string {
	// cheap way to build path
	var list []string

	current := p
	for {
		list = append(list, current.Value)
		if current.Parent == nil {
			break
		}
		current = p.Parent
	}

	joined := strings.Join(list, ".")
	clean := strings.Replace(joined, ".[", "[", -1)
	return clean

}
