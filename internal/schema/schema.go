package schema

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"strings"

	"gograph/internal/log"

	"github.com/chanced/caps"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/ast"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astnormalization"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astparser"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astprinter"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/asttransform"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astvalidation"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/operationreport"
	"github.com/yargevad/filepathx"
)

type FieldDefinition struct {
	schema *Schema
	ref    int
}

func (f *FieldDefinition) Name() string {
	return f.schema.Name(&f.schema.ast.FieldDefinitions[f.ref].Name)
}

func (f *FieldDefinition) Type() *Type {
	return &Type{schema: f.schema, ref: f.schema.ast.FieldDefinitions[f.ref].Type}
}

type InputValueDefinition struct {
	schema *Schema
	ref    int
}

func (f *InputValueDefinition) Name() string {
	return f.schema.Name(&f.schema.ast.InputValueDefinitions[f.ref].Name)
}

func (f *InputValueDefinition) Type() *Type {
	return &Type{schema: f.schema, ref: f.schema.ast.InputValueDefinitions[f.ref].Type}
}

// A graphql Type ast wrapper
// ----------------------------
type Type struct {
	schema *Schema
	ref    int
}

func (t *Type) Dump() string {

	ofType := t.OfType()
	var child string
	if ofType != nil {
		child = fmt.Sprintf(" -> %v", ofType.Dump())
	}
	var format string
	if t.IsEnum() {
		format = "enum"
	}
	if t.IsList() {
		format = "list"
	}
	if t.IsScalar() {
		format = "scalar"
	}
	var isUnion string
	if t.IsUnion() {
		isUnion = "scalar"
	}

	var nullable string
	if t.IsNonNull() {
		nullable = "nonull"
	}

	return fmt.Sprintf("{Name:%v, Kind:%v Members:%v %v %v %v}%v", t.Name(), t.Kind(), len(t.Members()), format, nullable, child, isUnion)
}

func (t *Type) Members() []FieldDefinition {
	// There must be a better way but i didn't find it
	name := t.Name()
	idx := slices.IndexFunc(t.schema.ast.ObjectTypeDefinitions, func(v ast.ObjectTypeDefinition) bool {
		return t.schema.Name(&v.Name) == name
	})

	if idx > 0 {
		slice := []FieldDefinition{}

		// Found an object definition
		definition := t.schema.ast.ObjectTypeDefinitions[idx]

		for _, fd := range definition.FieldsDefinition.Refs {
			slice = append(slice, FieldDefinition{schema: t.schema, ref: fd})
		}

		return slice
	}
	return nil
}

func (t *Type) TargetType() *Type {
	d := t.schema.ast
	ref := t.ref

	switch d.Types[ref].TypeKind {
	case ast.TypeKindNonNull:
		subt := &Type{schema: t.schema, ref: d.Types[ref].OfType}
		return subt.TargetType()
	case ast.TypeKindNamed:
		return t
	case ast.TypeKindList:
		subt := &Type{schema: t.schema, ref: d.Types[ref].OfType}
		return subt.TargetType()
	default:
		log.Println("Uknown graphql typekind", d.Types[ref].TypeKind)
		return nil
	}
}

func (t *Type) OfType() *Type {
	d := t.schema.ast

	if d.Types[t.ref].OfType >= 0 {
		return &Type{schema: t.schema, ref: d.Types[t.ref].OfType}
	}
	return nil
}

func (t *Type) Name() string {
	return t.schema.ast.TypeNameString(t.ref)
}

func (t *Type) TargetName() string {
	target := t.TargetType()
	if target == t {
		return t.schema.ast.TypeNameString(t.ref)
	}
	return target.TargetName()
}

func (t *Type) Kind() ast.TypeKind {
	return t.schema.ast.Types[t.ref].TypeKind
}

