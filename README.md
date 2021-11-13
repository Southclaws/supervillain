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

## Caveats

- Does not support self-referential types - should be a simple fix.
- Sometimes outputs in the wrong order - it really needs an intermediate DAG to solve this.
