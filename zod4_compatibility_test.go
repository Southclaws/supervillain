package supervillain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestZod4Compatibility ensures the generated schemas are compatible with Zod 4.
// This test validates that the basic APIs we generate are still valid in Zod 4.
func TestZod4Compatibility(t *testing.T) {
	type User struct {
		Name     string            `json:"name"`
		Age      int               `json:"age"`
		Email    *string           `json:"email,omitempty"`
		Tags     []string          `json:"tags"`
		Metadata map[string]string `json:"metadata"`
		Active   bool              `json:"active"`
		Balance  float64           `json:"balance"`
	}

	// Generate schema and verify it contains Zod 4 compatible syntax
	schema := StructToZodSchema(User{})

	// Test all the core Zod APIs that we generate
	assert.Contains(t, schema, "z.object({", "Should generate z.object syntax")
	assert.Contains(t, schema, "z.string()", "Should generate z.string syntax")
	assert.Contains(t, schema, "z.number()", "Should generate z.number syntax")
	assert.Contains(t, schema, "z.boolean()", "Should generate z.boolean syntax")
	assert.Contains(t, schema, ".optional()", "Should generate .optional() modifier")
	assert.Contains(t, schema, ".nullable()", "Should generate .nullable() modifier")
	assert.Contains(t, schema, ".array()", "Should generate .array() modifier")
	assert.Contains(t, schema, "z.record(", "Should generate z.record syntax")
	assert.Contains(t, schema, "z.infer<typeof", "Should generate z.infer type inference")

	// Verify the complete expected output structure
	expected := `export const UserSchema = z.object({
  name: z.string(),
  age: z.number(),
  email: z.string().optional(),
  tags: z.string().array().nullable(),
  metadata: z.record(z.string(), z.string()).nullable(),
  active: z.boolean(),
  balance: z.number(),
})
export type User = z.infer<typeof UserSchema>

`
	assert.Equal(t, expected, schema)
}

// TestZod4NestedStructCompatibility tests nested struct generation
func TestZod4NestedStructCompatibility(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type User struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	schema := StructToZodSchema(User{})

	// Should generate separate schema for nested struct
	assert.Contains(t, schema, "AddressSchema", "Should generate AddressSchema")
	assert.Contains(t, schema, "export const AddressSchema", "Should export AddressSchema")
	assert.Contains(t, schema, "export type Address", "Should export Address type")
	assert.Contains(t, schema, "address: AddressSchema", "Should reference nested schema")

	expected := `export const AddressSchema = z.object({
  street: z.string(),
  city: z.string(),
})
export type Address = z.infer<typeof AddressSchema>

export const UserSchema = z.object({
  name: z.string(),
  address: AddressSchema,
})
export type User = z.infer<typeof UserSchema>

`
	assert.Equal(t, expected, schema)
}

// TestZod4ComplexTypesCompatibility tests more complex type combinations
func TestZod4ComplexTypesCompatibility(t *testing.T) {
	type Item struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	type User struct {
		Items         []Item             `json:"items"`
		OptionalItems *[]Item            `json:"optional_items,omitempty"`
		ItemMap       map[string]Item    `json:"item_map"`
		Metadata      map[string]string  `json:"metadata"`
		Tags          *[]string          `json:"tags,omitempty"`
		Config        *map[string]string `json:"config,omitempty"`
	}

	schema := StructToZodSchema(User{})

	// Verify complex type combinations work
	assert.Contains(t, schema, "ItemSchema.array()", "Should handle array of structs")
	assert.Contains(t, schema, ".array().optional().nullable()", "Should chain optional and nullable")
	assert.Contains(t, schema, "z.record(z.string(), ItemSchema)", "Should handle map with struct values")
	assert.Contains(t, schema, "z.record(z.string(), z.string())", "Should handle string maps")

	// These patterns should work in both Zod 3 and Zod 4
	expectedPatterns := []string{
		"ItemSchema.array().nullable()",                           // array of structs
		"ItemSchema.array().optional().nullable()",               // optional array of structs
		"z.record(z.string(), ItemSchema).nullable()",            // map with struct values
		"z.record(z.string(), z.string()).nullable()",            // string map
		"z.string().array().optional().nullable()",               // optional string array
		"z.record(z.string(), z.string()).optional().nullable()", // optional string map
	}

	for _, pattern := range expectedPatterns {
		assert.Contains(t, schema, pattern, "Pattern should be present: %s", pattern)
	}
}