//	func (t *Type) SubType() *Type {
//		return &Type{schema: t.schema, ref: ref].OfType}
//	}
func (t *Type) IsEnum() bool {
	return t.schema.ast.TypeIsEnum(t.ref, t.schema.ast)
}
func (t *Type) IsList() bool {
	return t.schema.ast.TypeIsList(t.ref)
}
func (t *Type) IsNonNull() bool {
	return t.schema.ast.TypeIsNonNull(t.ref)
}
func (t *Type) IsScalar() bool {
	return t.schema.ast.TypeIsScalar(t.ref, t.schema.ast)
}
func (t *Type) IsNamed() bool {
	return t.Kind() == ast.TypeKindNamed
}
func (t *Type) Union() *ast.UnionTypeDefinition {
	name := t.TargetName()
	idx := slices.IndexFunc(t.schema.ast.UnionTypeDefinitions, func(x ast.UnionTypeDefinition) bool {
		return t.schema.Name(&x.Name) == name
	})

	if idx >= 0 {
		unionTypeDef := &t.schema.ast.UnionTypeDefinitions[idx]
		return unionTypeDef
	}
	return nil
}

func (t *Type) UnionMemberType() []Type {
	unionTypeDef := t.Union()
	if unionTypeDef == nil {
		return nil
	}

	slice := []Type{}

	// Found an object definition
	for _, ref := range unionTypeDef.UnionMemberTypes.Refs {
		slice = append(slice, Type{schema: t.schema, ref: ref})
	}

	return slice
}

func (t *Type) IsInput() bool {
	if t.Kind() != ast.TypeKindNamed {
		return false
	}
	name := t.Name()
	idx := slices.IndexFunc(t.schema.ast.InputObjectTypeDefinitions,
		func(a ast.InputObjectTypeDefinition) bool { return t.schema.Name(&a.Name) == name })
	return idx >= 0
}

func (t *Type) InputMembers() []InputValueDefinition {
	slice := []InputValueDefinition{}
	// find the input type for this type
	if t.Kind() != ast.TypeKindNamed {
		return slice
	}
	name := t.Name()
	idx := slices.IndexFunc(t.schema.ast.InputObjectTypeDefinitions,
		func(a ast.InputObjectTypeDefinition) bool { return t.schema.Name(&a.Name) == name })
	if idx < 0 {
		return slice
	}
	inputObjectTypeDef := t.schema.ast.InputObjectTypeDefinitions[idx]

	for _, inputFieldRef := range inputObjectTypeDef.InputFieldsDefinition.Refs {
		slice = append(slice, InputValueDefinition{schema: t.schema, ref: inputFieldRef})
	}
	return slice
}

func (t *Type) IsUnion() bool {
	return t.Union() != nil
}

func (t *Type) String() string {
	writer := bytes.NewBufferString("")
	t.schema.ast.PrintType(t.ref, writer)
	return writer.String()
}

type QuerySelectorOptions struct {
	IgnoreUnderscored bool
	MaxDepth          uint8
}

func Indent(out *strings.Builder, char string, indent uint8, nl bool) {
	if nl {
		out.WriteString("\n")
	}
	for range indent {
		out.WriteString(char)
	}
}

func (argType *Type) Variables(arg *Argument) interface{} {
	var data interface{}
	// TODO: This is a quick hack very messy and not working for anything but the most basic case
	//   - Better default value for scalar with a customizable generator
	//     or guessing for argument name
	//   - Enums
	//   - Better handling/conversion of default values
	//   - Ability to merge user provided values with default values
	//   - ignore @deprecated values

	argOptional := true
	if arg != nil {
		argOptional = arg.IsOptional()
	}

	// No need to generate for non null
	if !argOptional || argType.IsNonNull() {

		// return empty list for list
		if argType.IsList() {
			// log.Println("TYPE IS LIST = ", argType.Name())
			// empty list
			data = []string{}
			return data
		}

		targetType := argType.TargetType()

		if arg != nil && arg.HasDefaultValue() {
			// log.Println("TYPE HAS DEFAULT VALUE = ", argType.Name())
			// TODO: probably need to convert this to an appropriate format
			// string, int, etc..
			data = arg.DefaultValueString()
			return data
		}

		if targetType.IsEnum() {
			// log.Println("TYPE IS ENUM = ", targetType.Name())
			// TODO: Enum default value
			data = "ENUM_TODO"

			return data
		}

		if targetType.IsScalar() {
			// log.Println("TYPE IS SCALAR = ", targetType.Name())

			switch targetType.Name() {
			case "Date":
				data = "" //?
			case "DateTime":
				data = "" //?
			case "Int":
				data = 0
			case "Float":
				data = 0
			case "String":
				data = "String"
			case "Boolean":
				data = false
			case "ID":
				data = "ID"
			default:
				data = ""
			}
			return data
		}

		if targetType.IsInput() {
			inputMembersData := make(map[string]interface{})
			for _, member := range targetType.InputMembers() {
				memberType := member.Type()
				d := memberType.Variables(nil)
				if d != nil {
					inputMembersData[member.Name()] = d
				}
			}
			return inputMembersData
		}

		//
		data = targetType.Variables(nil)
		return data

	}
	return data
}

