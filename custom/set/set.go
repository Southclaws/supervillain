package set

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/m4tty-d/supervillain"
)

type Set[T comparable] map[T]struct{}

func (s Set[T]) MarshalJSON() ([]byte, error) {
	keys := make([]T, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return json.Marshal(keys)
}

func (s Set[T]) ZodSchema(c *supervillain.Converter, t reflect.Type, name, generic string, indent int) string {
	return fmt.Sprintf("%s.array()", c.ConvertType(t.Key(), name, indent))
}
