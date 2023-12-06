package state_test

import (
	"testing"

	"github.com/Southclaws/supervillain"
	"github.com/Southclaws/supervillain/custom/state"
	"github.com/stretchr/testify/assert"
)

func TestState(t *testing.T) {
	type Job struct {
		State state.State
	}
	assert.Equal(t,
		`export const JobSchema = z.object({
  State: z.string(),
})
export type Job = z.infer<typeof JobSchema>

`,
		supervillain.StructToZodSchema(Job{}))
}