func (t *Type) StringQuerySelector(out *strings.Builder, opt *QuerySelectorOptions, depth uint8, indent uint8) {

	targetType := t.TargetType()

	if targetType.IsUnion() {
		unionTypes := targetType.UnionMemberType()
		out.WriteString("\n")
		for _, unionType := range unionTypes {

			Indent(out, "  ", indent, true)
			out.WriteString("... on ")
			out.WriteString(unionType.TargetName())
			out.WriteString("{")
			unionType.StringQuerySelector(out, opt, depth, indent+1)
			Indent(out, "  ", indent, true)
			out.WriteString("}")
		}
	} else {

		for _, member := range targetType.Members() {
			memberName := member.Name()

			// Ignore underscored
			if opt.IgnoreUnderscored && memberName[0:2] == "__" {
				continue
			}

			memberTargetType := member.Type().TargetType()

			if !memberTargetType.IsScalar() && !memberTargetType.IsEnum() {
				if depth+1 <= opt.MaxDepth {
					// If the type has no scalar members we might hit into a depth without returning any selectors
					//    so we need to check that there are selector before adding the object
					var memberQuerySelectorStringBuilder strings.Builder
					memberTargetType.StringQuerySelector(&memberQuerySelectorStringBuilder, opt, depth+1, indent+1)
					memberQuerySelectorString := memberQuerySelectorStringBuilder.String()

					if len(strings.TrimSpace(memberQuerySelectorString)) > 0 {
						Indent(out, "  ", indent, true)
						out.WriteString(member.Name())
						out.WriteString(" {")

						out.WriteString(memberQuerySelectorString)

						Indent(out, "  ", indent, true)
						out.WriteString("}")
					}
				}
			} else {

				Indent(out, "  ", indent, true)
				out.WriteString(member.Name())

			}
		}
	}
}

// A graphql Operation Argument ast wrapper
// ----------------------------
type OperationType int

const (
	Query OperationType = iota
	Mutation
)

type Argument struct {
	schema *Schema
	ref    int
}

func (arg *Argument) Name() string {
	return arg.schema.ast.InputValueDefinitionNameString(arg.ref)
}

func (arg *Argument) Type() *Type {
	ref := arg.schema.ast.InputValueDefinitionType(arg.ref)
	return &Type{
		schema: arg.schema,
		ref:    ref,
	}
}

func (arg *Argument) HasDefaultValue() bool {
	return arg.schema.ast.InputValueDefinitionHasDefaultValue(arg.ref)
}

func (arg *Argument) IsOptional() bool {
	return arg.schema.ast.InputValueDefinitionArgumentIsOptional(arg.ref)
}

func (arg *Argument) DefaultValueKind() ast.ValueKind {
	return arg.schema.ast.InputValueDefinitionDefaultValue(arg.ref).Kind
}

func (arg *Argument) DefaultValue() string {
	defaultValue := arg.schema.ast.InputValueDefinitionDefaultValue(arg.ref)
	if defaultValue.Kind <= 0 {
		return ""
	}
	return arg.schema.ast.ValueContentString(defaultValue)
}

