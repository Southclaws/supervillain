# Supervillain

Converts Go structs to Zod schemas.

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

You can pass type name mappings to custom conversion functions:

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

The function signature for custom type handlers is:

```go
func(c *supervillain.Converter, t reflect.Type, typeName, genericTypeName string, indentLevel int) string
```

You can use the Converter to process nested types. The `genericTypeName` is the name of the `T` in `Generic[T]` and the indent level is for passing to other converter APIs.

## Caveats

- Does not support self-referential types - should be a simple fix.
- Sometimes outputs in the wrong order - it really needs an intermediate DAG to solve this.
