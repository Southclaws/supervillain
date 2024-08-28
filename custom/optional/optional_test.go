package optional_test

import (
	"testing"

	"4d63.com/optional"
	"github.com/Southclaws/supervillain"
	customoptional "github.com/Southclaws/supervillain/custom/optional"
	"github.com/stretchr/testify/assert"
)

func TestCustom(t *testing.T) {
	c := supervillain.NewConverter(map[string]supervillain.CustomFn{
		customoptional.OptionalType: customoptional.OptionalFunc,
	})

	type Profile struct {
		Bio     string
		Twitter optional.Optional[string]
	}

	type User struct {
		MaybeName    optional.Optional[string]
		MaybeAge     optional.Optional[int]
		MaybeHeight  optional.Optional[float64]
		MaybeProfile optional.Optional[Profile]
	}
	assert.Equal(t,
		`export const ProfileSchema = z.object({
  Bio: z.string(),
  Twitter: z.string().optional(),
})
export type Profile = z.infer<typeof ProfileSchema>

export const UserSchema = z.object({
  MaybeName: z.string().optional(),
  MaybeAge: z.number().optional(),
  MaybeHeight: z.number().optional(),
  MaybeProfile: ProfileSchema.optional(),
})
export type User = z.infer<typeof UserSchema>

`,
		c.Convert(User{}))
}
