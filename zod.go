package supervillain

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"
)

func StructToZodSchema(input interface{}) string {
	c := converter{
		prefix:  "",
		outputs: make(map[string]entry),
	}

	t := reflect.TypeOf(input)

	c.addSchema(t.Name(), c.convertStructTopLevel(t))

	output := strings.Builder{}
	sorted := []entry{}
	for _, ent := range c.outputs {
		sorted = append(sorted, ent)
	}

	sort.Sort(ByOrder(sorted))

	for _, ent := range sorted {
		output.WriteString(ent.data)
		output.WriteString("\n\n")
	}
	return output.String()
}

func StructToZodSchemaWithPrefix(prefix string, input interface{}) string {
	c := converter{
		prefix:  prefix,
		outputs: make(map[string]entry),
	}

	t := reflect.TypeOf(input)

	c.addSchema(t.Name(), c.convertStructTopLevel(t))

	output := strings.Builder{}
	sorted := []entry{}
	for _, ent := range c.outputs {
		sorted = append(sorted, ent)
	}

	sort.Sort(ByOrder(sorted))

	for _, ent := range sorted {
		output.WriteString(ent.data)
		output.WriteString("\n\n")
	}
	return output.String()
}

var typeMapping = map[reflect.Kind]string{
	reflect.Bool:       "boolean",
	reflect.Int:        "number",
	reflect.Int8:       "number",
	reflect.Int16:      "number",
	reflect.Int32:      "number",
	reflect.Int64:      "number",
	reflect.Uint:       "number",
	reflect.Uint8:      "number",
	reflect.Uint16:     "number",
	reflect.Uint32:     "number",
	reflect.Uint64:     "number",
	reflect.Uintptr:    "number",
	reflect.Float32:    "number",
	reflect.Float64:    "number",
	reflect.Complex64:  "number",
	reflect.Complex128: "number",
	reflect.String:     "string",
}

type entry struct {
	order int
	data  string
}

type ByOrder []entry

func (a ByOrder) Len() int           { return len(a) }
func (a ByOrder) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByOrder) Less(i, j int) bool { return a[i].order < a[j].order }

type converter struct {
	prefix  string
	structs int
	outputs map[string]entry
}

func (c *converter) addSchema(name string, data string) {
	order := c.structs
	c.outputs[name] = entry{order, data}
	c.structs = order + 1
}

func schemaName(prefix, name string) string {
	return fmt.Sprintf("%s%sSchema", prefix, name)
}

func fieldName(input reflect.StructField) string {
	if json := input.Tag.Get("json"); json != "" {
		args := strings.Split(json, ",")
		if len(args[0]) > 0 {
			return args[0]
		}
		// This is also valid:
		// json:",omitempty"
		// so in this case, args[0] will be empty, so fall through to using the
		// raw field name.
	}
	return strcase.ToLowerCamel(strcase.ToSnake(input.Name))
}

func typeName(t reflect.Type) string {
	if t.Kind() == reflect.Struct {
		return t.Name()
	}
	if t.Kind() == reflect.Ptr {
		return typeName(t.Elem())
	}
	if t.Kind() == reflect.Slice {
		return typeName(t.Elem())
	}
	return "UNKNOWN"
}

func (c *converter) convertStructTopLevel(t reflect.Type) string {
	output := strings.Builder{}

	name := t.Name()

	output.WriteString(fmt.Sprintf(
		`export const %s = %s
`,
		schemaName(c.prefix, name), c.convertStruct(t, 0)))

	output.WriteString(fmt.Sprintf(`export type %s%s = z.infer<typeof %s%sSchema>`,
		c.prefix, name, c.prefix, name))

	return output.String()
}

func (c *converter) convertStruct(input reflect.Type, indent int) string {
	output := strings.Builder{}

	output.WriteString(`z.object({
`)

	fields := input.NumField()
	for i := 0; i < fields; i++ {
		field := input.Field(i)
		optional := isOptional(field)
		nullable := isNullable(field)

		line := c.convertField(field, indent+1, optional, nullable)

		output.WriteString(line)
	}

	output.WriteString(indentation(indent))
	output.WriteString(`})`)

	return output.String()
}

func (c *converter) convertType(t reflect.Type, name string, indent int) string {
	if t.Kind() == reflect.Ptr {
		inner := t.Elem()
		return c.convertType(inner, name, indent)
	}

	if t.Kind() == reflect.Slice {
		return fmt.Sprintf(
			"%s.array()",
			c.convertType(t.Elem(), name, indent))
	}

	if t.Kind() == reflect.Struct {
		// Handle nested un-named structs - these are inline.
		if t.Name() == "" {
			return c.convertStruct(t, indent)
		} else if t.Name() == "Time" {
			// timestamps are serialised to strings.
			return "z.string()"
		} else {
			c.addSchema(name, c.convertStructTopLevel(t))
			return schemaName(c.prefix, name)
		}
	}

	if t.Kind() == reflect.Map {
		return c.convertMap(t, name, indent)
	}

	ztype, ok := typeMapping[t.Kind()]
	if !ok {
		panic(fmt.Sprint("cannot handle: ", t.Kind()))
	}

	return fmt.Sprintf("z.%s()", ztype)
}

func (c *converter) convertField(f reflect.StructField, indent int, optional, nullable bool) string {
	name := fieldName(f)

	optionalCall := ""
	if optional {
		optionalCall = ".optional()"
	}
	nullableCall := ""
	if nullable {
		nullableCall = ".nullable()"
	}

	return fmt.Sprintf(
		"%s%s: %s%s%s,\n",
		indentation(indent),
		name,
		c.convertType(f.Type, typeName(f.Type), indent),
		optionalCall,
		nullableCall)
}

func (c *converter) convertMap(t reflect.Type, name string, indent int) string {
	return fmt.Sprintf(`z.map(%s, %s)`,
		c.convertType(t.Key(), name, indent),
		c.convertType(t.Elem(), name, indent))
}

func isNullable(field reflect.StructField) bool {
	// pointers can be nil, which are mapped to null in JS/TS.
	if field.Type.Kind() == reflect.Ptr {
		return true
	}
	// arrays of pointer types may contain null values
	if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Ptr {
		return true
	}
	return false
}

func isOptional(field reflect.StructField) bool {
	// omitempty zero-values are omitted and are mapped to undefined in JS/TS.
	if strings.Contains(field.Tag.Get("json"), "omitempty") {
		return true
	}
	return false
}

func indentation(level int) string {
	return strings.Repeat(" ", level*2)
}
