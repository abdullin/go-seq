package seq

import "testing"

func TestPathString(t *testing.T) {

	cases := [][]string{
		{"one", "one"},
		{"first", "last", "first.last"},
		{"list", "[1]", "property", "list[1].property"},
	}

	for _, c := range cases {
		t.Run(c[len(c)-1], func(t *testing.T) {
			p := NewPath(c[:len(c)-1]...)

			actual := p.String()
			expected := c[len(c)-1]
			if actual != expected {
				t.Fatalf("Expected %q, got %q", expected, actual)
			}
		})
	}

}