func (arg *Argument) DefaultValueString() string {
	defaultValue := arg.schema.ast.InputValueDefinitionDefaultValue(arg.ref)
	if defaultValue.Kind <= 0 {
		return ""
	}
	writer := bytes.NewBufferString("")
	arg.schema.ast.PrintValue(defaultValue, writer)
	return writer.String()
}

// A graphql Operation ast wrapper
// ----------------------------
type Operation struct {
	schema        *Schema
	field         *ast.FieldDefinition
	OperationType OperationType
}

func (o *Operation) Name() string {
	return string(o.schema.ast.Input.ByteSlice(o.field.Name))
}

func (o *Operation) Type() *Type {
	return &Type{
		schema: o.schema,
		ref:    o.field.Type,
	}
}

func (o *Operation) String() string {
	var out strings.Builder
	out.WriteString(o.Name())

	arguments := o.Arguments()
	if len(arguments) > 0 {
		out.WriteString("(")
		for i, arg := range arguments {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(arg.Name())
			out.WriteString(": ")
			out.WriteString(arg.Type().String())
			if arg.HasDefaultValue() {
				def := arg.DefaultValueString()
				out.WriteString(" = ")
				out.WriteString(def)
			}

		}
		out.WriteString(")")
	}
	out.WriteString(": ")
	out.WriteString(o.Type().String())
	return out.String()
}

func (o *Operation) Variables() map[string]interface{} {

	result := make(map[string]interface{})
	arguments := o.Arguments()

	if len(arguments) > 0 {
		for _, arg := range arguments {
			argType := arg.Type()
			result[arg.Name()] = argType.Variables(&arg)
		}
	}

	return result
}

type QueryString struct {
	Name string
	Text string
}

func (o *Operation) QueryString(options *QuerySelectorOptions) *QueryString {
	var out strings.Builder

	var operationType string
	switch o.OperationType {
	case Query:
		operationType = "query"
	case Mutation:
		operationType = "mutation"
	}

	result := &QueryString{
		Name: caps.ToCamel(o.Name()),
		Text: "",
	}

	out.WriteString(operationType)
	out.WriteString(" ")
	out.WriteString(result.Name)

	arguments := o.Arguments()
	if len(arguments) > 0 {
		// argument wrapper
		out.WriteString("(")
		for i, arg := range arguments {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString("$")
			out.WriteString(arg.Name())
			out.WriteString(": ")
			// out.WriteString(arg.Type().String())
			// out.WriteString(arg.Type().TargetType().Name())
			// FIXME: this is a dirty hack
			argTypeDef := arg.Type().String()
			lastChar := string(argTypeDef[len(argTypeDef)-1])

			if !arg.IsOptional() && lastChar != "!" {
				out.WriteString(argTypeDef)
				out.WriteString("!")
			} else if arg.IsOptional() && lastChar == "!" {
				out.WriteString(argTypeDef[0 : len(argTypeDef)-2])
			} else {
				out.WriteString(argTypeDef)
			}

			if arg.HasDefaultValue() {
				def := arg.DefaultValueString()
				out.WriteString(" = ")
				out.WriteString(def)
			}

		}
		out.WriteString(")")
	}
	// Open Query
	out.WriteString("{")
	out.WriteString("\n")
	out.WriteString("  ")
	out.WriteString(o.Name())
	if len(arguments) > 0 {
		out.WriteString("(")
		for i, arg := range arguments {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(arg.Name())
			out.WriteString(": ")
			out.WriteString("$")
			out.WriteString(arg.Name())
		}
		out.WriteString(")")
	}

	if !o.Type().IsScalar() {
		// Beging field selectin
		out.WriteString(" {")

		// out.WriteString(o.Type().String())

		// Write the type selector
		o.Type().StringQuerySelector(&out, options, 0, 2)

		// End field selection
		out.WriteString("\n  }")
	}

	// Close query..
	out.WriteString("\n")
	out.WriteString("}")

	result.Text = out.String()

	return result
}

