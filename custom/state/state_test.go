package state_test

import (
	"testing"

	"github.com/m4tty-d/supervillain"
	"github.com/m4tty-d/supervillain/custom/state"
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
