package supervillain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldName(t *testing.T) {
	assert.Equal(t,
		fieldName("RCONPassword"),
		"rconPassword",
	)

	assert.Equal(t,
		fieldName("LANMode"),
		"lanMode",
	)

	assert.Equal(t,
		fieldName("ABC"),
		"abc",
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
		Nickname *string
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

func TestStringArrayOptional(t *testing.T) {
	type User struct {
		Name string
		Tags []*string
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  name: z.string(),
  tags: z.string().array().optional(),
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
  nickname: z.string().optional(),
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
