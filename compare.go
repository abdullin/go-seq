package seq

import (
	"fmt"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"reflect"
	"strings"
)

type Issue struct {
	Expected, Actual any
	Path             []string
}

type Issues []Issue

func Format(val any) string {
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

func (d Issue) PathStr() string {
	return JoinPath(d.Path)
}

func JoinPath(path []string) string {
	return strings.Join(path, ".")
}

func (d Issue) String() string {
	return fmt.Sprintf("Expected %v to be %v but got %v",
		d.Path,
		Format(d.Expected),
		Format(d.Actual))
}

func Diff(expected, actual proto.Message, path ...string) Issues {

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

				Actual: actual,
				Path:   path,
			},
		}
	}

	return compare(expected.ProtoReflect(), actual.ProtoReflect(), path)
}

func compare(expected, actual protoreflect.Message, path []string) (r Issues) {
	e, a := expected, actual
	ed, ad := e.Descriptor(), a.Descriptor()
	if ed != ad {

		r = append(r, Issue{
			Expected: string(e.Descriptor().Name()),
			Actual:   string(a.Descriptor().Name()),
			Path:     append(path, "type"),
		})
		return r
	}

	for i := 0; i < ed.Fields().Len(); i++ {
		field := ed.Fields().Get(i)

		s := field.TextName()

		pth := append(path, s)

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

func handleList(field protoreflect.FieldDescriptor, ev protoreflect.Value, av protoreflect.Value, pth []string) (r Issues) {
	el := ev.List()
	al := av.List()

	if el.Len() != al.Len() {
		return []Issue{{Expected: el.Len(), Actual: al.Len(), Path: append(pth, "length")}}
	} else {
		for i := 0; i < el.Len(); i++ {
			ev, av := el.Get(i), al.Get(i)
			deltas := handleSingular(field, ev, av, append(pth, fmt.Sprintf("[%d]", i)))
			r = append(r, deltas...)
		}
	}
	return r
}

func handleSingular(field protoreflect.FieldDescriptor, ev protoreflect.Value, av protoreflect.Value, pth []string) Issues {
	switch field.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return compare(ev.Message(), av.Message(), pth)
	default:
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
