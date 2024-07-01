package set_test

import (
	"testing"

	"github.com/m4tty-d/supervillain"
	"github.com/m4tty-d/supervillain/custom/set"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	type User struct {
		Nicknames       set.Set[string]
		FavoriteNumbers set.Set[int]
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  Nicknames: z.string().array(),
  FavoriteNumbers: z.number().array(),
})
export type User = z.infer<typeof UserSchema>

`,
		supervillain.StructToZodSchema(User{}))
}
