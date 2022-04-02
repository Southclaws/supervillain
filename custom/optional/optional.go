package optional

import (
	"fmt"
	"reflect"

	"github.com/Southclaws/supervillain"
)

var (
	OptionalType = "4d63.com/optional.Optional"
	OptionalFunc = func(c *supervillain.Converter, t reflect.Type, s string, g string, i int) string {
		return fmt.Sprintf("%s.optional()", c.ConvertType(t.Elem(), s, i))
	}
)
