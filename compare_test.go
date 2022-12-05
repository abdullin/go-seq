package seq

import (
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"testing"
)

type test struct {
	name     string
	e, a     proto.Message
	expected Issues
}

func TestCompare(t *testing.T) {

	empty := &Empty{}

	es := &Simple{I32: -32, I64: -64, U32: 32, U64: 64, Bool: true,
		Str: "test"}
	as := &Simple{I32: 32, I64: 64, U32: 33, U64: 65, Bool: false, Str: "tost"}

	simpleDeltas := []Issue{
		{es.I32, as.I32, NewPath("I32")},
		{es.I64, as.I64, NewPath("I64")},
		{es.U32, as.U32, NewPath("U32")},
		{es.U64, as.U64, NewPath("U64")},
		{es.Bool, as.Bool, NewPath("Bool")},
		{es.Str, as.Str, NewPath("Str")},
	}

	el := &Lists{
		Len:     []int32{1, 2, 3, 4},
		Missing: []int32{1, 2, 3, 4},
		Mistake: []*Simple{{I32: 1}},
	}
	al := &Lists{
		Len:     []int32{1, 2, 3},
		Missing: []int32{1, 2, 2, 4},
		Mistake: []*Simple{{I32: 2}},
	}
	listDeltas := []Issue{
		{4, 3, NewPath("Len", "length")},
		{int32(3), int32(2), NewPath("Missing", "[2]")},
		{int32(1), int32(2), NewPath("Mistake", "[0]", "I32")},
	}

	complexExpected := &ComplexNested{
		Locs: []*ComplexNested_Loc{
			{Uid: "uid:1", Name: "Shelf1"},
		},
	}
	complexActual := &ComplexNested{
		Locs: []*ComplexNested_Loc{
			{Uid: "00000000-0000-0000-0000-000000000001", Name: "Shelf1", Parent: "00000000-0000-0000-0000-000000000000"},
		},
	}

	expectedUids := &Uids{Uid: []string{
		"00000000-0000-0000-0000-000000000001",
		"uid:1",
		"00000000-0000-0000-0000-000000000000",
		"uid:0",
		"uid:0",
	}}

	actualUids := &Uids{
		Uid: []string{
			"00000000-0000-0000-0000-000000000001",
			"00000000-0000-0000-0000-000000000001",
			"00000000-0000-0000-0000-000000000000",
			"00000000-0000-0000-0000-000000000000",
			"",
		},
	}

	cases := []*test{
		{"similar instances", &Empty{}, &Empty{}, nil},
		{"same instance", empty, empty, nil},
		{"different instances", &Empty{}, &Simple{}, []Issue{
			{"Empty", "Simple", NewPath("type")},
		}},
		{"same simple message", es, es, nil},
		{"same lists", el, el, nil},
		{"nested fields", es, as, simpleDeltas},
		{"lists", el, al, listDeltas},
		{"uids", expectedUids, actualUids, nil},
		{"complex", complexExpected, complexActual, []Issue{
			{"", "00000000-0000-0000-0000-000000000000", NewPath("locs", "[0]", "parent")},
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := Diff(c.e, c.a, nil)
			diff := cmp.Diff(c.expected, actual)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}
