package supervillain

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFieldName(t *testing.T) {
	assert.Equal(t,
		fieldName(reflect.StructField{Name: "RCONPassword"}),
		"rconPassword",
	)

	assert.Equal(t,
		fieldName(reflect.StructField{Name: "LANMode"}),
		"lanMode",
	)

	assert.Equal(t,
		fieldName(reflect.StructField{Name: "ABC"}),
		"abc",
	)
}

func TestFieldNameJsonTag(t *testing.T) {
	type S struct {
		NotTheFieldName string `json:"fieldName"`
	}

	assert.Equal(t,
		fieldName(reflect.TypeOf(S{}).Field((0))),
		"fieldName",
	)
}

func TestFieldNameJsonTagOmitEmpty(t *testing.T) {
	type S struct {
		NotTheFieldName string `json:"fieldName,omitempty"`
	}

	assert.Equal(t,
		fieldName(reflect.TypeOf(S{}).Field((0))),
		"fieldName",
	)
}

func TestSchemaName(t *testing.T) {
	assert.Equal(t,
		schemaName("", "User"),
		"UserSchema",
	)
	assert.Equal(t,
		schemaName("Bot", "User"),
		"BotUserSchema",
	)
}

func TestStructSimple(t *testing.T) {
	type User struct {
		Name   string
		Age    int
		Height float64
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  name: z.string(),
  age: z.number(),
  height: z.number(),
})
export type User = z.infer<typeof UserSchema>

`,
		StructToZodSchema(User{}))
}

func TestStructSimplePrefix(t *testing.T) {
	type User struct {
		Name   string
		Age    int
		Height float64
	}
	assert.Equal(t,
		`export const BotUserSchema = z.object({
  name: z.string(),
  age: z.number(),
  height: z.number(),
})
export type BotUser = z.infer<typeof BotUserSchema>

`,
		StructToZodSchemaWithPrefix("Bot", User{}))
}

func TestStringArray(t *testing.T) {
	type User struct {
		Tags []string
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  tags: z.string().array(),
})
export type User = z.infer<typeof UserSchema>

`,
		StructToZodSchema(User{}))
}

func TestStructArray(t *testing.T) {
	type User struct {
		Favourites []struct {
			Name string
		}
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  favourites: z.object({
    name: z.string(),
  }).array(),
})
export type User = z.infer<typeof UserSchema>

`,
		StructToZodSchema(User{}))
}

func TestStringOptional(t *testing.T) {
	type User struct {
		Name     string
		Nickname string `json:",omitempty"`
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  name: z.string(),
  nickname: z.string().optional(),
})
export type User = z.infer<typeof UserSchema>

`,
		StructToZodSchema(User{}))
}

func TestStringNullable(t *testing.T) {
	type User struct {
		Name     string
		Nickname *string
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  name: z.string(),
  nickname: z.string().nullable(),
})
export type User = z.infer<typeof UserSchema>

`,
		StructToZodSchema(User{}))
}

func TestStringOptionalNullable(t *testing.T) {
	type User struct {
		Name     string
		Nickname *string `json:",omitempty"` // nil values are omitted
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  name: z.string(),
  nickname: z.string().optional().nullable(),
})
export type User = z.infer<typeof UserSchema>

`,
		StructToZodSchema(User{}))
}

func TestStringArrayNullable(t *testing.T) {
	type User struct {
		Name string
		Tags []*string
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  name: z.string(),
  tags: z.string().array().nullable(),
})
export type User = z.infer<typeof UserSchema>

`,
		StructToZodSchema(User{}))
}

func TestEverything(t *testing.T) {
	type Post struct {
		Title string
	}
	type User struct {
		Name       string
		Nickname   *string // pointers become optional
		Age        int
		Height     float64
		Tags       []string
		Favourites []struct { // nested structs are kept inline
			Name string
		}
		Posts []Post // external structs are emitted as separate exports
	}
	assert.Equal(t,
		`export const PostSchema = z.object({
  title: z.string(),
})
export type Post = z.infer<typeof PostSchema>

export const UserSchema = z.object({
  name: z.string(),
  nickname: z.string().nullable(),
  age: z.number(),
  height: z.number(),
  tags: z.string().array(),
  favourites: z.object({
    name: z.string(),
  }).array(),
  posts: PostSchema.array(),
})
export type User = z.infer<typeof UserSchema>

`, StructToZodSchema(User{}))
}

func TestStructTime(t *testing.T) {
	type User struct {
		Name string
		When time.Time
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  name: z.string(),
  when: z.string(),
})
export type User = z.infer<typeof UserSchema>

`,
		StructToZodSchema(User{}))
}
