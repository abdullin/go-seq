package seq

import (
	"fmt"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"reflect"
	"strconv"
	"strings"
)

// Issue reports a single difference between object trees
type Issue struct {
	Expected, Actual interface{}
	Path             *Path
}

type Issues []Issue

func Format(val interface{}) string {
	if val == nil {
		return "<nil>"
	}
	switch v := val.(type) {
	case proto.Message:

		repr := prototext.MarshalOptions{Multiline: false}.Format(v)
		return string(v.ProtoReflect().Descriptor().Name()) + ":" + repr
	case []proto.Message:
		names := []string{}
		for _, m := range v {
			names = append(names, string(m.ProtoReflect().Descriptor().Name()))
		}
		return fmt.Sprintf("[%s]", strings.Join(names, ", "))

	case error:
		return fmt.Sprintf("Error '%v'", v.Error())
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

func (d Issue) String() string {
	return fmt.Sprintf("Expected %v to be %v but got %v",
		d.Path.String(),
		Format(d.Expected),
		Format(d.Actual))
}

func Diff(expected, actual proto.Message, path *Path) Issues {

	enil, anil := expected == nil, actual == nil
	if enil && anil {
		// both are nil. Good
		return nil
	}

	if enil != anil {
		// one of them is nil. Quit now, too
		return []Issue{
			{
				Expected: expected,
				Actual:   actual,
				Path:     path,
			},
		}
	}

	return compare(expected.ProtoReflect(), actual.ProtoReflect(), path)
}

func compare(expected, actual protoreflect.Message, path *Path) (r Issues) {
	e, a := expected, actual
	ed, ad := e.Descriptor(), a.Descriptor()
	if ed != ad {

		r = append(r, Issue{
			Expected: string(e.Descriptor().Name()),
			Actual:   string(a.Descriptor().Name()),
			Path:     path.Extend("type"),
		})
		return r
	}

	for i := 0; i < ed.Fields().Len(); i++ {
		field := ed.Fields().Get(i)

		s := field.TextName()

		pth := path.Extend(s)

		ev := e.Get(field)
		av := a.Get(field)

		switch {
		case field.IsList():
			r = append(r, handleList(field, ev, av, pth)...)
		case field.IsMap():
			panic("maps not handled")
		default:
			deltas := handleSingular(field, ev, av, pth)
			r = append(r, deltas...)
		}
	}

	return r
}

func handleList(field protoreflect.FieldDescriptor, ev protoreflect.Value, av protoreflect.Value, pth *Path) (r Issues) {
	el := ev.List()
	al := av.List()

	if el.Len() != al.Len() {
		return []Issue{{Expected: el.Len(), Actual: al.Len(), Path: pth.Extend("length")}}
	} else {
		for i := 0; i < el.Len(); i++ {
			ev, av := el.Get(i), al.Get(i)
			deltas := handleSingular(field, ev, av, pth.Extend(fmt.Sprintf("[%d]", i)))
			r = append(r, deltas...)
		}
	}
	return r
}

func ParseExpectedUuid(s string) (i int64, ok bool) {

	if !strings.HasPrefix(s, "uid:") {
		return 0, false
	}

	i, err := strconv.ParseInt(s[4:], 10, 64)
	if err != nil {
		return 0, false
	}
	return i, true
}

func ParseExpectedUid(s string) (int64, bool) {
	if !strings.HasPrefix(s, "uid:") {
		return 0, false
	}
	i, err := strconv.ParseInt(s[4:], 10, 64)
	if err != nil {
		return 0, false
	}
	return i, true
}

func ParseActualUid(s string) (int64, bool) {
	if !strings.HasPrefix(s, "00000000") {
		return 0, false
	}

	trimmed := strings.TrimLeft(s, "0-")
	if len(trimmed) == 0 {
		return 0, true
	}

	i, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, false
	}
	return i, true
}

func handleSingular(field protoreflect.FieldDescriptor, ev protoreflect.Value, av protoreflect.Value, pth *Path) Issues {
	switch field.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return compare(ev.Message(), av.Message(), pth)
	default:

		if field.Kind() == protoreflect.StringKind {
			// special case - we are hunting for guids

			if eu, ok := ParseExpectedUid(ev.String()); ok {
				if au, ok := ParseActualUid(av.String()); ok {
					if eu == au {
						return nil
					} else {
						return []Issue{{
							Expected: fmt.Sprintf("uid:%d", eu),
							Actual:   fmt.Sprintf("uid:%d", au),
							Path:     pth,
						}}
					}
				}
			}

		}

		if !reflect.DeepEqual(ev.Interface(), av.Interface()) {
			return []Issue{{
				Expected: ev.Interface(),
				Actual:   av.Interface(),
				Path:     pth,
			}}
		}

	}
	return nil
}
