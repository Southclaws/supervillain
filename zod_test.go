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
		"RCONPassword",
	)

	assert.Equal(t,
		fieldName(reflect.StructField{Name: "LANMode"}),
		"LANMode",
	)

	assert.Equal(t,
		fieldName(reflect.StructField{Name: "ABC"}),
		"ABC",
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
  Name: z.string(),
  Age: z.number(),
  Height: z.number(),
})
export type User = z.infer<typeof UserSchema>

`,
		StructToZodSchema(User{}))
}

func TestStructSimpleWithOmittedField(t *testing.T) {
	type User struct {
		Name        string
		Age         int
		Height      float64
		NotExported string `json:"-"`
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  Name: z.string(),
  Age: z.number(),
  Height: z.number(),
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
  Name: z.string(),
  Age: z.number(),
  Height: z.number(),
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
  Tags: z.string().array().nullable(),
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
  Favourites: z.object({
    Name: z.string(),
  }).array().nullable(),
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
  Name: z.string(),
  Nickname: z.string().optional(),
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
  Name: z.string(),
  Nickname: z.string().nullable(),
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
  Name: z.string(),
  Nickname: z.string().optional().nullable(),
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
  Name: z.string(),
  Tags: z.string().array().nullable(),
})
export type User = z.infer<typeof UserSchema>

`,
		StructToZodSchema(User{}))
}

func TestInterfaceAny(t *testing.T) {
	type User struct {
		Name     string
		Metadata interface{}
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  Name: z.string(),
  Metadata: z.any(),
})
export type User = z.infer<typeof UserSchema>

`,
		StructToZodSchema(User{}))
}

func TestMapStringToString(t *testing.T) {
	type User struct {
		Name     string
		Metadata map[string]string
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  Name: z.string(),
  Metadata: z.record(z.string(), z.string()),
})
export type User = z.infer<typeof UserSchema>

`,
		StructToZodSchema(User{}))
}

func TestMapStringToInterface(t *testing.T) {
	type User struct {
		Name     string
		Metadata map[string]interface{}
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  Name: z.string(),
  Metadata: z.record(z.string(), z.any()),
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
  Title: z.string(),
})
export type Post = z.infer<typeof PostSchema>

export const UserSchema = z.object({
  Name: z.string(),
  Nickname: z.string().nullable(),
  Age: z.number(),
  Height: z.number(),
  Tags: z.string().array().nullable(),
  Favourites: z.object({
    Name: z.string(),
  }).array().nullable(),
  Posts: PostSchema.array().nullable(),
})
export type User = z.infer<typeof UserSchema>

`, StructToZodSchema(User{}))
}

func TestConvertSlice(t *testing.T) {
	type Foo struct {
		Bar string
		Baz string
		Quz string
	}

	type Zip struct {
		Zap *Foo
	}

	type Whim struct {
		Wham *Foo
	}
	c := NewConverter(map[string]CustomFn{})
	types := []interface{}{
		Zip{},
		Whim{},
	}
	assert.Equal(t,
		`export const ZipSchema = z.object({
  Zap: FooSchema.nullable(),
})
export type Zip = z.infer<typeof ZipSchema>

export const FooSchema = z.object({
  Bar: z.string(),
  Baz: z.string(),
  Quz: z.string(),
})
export type Foo = z.infer<typeof FooSchema>

export const WhimSchema = z.object({
  Wham: FooSchema.nullable(),
})
export type Whim = z.infer<typeof WhimSchema>

`, c.ConvertSlice(types))
}

func TestStructTime(t *testing.T) {
	type User struct {
		Name string
		When time.Time
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  Name: z.string(),
  When: z.string(),
})
export type User = z.infer<typeof UserSchema>

`,
		StructToZodSchema(User{}))
}

func TestCustom(t *testing.T) {
	c := NewConverter(map[string]CustomFn{
		"github.com/Southclaws/supervillain.Decimal": func(c *Converter, t reflect.Type, s, g string, i int) string {
			return "z.string()"
		},
	})

	type Decimal struct {
		value    int
		exponent int
	}

	type User struct {
		Name  string
		Money Decimal
	}
	assert.Equal(t,
		`export const UserSchema = z.object({
  Name: z.string(),
  Money: z.string(),
})
export type User = z.infer<typeof UserSchema>

`,
		c.Convert(User{}))
}
