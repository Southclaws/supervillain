package decimal

import (
	"reflect"

	"github.com/m4tty-d/supervillain"
)

var (
	DecimalType = "github.com/shopspring/decimal.Decimal"
	DecimalFunc = func(c *supervillain.Converter, t reflect.Type, s, g string, i int) string {
		// Shopspring's decimal type serialises to a string.
		return "z.string()"
	}
)
