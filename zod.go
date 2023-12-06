package supervillain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

type Option interface {
	apply(c *Converter)
}

type strictCustomSchemasOption bool

func (s strictCustomSchemasOption) apply(c *Converter) {
	c.strictCustomSchemas = bool(s)
}

func WithStrictCustomSchemas(s bool) Option {
	return strictCustomSchemasOption(s)
}

func NewConverter(custom map[string]CustomFn, opts ...Option) Converter {
	c := Converter{
		prefix:  "",
		outputs: make(map[string]entry),
		custom:  custom,
	}
	for _, opt := range opts {
		opt.apply(&c)
	}

	return c
}

func (c *Converter) Convert(input interface{}) string {
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

func (c *Converter) ConvertSlice(inputs []interface{}) string {
	for _, input := range inputs {
		t := reflect.TypeOf(input)
		c.addSchema(t.Name(), c.convertStructTopLevel(t))
	}
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

func StructToZodSchema(input interface{}, opts ...Option) string {
	c := Converter{
		prefix:  "",
		outputs: make(map[string]entry),
	}

	for _, opt := range opts {
		opt.apply(&c)
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

func StructToZodSchemaWithPrefix(prefix string, input interface{}, opts ...Option) string {
	c := Converter{
		prefix:  prefix,
		outputs: make(map[string]entry),
	}

	for _, opt := range opts {
		opt.apply(&c)
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
	reflect.Interface:  "any",
}

type entry struct {
	order int
	data  string
}

type ByOrder []entry

func (a ByOrder) Len() int           { return len(a) }
func (a ByOrder) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByOrder) Less(i, j int) bool { return a[i].order < a[j].order }

type CustomFn func(*Converter, reflect.Type, string, string, int) string

type Converter struct {
	prefix              string
	structs             int
	outputs             map[string]entry
	custom              map[string]CustomFn
	strictCustomSchemas bool
}

func (c *Converter) addSchema(name string, data string) {
	//First check if the object already exists. If it does do not replace. This is needed for second order
	_, ok := c.outputs[name]
	if !ok {
		order := c.structs
		c.outputs[name] = entry{order, data}
		c.structs = order + 1
	}
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

	// When Golang marshals a struct to JSON and it doesn't have any JSON tags
	// that give the fields names, it defaults to just using the field's name.
	return input.Name
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
	if t.Kind() == reflect.Map {
		return typeName(t.Elem())
	}

	return "UNKNOWN"
}

func (c *Converter) convertStructTopLevel(t reflect.Type) string {
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

func (c *Converter) convertStruct(input reflect.Type, indent int) string {
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

var matchGenericTypeName = regexp.MustCompile(`(.+)\[(.+)\]`)

// checking it a reflected type is a generic isn't supported as far as I can see
// so this simple check looks for a `[` character in the type name: `T1[T2]`.
func isGeneric(t reflect.Type) bool {
	return strings.Contains(t.Name(), "[")
}

// gets the full name and if it's a generic type, strips out the [T] part.
func getFullName(t reflect.Type) (string, string) {
	var typename string
	var generic string

	if isGeneric(t) {
		m := matchGenericTypeName.FindAllStringSubmatch(t.Name(), 1)[0]

		typename = m[1]
		generic = m[2]
	} else {
		typename = t.Name()
	}

	return fmt.Sprintf("%s.%s", t.PkgPath(), typename), generic
}

type ConstantSchema interface {
	ZodSchema() string
}

type DynamicSchema interface {
	ZodSchema(c *Converter, t reflect.Type, name, generic string, indent int) string
}

type DynamicFunctionSchema interface {
	ZodSchema(convert func(t reflect.Type, name string, indent int) string, t reflect.Type, name, generic string, indent int) string
}

func (c *Converter) isCustom(t reflect.Type) bool {
	fullName, _ := getFullName(t)
	_, inMap := c.custom[fullName]
	return (inMap ||
		t.Implements(reflect.TypeOf((*ConstantSchema)(nil)).Elem())) ||
		t.Implements(reflect.TypeOf((*DynamicSchema)(nil)).Elem()) ||
		t.Implements(reflect.TypeOf((*DynamicFunctionSchema)(nil)).Elem())
}

func (c *Converter) handleCustomType(t reflect.Type, name string, indent int) (string, bool) {
	fullName, generic := getFullName(t)

	custom, ok := c.custom[fullName]
	if ok {
		return custom(c, t, name, generic, indent), true
	}

	switch v := reflect.Zero(t).Interface().(type) {
	case ConstantSchema:
		return v.ZodSchema(), true
	case DynamicSchema:
		return v.ZodSchema(c, t, name, generic, indent), true
	case DynamicFunctionSchema:
		return v.ZodSchema(c.ConvertType, t, name, generic, indent), true
	}

	if _, ok := t.MethodByName("ZodSchema"); ok {
		panic(fmt.Sprint("found a ZodSchema method with unexpected signature on type: ", fullName))
	}

	return "", false
}

func (c *Converter) ConvertType(t reflect.Type, name string, indent int) string {
	if t.Kind() == reflect.Ptr {
		inner := t.Elem()
		return c.ConvertType(inner, name, indent)
	}

	if custom, ok := c.handleCustomType(t, name, indent); ok {
		return custom
	}

	fullName, _ := getFullName(t)
	if fullName == "time.Time" {
		// timestamps are serialised to strings.
		return "z.string()"
	}

	if c.strictCustomSchemas && t.Implements(reflect.TypeOf((*json.Marshaler)(nil)).Elem()) {
		panic(fmt.Sprint("found type with custom marshalling but no custom schema: ", fullName))
	}

	if t.Kind() == reflect.Slice {
		return fmt.Sprintf(
			"%s.array()",
			c.ConvertType(t.Elem(), name, indent))
	}

	if t.Kind() == reflect.Struct {
		// Handle nested un-named structs - these are inline.
		if t.Name() == "" {
			return c.convertStruct(t, indent)
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

func (c *Converter) convertField(f reflect.StructField, indent int, optional, nullable bool) string {
	name := fieldName(f)

	// fields named `-` are not exported to JSON so don't export zod types
	if name == "-" {
		return ""
	}

	// because nullability is processed before custom types, this makes sure
	// the custom type has control over nullability.
	isCustom := c.isCustom(f.Type)

	optionalCall := ""
	if optional {
		optionalCall = ".optional()"
	}
	nullableCall := ""
	if nullable && !isCustom {
		nullableCall = ".nullable()"
	}

	return fmt.Sprintf(
		"%s%s: %s%s%s,\n",
		indentation(indent),
		name,
		c.ConvertType(f.Type, typeName(f.Type), indent),
		optionalCall,
		nullableCall)
}

func (c *Converter) convertMap(t reflect.Type, name string, indent int) string {
	return fmt.Sprintf(`z.record(%s, %s)`,
		c.ConvertType(t.Key(), name, indent),
		c.ConvertType(t.Elem(), name, indent))
}

func isNullable(field reflect.StructField) bool {
	// interfaces are currently exported with "any" type, which already includes "null"
	if isInterface(field) {
		return false
	}
	// pointers can be nil, which are mapped to null in JS/TS.
	if field.Type.Kind() == reflect.Ptr {
		// However, if a pointer field is tagged with "omitempty", it usually cannot be exported as "null"
		// since nil is a pointer's empty value.
		if isOptional(field) {
			// Unless it is a pointer to a slice, a map, a pointer, or an interface
			// because values with those types can themselves be nil and will be exported as "null".
			k := field.Type.Elem().Kind()
			return k == reflect.Ptr || k == reflect.Slice || k == reflect.Map
		}
		return true
	}
	// nil slices and maps are exported as null so these types are usually nullable
	if field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Map {
		// unless the are also optional in which case they are no longer nullable
		return !isOptional(field)
	}
	return false
}

// Checks whether the first non-pointer type is an interface
func isInterface(field reflect.StructField) bool {
	t := field.Type
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Interface
}

func isOptional(field reflect.StructField) bool {
	// Non-pointer struct types and direct or indirect interface types should never be optional().
	// Struct fields that are themselves structs ignore the "omitempty" tag because
	// structs do not have an empty value.
	// Interfaces are currently exported with "any" type, which already includes "undefined"
	if field.Type.Kind() == reflect.Struct || isInterface(field) {
		return false
	}
	// Otherwise, omitempty zero-values are omitted and are mapped to undefined in JS/TS.
	return strings.Contains(field.Tag.Get("json"), "omitempty")
}

func indentation(level int) string {
	return strings.Repeat(" ", level*2)
}