func (operation *Operation) Arguments() []Argument {
	slice := []Argument{}
	for _, t := range operation.field.ArgumentsDefinition.Refs {
		slice = append(slice, Argument{schema: operation.schema, ref: t})
	}
	return slice
}

// A graphql Schema ast wrapper
// ----------------------------
type Schema struct {
	merged_with_base bool
	normalized       bool
	ast              *ast.Document
}

func (s *Schema) MergeWithBase() {
	if !s.merged_with_base {
		log.Debugf("Merging schema with base")
		err := asttransform.MergeDefinitionWithBaseSchema(s.ast)
		if err != nil {
			log.Fatalln("Merge with base failed", err)
		}
		s.merged_with_base = true
	}
}

func (s *Schema) Normalize() {
	if !s.normalized {

		log.Debugf("Normalizing schema")
		s.MergeWithBase()

		report := &operationreport.Report{}

		// Normalize the schema, this will apply all "extends" to the base type
		// normalizer := astnormalization.NewSubgraphDefinitionNormalizer()
		normalizer := astnormalization.NewDefinitionNormalizer()
		normalizer.NormalizeDefinition(s.ast, report)

		if report.HasErrors() {
			log.Fatalln("Schema normalization failed", report.Error())
		}

		s.normalized = true
	}
}
func (schema *Schema) PrintIndentString() string {
	writer := bytes.NewBufferString("")
	astprinter.PrintIndent(schema.ast, schema.ast, []byte("  "), writer)
	return writer.String()
}

// Save the schema to a file
func (schema *Schema) Save(path string) error {

	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	return astprinter.PrintIndent(schema.ast, schema.ast, []byte("  "), file)
}

func (schema *Schema) Name(pos *ast.ByteSliceReference) string {

	if pos.End > uint32(schema.ast.Input.Length) || pos.Start > pos.End {
		return "" // Return nil slice for invalid range
	}
	return string(schema.ast.Input.ByteSlice(*pos))
}

func (s *Schema) FindOperationByName(operationName string) *Operation {
	operations := s.ListAllOperations(false)
	idx := slices.IndexFunc(operations, func(c Operation) bool { return c.Name() == operationName })
	if idx < 0 {
		return nil
	}
	return &operations[idx]
}

// List the available operations in Query or Mutation
//
// ignoreUnderscored : ignore operation starting with __ such as __typename, __type, etc..
func (schema *Schema) ListAllOperations(ignoreUnderscored bool) []Operation {
	queries := schema.ListOperations(Query, ignoreUnderscored)
	mutations := schema.ListOperations(Mutation, ignoreUnderscored)
	return append(queries, mutations...)
}

func (schema *Schema) ListOperations(operation OperationType, ignoreUnderscored bool) []Operation {

	/*
		BETTER?
		for _, node := range document.RootNodes {
				if node.Kind != ast.NodeKindOperationDefinition {
					continue
				}
				operationCount++
				name := document.RootOperationTypeDefinitionNameString(node.Ref)
				operationNames = append(operationNames, name)
				operationType := document.RootOperationTypeDefinitions[node.Ref].OperationType
				operationTypes = append(operationTypes, operationType)
			}

	*/
	schema.MergeWithBase()

	// Figure out the name for Query and Mutation based on the schema definition
	//   can be empty in which case use default
	var name string
	switch operation {
	case Mutation:
		name = string(schema.ast.Index.MutationTypeName)
		if len(name) == 0 {
			name = "Mutation"
		}
	default:
		name = string(schema.ast.Index.QueryTypeName)
		operation = Query
		if len(name) == 0 {
			name = "Query"
		}
	}

	slice := []Operation{}

	for _, op := range schema.ast.ObjectTypeDefinitions {
		objectName := schema.Name(&op.Name)
		if objectName == name {
			for _, fieldRefIndex := range op.FieldsDefinition.Refs {
				field := schema.ast.FieldDefinitions[fieldRefIndex]

				name := schema.Name(&field.Name)

				if ignoreUnderscored && name[0:2] == "__" {
					continue
				}

				operation := Operation{
					schema:        schema,
					field:         &field,
					OperationType: operation,
				}
				slice = append(slice, operation)
			}
		}
	}

	return slice
}

