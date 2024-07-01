package optional

import (
	"fmt"
	"reflect"

	"github.com/m4tty-d/supervillain"
)

var (
	OptionalType = "4d63.com/optional.Optional"
	OptionalFunc = func(c *supervillain.Converter, t reflect.Type, s string, g string, i int) string {
		return fmt.Sprintf("%s.optional().nullish()", c.ConvertType(t.Elem(), s, i))
	}
)
