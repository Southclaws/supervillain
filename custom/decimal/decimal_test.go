package decimal_test

import (
	"testing"

	"github.com/m4tty-d/supervillain"
	customdecimal "github.com/m4tty-d/supervillain/custom/decimal"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCustom(t *testing.T) {
	c := supervillain.NewConverter(map[string]supervillain.CustomFn{
		customdecimal.DecimalType: customdecimal.DecimalFunc,
	})

	type User struct {
		Money decimal.Decimal
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  Money: z.string(),
})
export type User = z.infer<typeof UserSchema>

`,
		c.Convert(User{}))
}