// Validate a query against the schema
func (schema *Schema) Validate(query string) error {
	schema.Normalize()

	report := &operationreport.Report{}
	document := ast.NewSmallDocument()
	parser := astparser.NewParser()

	document.Input.ResetInputBytes([]byte(query))
	parser.Parse(document, report)

	if report.HasErrors() {
		log.Fatalln("Parse failed", report.Error())
	}

	// fmt.Println("------------VALIDATING---------")
	// fmt.Println(query)
	// fmt.Println("------------/VALIDATING---------")

	// you can customize what rules the normalizer should apply
	normalizer := astnormalization.NewWithOpts(
		astnormalization.WithExtractVariables(),
		astnormalization.WithInlineFragmentSpreads(),
		astnormalization.WithRemoveFragmentDefinitions(),
		astnormalization.WithRemoveNotMatchingOperationDefinitions(),
		astnormalization.WithNormalizeDefinition(),
	)

	// It's generally recommended to always give your operation a name
	// If it doesn't have a name, just add one to the AST before normalizing it
	// This is not strictly necessary, but ensures that all normalization rules work as expected

	// Find the name of the first operation in the document
	//   ==> (NormalizedNamedOperation doesn't work without it)
	queryNodeIdx := slices.IndexFunc(document.RootNodes, func(n ast.Node) bool { return n.Kind == ast.NodeKindOperationDefinition })
	if queryNodeIdx < 0 {
		log.Fatalln("No operation definition in document")
	}
	operationName := document.OperationDefinitionNameString(document.RootNodes[queryNodeIdx].Ref)
	log.Println("Found operation ", operationName)
	if len(operationName) == 0 {
		log.Fatalln("Unable to retrieve operation name")
	}

	normalizer.NormalizeNamedOperation(document, schema.ast, []byte(operationName), report)

	// out, err := astprinter.PrintStringIndent(document, nil, "  ")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("--------NORMALIZED-----------")
	// fmt.Println(out)
	// fmt.Println("--------/NORMALIZED-----------")

	if report.HasErrors() {
		log.Fatalln("Normalize failed", report.Error())
	}

	validator := astvalidation.DefaultOperationValidator()
	validator.Validate(document, schema.ast, report)
	if report.HasErrors() {
		log.Fatalln("Validation failed", report.Error())
	}

	log.Println("Query is valid", report.Error())
	return nil
}

func aggregateFiles(files []string) ([]byte, error) {
	var contentBuilder strings.Builder
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		contentBuilder.WriteString("# ")
		contentBuilder.WriteString(file)
		contentBuilder.WriteString("\n")
		contentBuilder.Write(data)
		contentBuilder.WriteString("\n")

	}
	return []byte(contentBuilder.String()), nil
}

func LoadSchemaFromGlob(pattern string) (*Schema, error) {

	log.Verboseln("Loading schema from", pattern)
	schema, err := LoadSchemaTextFromGlob(pattern)

	if err != nil {
		return nil, err
	}

	report := &operationreport.Report{}

	schemaDocument := ast.NewSmallDocument()
	schemaParser := astparser.NewParser()
	schemaDocument.Input.ResetInputBytes(schema)
	schemaParser.Parse(schemaDocument, report)

	if report.HasErrors() {
		panic(report.Error())
	}

	userSchema := &Schema{ast: schemaDocument}

	userSchema.Normalize()

	log.Verbose("Schema has", len(userSchema.ListAllOperations(true)), "operations")

	if report.HasErrors() {
		panic(report.Error())
	}

	return userSchema, nil
}

func LoadSchemaTextFromGlob(pattern string) ([]byte, error) {

	matches, err := filepathx.Glob(pattern)
	if err != nil {
		log.Println("Error finding files:", err)
		return nil, err
	}

	for _, match := range matches {
		log.Debugf("matched: %v", match)
	}

	content, err := aggregateFiles(matches)

	if err != nil {
		log.Println("Failed to read files", err)
		return nil, err
	}

	return content, nil
}
