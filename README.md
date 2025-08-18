# Supervillain

Converts Go structs to Zod schemas.

**✅ Compatible with Zod 3.x and Zod 4.x**

Usage:

```go
type Post struct {
    Title string
}
type User struct {
    Name   string
    Nickname *string // pointers become optional
    Age    int
    Height float64
    Tags []string
    Favourites []struct { // nested structs are kept inline
        Name string
    }
    Posts []Post // external structs are emitted as separate exports
}

StructToZodSchema(User{})
```

Outputs:

```typescript
export const PostSchema = z.object({
  title: z.string(),
});
export type Post = z.infer<typeof PostSchema>;

export const UserSchema = z.object({
  name: z.string(),
  nickname: z.string().optional(),
  age: z.number(),
  height: z.number(),
  tags: z.string().array(),
  favourites: z
    .object({
      name: z.string(),
    })
    .array(),
  posts: PostSchema.array(),
});
export type User = z.infer<typeof UserSchema>;
```

## Custom Types

### Skipping fields

If a field is declared with a JSON tag prefixed with `-`, and subsequently also
present inside an `inline`d struct, the embedded field will be omitted.

```go
package main

import (
	"fmt"
)

type BaseType struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

type FinalType struct {
	Start    int `json:"-start"`
	End      int `json:"-end"`
	BaseType `json:",inline"`
}

func main() {
	fmt.Println(StructToZodSchema(FinalType{}))

	// export const FinalTypeSchema = z.object({
	//   id: z.string(),
	//   name: z.string(),
	// })
	// export type FinalType = z.infer<typeof FinalTypeSchema>
}
```

### ZodSchema() method

You can define a custom conversion using a `ZodSchema()` method. This should have one of the following types:
```go
ZodSchema() string
ZodSchema(c *supervillain.Converter, t reflect.Type, name, generic string, indent int) string
ZodSchema(convert func(t reflect.Type, name string, indent int) string, t reflect.Type, name, generic string, indent int) string
```
(The first signature is available to simplify the simple case; the last signature is available in case you do not want the package defining the type to depend on supervillain.)

Zod will obtain a schema by creating a zero value of your type and calling its ZodSchema() method.

```go
type State int

func (s State) MarshalJSON() ([]byte, error) {
  return json.Marshal(fmt.Sprint(s))
}

func (s State) ZodSchema() string {
  return "z.string()"
}

type Job struct {
  State State
}

c.Convert(Job{})
```

Outputs:

```typescript
export const JobSchema = z.object({
  State: z.string(),
})
export type Job = z.infer<typeof JobSchema>
```

### Mapping

If you don't control the type yourself, you can also pass a map of type names to custom conversion functions:

```go
c := supervillain.NewConverter(map[string]supervillain.CustomFn{
    "github.com/shopspring/decimal.Decimal": func(c *supervillain.Converter, t reflect.Type, s, g string, i int) string {
        // Shopspring's decimal type serialises to a string.
        return "z.string()"
    },
})

c.Convert(User{
    Money decimal.Decimal
})
```

Outputs:

```typescript
export const UserSchema = z.object({
  Money: z.string(),
})
export type User = z.infer<typeof UserSchema>
```

There are some custom types with tests in the "custom" directory.

## Zod Version Compatibility

This library generates TypeScript code that is compatible with both Zod 3.x and Zod 4.x. The generated schemas use standard Zod APIs that remain consistent across versions:

- `z.object({})` - Object schemas
- `z.string()`, `z.number()`, `z.boolean()` - Primitive types  
- `z.array()`, `z.record()` - Collection types
- `.optional()`, `.nullable()` - Type modifiers
- `z.infer<typeof Schema>` - Type inference

The tool focuses on generating schemas for the subset of Go types that map cleanly to TypeScript, ensuring broad compatibility with the Zod ecosystem.

The function signature for custom type handlers is:

```go
func(c *supervillain.Converter, t reflect.Type, typeName, genericTypeName string, indentLevel int) string
```

You can use the Converter to process nested types. The `genericTypeName` is the name of the `T` in `Generic[T]` and the indent level is for passing to other converter APIs.

### Custom Schema Enforcement

Types with a custom MarshalJSON() method but no custom schema are typically problematic, since the generated schema may not match the custom marshalled format. You can use the `WithStrictCustomSchemas` option to cause conversion to fail (panic) if such a type is found:

```go
c := NewConverter(map[string]CustomFn{}, WithStrictCustomSchemas(true))
// or
StructToZodSchema(User{}, WithStrictCustomSchemas(true))
```

## Caveats

- Does not support self-referential types - should be a simple fix.
- Sometimes outputs in the wrong order - it really needs an intermediate DAG to solve this.
