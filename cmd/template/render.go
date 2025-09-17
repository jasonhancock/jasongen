package template

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	version "github.com/jasonhancock/cobra-version"
	"github.com/jasonhancock/go-helpers"
	"github.com/kenshaw/snaker"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/imports"
)

func init() {
	if err := snaker.DefaultInitialisms.Add("GID"); err != nil {
		panic(fmt.Errorf("adding GID to initialisms: %w", err))
	}
}

const (
	extensionGoType           = "x-go-type"
	extensionGoImport         = "x-go-import"
	extensionGoImportAlias    = "x-go-import-alias"
	extensionGoPropertyNames  = "x-go-property-names"
	extensionGoDoNotSerialize = "x-go-do-not-serialize"
	extensionRetrievalName    = "x-retrieval-name"
)

type generatorInfo struct {
	Name    string
	Version string
}

type Security struct {
	Name    string
	NumArgs int

	ArgPermutations []argPermutation
}

type argPermutation struct {
	Hash string
	Args []string
}

func (a argPermutation) ArgStr() string {
	return quotedStrings(a.Args...)
}

func quotedStrings(strs ...string) string {
	if len(strs) == 0 {
		return ""
	}

	return `"` + strings.Join(strs, `", "`) + `"`
}

var primitiveTypes = map[string]struct{}{
	"any":     {},
	"bool":    {},
	"string":  {},
	"int8":    {},
	"int16":   {},
	"int32":   {},
	"int64":   {},
	"float32": {},
	"float64": {},
	"[]byte":  {},
}

// typeName returns a camel cased typed name. Good for identifiers and types.
func typeName(str string) string {
	if _, isPrimitive := primitiveTypes[str]; isPrimitive {
		return str
	}

	if strings.Contains(str, ".") {
		// qualified type name
		return str
	}

	return snaker.ForceCamelIdentifier(str)
}

var reserved = map[string]struct{}{
	"type": {},
}

// argName returns a lower cased version of an identifier, useful for unexported
// variable names and names of arguments to functions.
func argName(str string) string {
	if !strings.Contains(str, "_") {
		// looks like it's not snake case.

		lc := helpers.LCFirst(str)
		if _, ok := reserved[lc]; ok {
			return "_" + lc
		}
		return lc
	}

	// This could likely be improved. Right now it really only supports snake_case
	pieces := strings.Split(str, "_")
	for i := range pieces {
		if i == 0 {
			pieces[i] = strings.ToLower(pieces[i])
			continue
		}

		pieces[i] = snaker.ForceCamelIdentifier(pieces[i])
	}

	return strings.Join(pieces, "")
}

func snake(str string) string {
	return snaker.CamelToSnake(snaker.ForceCamelIdentifier(str))
}

func (s *Security) AuthzArgs() string {
	return "args ...string"
}

func (s *Security) AddArgPermutation(args []string) {
	b, _ := json.Marshal(args)
	perm := helpers.MD5(b)

	for _, v := range s.ArgPermutations {
		if v.Hash == perm {
			return
		}
	}

	s.ArgPermutations = append(s.ArgPermutations, argPermutation{
		Hash: perm,
		Args: args,
	})

	sort.Slice(s.ArgPermutations, func(i, j int) bool {
		fromI := strings.Join(s.ArgPermutations[i].Args, "|")
		fromJ := strings.Join(s.ArgPermutations[j].Args, "|")
		return fromI < fromJ
	})
}

func (s *Security) GetPermutationIndex(args []string) (int, error) {
	b, _ := json.Marshal(args)
	perm := helpers.MD5(b)

	for i := range s.ArgPermutations {
		if s.ArgPermutations[i].Hash == perm {
			return i, nil
		}
	}

	return -1, fmt.Errorf("argument permutation %q not found", strings.Join(args, ", "))
}

func templateDataFrom(
	input *libopenapi.DocumentModel[v3high.Document],
	packageName string,
	info version.Info,
	opts cmdOptions,
) (TemplateData, error) {
	data := TemplateData{
		PackageName: packageName,
		PkgModels:   opts.pkgModels,
		GeneratorInfo: generatorInfo{
			Name:    filepath.Base(os.Args[0]),
			Version: info.Version,
		},
	}

	discoveredSecurity := make(map[string]*Security)

	if input.Model.Paths != nil {
		for path := range input.Model.Paths.PathItems {
			pi := input.Model.Paths.PathItems[path]
			for k, op := range pi.GetOperations() {
				h := Handler{
					Name:               op.OperationId,
					op:                 op,
					Path:               path,
					Method:             k,
					SuccessStatusCode:  getStatusCode(op.Responses),
					SuccessContentType: getSuccessContentType(op.Responses),
					ResponseType:       getResponseType(op),
					ErrorResponseTypes: getErrorResponses(op),
					Params:             getParams(op),
					RequestBodyType:    getRequestBodyType(op),
					PkgModels:          opts.pkgModels,
					IsFileDownload:     getIsFileDownload(op.Responses),
				}

				if h.IsFileDownload {
					data.HasFileDownloads = true
					h.ResponseType = "*FileDownloadResponse"
				}

				if len(op.Security) > 0 {
					if len(op.Security) > 1 {
						return TemplateData{}, errors.New("more than one security thing not supported yet")
					}
					for _, v := range op.Security {
						if len(v.Requirements) > 1 {
							return TemplateData{}, errors.New("more than one security requirements not supported yet")
						}

						for secName, secArgs := range v.Requirements {
							if _, ok := discoveredSecurity[secName]; !ok {
								discoveredSecurity[secName] = &Security{
									Name:    secName,
									NumArgs: len(secArgs),
								}
							}

							sec := discoveredSecurity[secName]

							if sec.NumArgs == 0 && len(secArgs) != 0 {
								discoveredSecurity[secName].NumArgs = len(secArgs)
							} else if sec.NumArgs != len(secArgs) && len(secArgs) != 0 {
								return TemplateData{},
									fmt.Errorf(
										"inconsistent number of arguments for security %q. expected=%d actual=%d",
										secName,
										sec.NumArgs,
										len(secArgs),
									)
							}
							//if len(secArgs) > 0 {
							sec.AddArgPermutation(secArgs)
							//}

							h.Security = sec
							h.SecurityArgs = secArgs
						}
					}
				}

				data.Handlers = append(data.Handlers, h)

				if h.Params.HasQuery() {
					data.Models = append(data.Models, h.Params.buildQueryParamsModel(h.Name))
				}
			}
		}
	}

	if input.Model.Components != nil {
		for name, val := range input.Model.Components.Schemas {
			model, err := getModel(name, val)
			if err != nil {
				return TemplateData{}, err
			}

			data.Models = append(data.Models, model)
		}
	}

	data.Security = make([]Security, 0, len(discoveredSecurity))
	for _, v := range discoveredSecurity {
		data.Security = append(data.Security, *v)
	}

	sort.Slice(data.Handlers, func(i, j int) bool { return data.Handlers[i].ExportedName() < data.Handlers[j].ExportedName() })
	sort.Slice(data.Models, func(i, j int) bool { return data.Models[i].Name < data.Models[j].Name })
	sort.Slice(data.Security, func(i, j int) bool { return data.Security[i].Name < data.Security[j].Name })

	return data, nil
}

func mapStringString(input map[string]any) (map[string]string, error) {
	data := make(map[string]string, len(input))
	for k, v := range input {
		vStr, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("key %q not a string", k)
		}
		data[k] = vStr
	}
	return data, nil
}

func getModel(name string, s *base.SchemaProxy) (Model, error) {
	schema := s.Schema()

	m := Model{
		Name:        typeName(name),
		Description: schema.Description,
	}
	fieldNameMappings := make(map[string]string)
	if fieldNames, ok := s.Schema().Extensions[extensionGoPropertyNames]; ok {
		mappings, ok := fieldNames.(map[string]any)
		if !ok {
			return Model{}, fmt.Errorf("%s: %s not a map", name, extensionGoPropertyNames)
		}

		var err error
		fieldNameMappings, err = mapStringString(mappings)
		if err != nil {
			return Model{}, fmt.Errorf("%s: %w", name, err)
		}
	}

	if len(schema.Type) != 1 {
		return Model{}, errors.New("expected schema.Type to have len == 1 " + name)
	}

	if schema.Type[0] != "object" {
		// figure out what to do here.
		// specifically, this is for enums. The go-polaris-authapi uses enums and is a good place to figure out how to deal with these.
		return Model{}, errors.New("Temporary error: not an object " + name + " " + schema.Type[0])
	}

	required := make(map[string]struct{}, len(schema.Required))
	for _, fieldName := range schema.Required {
		required[fieldName] = struct{}{}
	}

	for fieldName, v := range schema.Properties {
		dataType := modelType(v).Type()
		goType, goImport := getGoTypeAndImport(v.Schema().Extensions)

		var noPointer bool
		switch mt := modelType(v).(type) {
		case *SliceModelType:
			noPointer = true
			m.AddImport(mt.Imports()...)
		case *MapModelType:
			noPointer = true
			m.AddImport(mt.Imports()...)
		default:
		}

		if goType != "" {
			dataType = goType
			m.AddImport(goImport)
		}

		_, req := required[fieldName]

		typeNameStr := fieldName
		if name, ok := fieldNameMappings[fieldName]; ok {
			typeNameStr = name
		}

		m.Fields = append(m.Fields, Field{
			Name:           typeName(typeNameStr),
			Type:           dataType,
			StructTag:      fieldName,
			Required:       req,
			NoPointer:      noPointer,
			DoNotSerialize: getDoNotSerialize(v.Schema().Extensions),
		})
	}

	sort.Slice(m.Fields, func(i, j int) bool {
		return m.Fields[i].Less(m.Fields[j])
	})

	return m, nil
}

func (f Field) Less(other Field) bool {
	if f.Name == "ID" {
		return true
	}

	if other.Name == "ID" {
		return false
	}

	if f.Name == "CreatedAt" && other.Name != "UpdatedAt" {
		return false
	}

	if f.Name == "UpdatedAt" {
		return false
	}

	if other.Name == "CreatedAt" {
		return true
	}

	return f.Name < other.Name
}

type Import struct {
	Package string
	Alias   string
}

func (i Import) String() string {
	if i.Alias == "" {
		return `"` + i.Package + `"`
	}

	return i.Alias + ` "` + i.Package + `"`
}

func getDoNotSerialize(extensions map[string]any) bool {
	s, ok := extensions[extensionGoDoNotSerialize]
	if !ok {
		return false
	}

	sBool, ok := s.(bool)
	if !ok {
		panic(extensionGoDoNotSerialize + " was set, but not to a boolean value")
	}

	return sBool
}

func getGoTypeAndImport(extensions map[string]any) (string, Import) {
	t, ok := extensions[extensionGoType]
	if !ok {
		return "", Import{}
	}
	dataType, ok := t.(string)
	if !ok {
		return "", Import{}
	}

	imp, ok := extensions[extensionGoImport]
	if !ok {
		panic("x-go-type was specified, but not an import")
	}
	impStr, ok := imp.(string)
	if !ok {
		panic("x-go-type was specified, but was not a string")
	}

	var impAlias string
	impA, ok := extensions[extensionGoImportAlias]
	if ok {
		impAlias, ok = impA.(string)
		if !ok {
			panic("x-go-import alias was specified, but not a string")
		}
	}

	return dataType, Import{
		Package: impStr,
		Alias:   impAlias,
	}
}

func getParams(op *v3high.Operation) []Param {
	params := make([]Param, 0, len(op.Parameters))

	for _, v := range op.Parameters {
		p := Param{
			Name:     v.Name,
			Type:     modelType(v.Schema).Type(),
			Location: v.In,
			Required: v.Required,
		}

		if rName, ok := v.Extensions[extensionRetrievalName]; ok {
			p.RetrievalName = rName.(string)
		}

		params = append(params, p)
	}

	return params
}

type ModelType interface {
	Type() string
	Imports() []Import
}

type BasicModelType string

func (b *BasicModelType) Type() string {
	return string(*b)
}

func (b *BasicModelType) Imports() []Import {
	return nil
}

func newPrimitiveModelType(str string) *BasicModelType {
	b := BasicModelType(str)
	return &b
}

func newObjectModelType(str string) *BasicModelType {
	b := BasicModelType(typeName(str))
	return &b
}

type SliceModelType struct {
	Items ModelType
}

func (b *SliceModelType) Type() string {
	return "[]" + b.Items.Type()
}

func (b *SliceModelType) Imports() []Import {
	return b.Items.Imports()
}

func newSliceModelType(items ModelType) *SliceModelType {
	return &SliceModelType{Items: items}
}

type MapModelType struct {
	Items ModelType
}

func newMapModelType(items ModelType) *MapModelType {
	return &MapModelType{Items: items}
}

func (b *MapModelType) Type() string {
	return "map[string]" + b.Items.Type()
}

func (b *MapModelType) Imports() []Import {
	return b.Items.Imports()
}

type importedModelType struct {
	Name   string
	Import Import
}

func newImportedModelType(name string, imp Import) *importedModelType {
	return &importedModelType{
		Name:   name,
		Import: imp,
	}
}

func (i *importedModelType) Type() string {
	return i.Name
}

func (i *importedModelType) Imports() []Import {
	return []Import{i.Import}
}

func modelType(schema *base.SchemaProxy) ModelType {
	if schema == nil {
		return newPrimitiveModelType("")
	}

	sch := schema.Schema()

	if len(sch.Type) == 0 {
		// The response type wasn't indicated.
		return newPrimitiveModelType("any")
	}

	if sch.Type[0] == "object" {
		if sch.AdditionalProperties != nil {
			// we have a map!
			if sch.AdditionalProperties.N == 1 && sch.AdditionalProperties.B {
				return newMapModelType(newPrimitiveModelType("any"))
			}
			// TODO: this probably has the same problem as slices where we need to detect the object type and act appropriately
			return newMapModelType(newPrimitiveModelType(sch.AdditionalProperties.A.Schema().Type[0]))
		}
		return newObjectModelType(strings.TrimPrefix(schema.GetReference(), "#/components/schemas/"))
	}

	if sch.Type[0] == "array" {
		dataType := modelType(sch.Items.A).Type()

		/*
			log.Println("jason....: " + dataType)

			l := logger.Default() // TODO: fix this
			goType, goImport := getGoTypeAndImport(l, sch.Items.A.Schema().Extensions)

			if goType != "" {
				dataType = goType

				// Something to solve is how to add the goImport to the model. I don't have a good answer yet.
				log.Println("TODO: ", goImport)
				//m.AddImport(goImport)
			}
		*/

		return newSliceModelType(newPrimitiveModelType(dataType))
	}

	switch sch.Type[0] {
	case "boolean":
		return newPrimitiveModelType("bool")
	case "integer":
		switch sch.Format {
		case "int8":
			return newPrimitiveModelType("int8")
		case "int16":
			return newPrimitiveModelType("int16")
		case "int32":
			return newPrimitiveModelType("int32")
		case "int64":
			return newPrimitiveModelType("int64")
		default:
			return newPrimitiveModelType("int64")
		}
	case "number":
		switch sch.Format {
		case "float":
			return newPrimitiveModelType("float32")
		case "double":
			return newPrimitiveModelType("float64")
		default:
			return newPrimitiveModelType("float64")
		}
	case "string":
		goType, goImport := getGoTypeAndImport(sch.Extensions)
		if goType != "" {
			return newImportedModelType(goType, goImport)
		}

		return newPrimitiveModelType("string")
	default:
		return newPrimitiveModelType(sch.Type[0])
	}
}

func getStatusCode(resp *v3high.Responses) string {
	if resp == nil {
		return "http.StatusOK"
	}

	for code := range resp.Codes {
		if !strings.HasPrefix(code, "2") {
			continue
		}
		return statusStringToName(code)
	}
	return "http.StatusOK"
}

func getSuccessContentType(resp *v3high.Responses) string {
	if resp == nil {
		return ""
	}

	for code, r := range resp.Codes {
		if !strings.HasPrefix(code, "2") {
			continue
		}

		// we'll prefer application/json
		if _, ok := r.Content["application/json"]; ok {
			return "application/json"
		}

		types := make([]string, 0, len(r.Content))
		for k := range r.Content {
			types = append(types, k)
		}
		sort.Strings(types)
		if len(types) > 0 {
			return types[0]
		}
	}

	return ""
}

func getIsFileDownload(resp *v3high.Responses) bool {
	// If there's a json response defined, then it's not a file download endpoint.
	if getSuccessContentType(resp) == "application/json" {
		return false
	}

	// Iterate through the successful response. If a response exists where type=string
	// and format=binary, then it's a file download, regardless of content type.
	for code, r := range resp.Codes {
		if !strings.HasPrefix(code, "2") {
			continue
		}

		for _, v := range r.Content {
			if v.Schema == nil {
				continue
			}

			s, err := v.Schema.BuildSchema()
			if s == nil || err != nil {
				continue
			}

			if helpers.Contains(s.Type, "string") && s.Format == "binary" {
				return true
			}
		}
	}

	return false
}

type errorResponse struct {
	Code string
	Type string
}

func getErrorResponses(op *v3high.Operation) []errorResponse {
	data := make([]errorResponse, 0)
	if op.Responses == nil {
		return data
	}

	for code, r := range op.Responses.Codes {
		if strings.HasPrefix(code, "2") {
			continue
		}

		j, ok := r.Content["application/json"]
		if !ok {
			continue
		}

		data = append(data, errorResponse{
			Code: code,
			Type: modelType(j.Schema).Type(),
		})
	}

	sort.Slice(data, func(i, j int) bool { return data[i].Code < data[j].Code })

	return data
}

func getResponseType(op *v3high.Operation) string {
	if op.Responses == nil {
		return ""
	}
	if getIsFileDownload(op.Responses) {
		return "foo,bar"
	}

	for code, r := range op.Responses.Codes {
		if !strings.HasPrefix(code, "2") {
			continue
		}

		j, ok := r.Content["application/json"]
		if !ok {
			if len(r.Content) == 0 {
				return ""
			}
			return "[]byte"
		}

		return modelType(j.Schema).Type()
	}

	return ""
}

func getRequestBodyType(op *v3high.Operation) string {
	if op.RequestBody == nil {
		return ""
	}

	if _, ok := op.RequestBody.Content["application/json"]; !ok {
		return ""
	}

	return modelType(op.RequestBody.Content["application/json"].Schema).Type()
}

func renderTemplate(tmpl string, data TemplateData, dest io.Writer, pkgModels string) error {
	funcs := sprig.FuncMap()
	funcs["typename"] = typeName
	funcs["argname"] = argName
	funcs["snake"] = snake
	funcs["httpstatus"] = statusStringToName
	funcs["models"] = models(pkgModels)

	t1, err := template.New("").Funcs(funcs).Parse(tmpl)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := t1.Execute(&buf, data); err != nil {
		return err
	}

	formatted, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		dest.Write(buf.Bytes())
		return fmt.Errorf("error while formatting code...wrote unformatted code to dest. error: %w", err)
	}

	_, err = dest.Write(formatted)
	return err
}

func models(pkgModels string) func(string) string {
	return func(in string) string {
		if pkgModels == "" {
			return in
		}
		if _, ok := primitiveTypes[in]; ok {
			return in
		}
		return "models." + in
	}
}

type TemplateData struct {
	GeneratorInfo    generatorInfo
	PackageName      string
	Handlers         []Handler
	Models           Models
	Security         []Security
	PkgModels        string
	HasFileDownloads bool
}

type Models []Model

// Imports returns the list of custom imports used by the models.
func (m Models) Imports() []Import {
	seen := make(map[Import]struct{})

	for _, mod := range m {
		for imp := range mod.imports {
			seen[imp] = struct{}{}
		}
	}

	uniques := make([]Import, 0, len(seen))
	for k := range seen {
		uniques = append(uniques, k)
	}

	sort.Slice(uniques, func(i, j int) bool {
		if uniques[i].Package == uniques[j].Package {
			return uniques[i].Alias < uniques[j].Alias
		}
		return uniques[i].Package < uniques[j].Package
	})
	return uniques
}

type Model struct {
	Name        string
	Fields      []Field
	Description string
	imports     map[Import]struct{}
}

func (m *Model) AddImport(imports ...Import) {
	if m.imports == nil {
		m.imports = make(map[Import]struct{})
	}

	for _, imp := range imports {
		m.imports[imp] = struct{}{}
	}
}

type Field struct {
	Name           string
	Type           string
	StructTag      string
	Required       bool
	NoPointer      bool
	DoNotSerialize bool
}

type Handler struct {
	op                 *v3high.Operation
	Name               string
	Path               string
	Method             string
	SuccessStatusCode  string
	SuccessContentType string
	ResponseType       string
	Params             Params
	RequestBodyType    string
	Security           *Security
	SecurityArgs       []string
	ErrorResponseTypes []errorResponse
	PkgModels          string
	IsFileDownload     bool
}

func (h Handler) Comment() string {
	return helpers.LCFirst(h.Description())
}

type Params []Param

func (p Params) HasQuery() bool {
	for _, v := range p {
		if v.Location == "query" {
			return true
		}
	}
	return false
}

// handlerName should be the operationId (not the typeName) of the handler.
func (p Params) buildQueryParamsModel(handlerName string) Model {
	m := Model{
		Name:        typeName(handlerName + "_params"),
		Description: "Query parameters for " + typeName(handlerName),
	}
	for _, v := range p {
		if v.Location != "query" {
			continue
		}
		m.Fields = append(m.Fields, v.Field())
	}

	return m
}

func (h Handler) Description() string {
	return h.op.Description
}

func (h Handler) ExportedName() string {
	return typeName(h.Name)
}

func (h Handler) UnexportedName() string {
	return argName(typeName(h.Name))
}

func (h Handler) ParameterizedURI() (string, error) {
	pathParams := make(map[string]Param)
	wildcardRetrievalName := ""

	for _, p := range h.Params {
		if p.Location != "path" {
			continue
		}

		if p.RetrievalName == "*" {
			wildcardRetrievalName = p.Name
		}

		pathParams[fmt.Sprintf(`{%s}`, p.Name)] = p
	}

	pieces := strings.Split(h.Path, "/")
	var paramList []string
	for i := range pieces {
		if pieces[i] == "*" && wildcardRetrievalName != "" {
			pieces[i] = "{" + wildcardRetrievalName + "}"
		}

		if !strings.HasPrefix(pieces[i], "{") || !strings.HasSuffix(pieces[i], "}") {
			continue
		}

		pParam, ok := pathParams[pieces[i]]
		if !ok {
			return "", fmt.Errorf(
				"path parameter %q found in uri, but not in parameters list",
				strings.TrimSuffix(strings.TrimPrefix(pieces[i], "{"), "}"),
			)
		}

		switch pParam.Type {
		case "string":
			pieces[i] = `%s`
		case "int", "int8", "int16", "int32", "int64":
			pieces[i] = `%d`
		case "bool":
			pieces[i] = `%t`
		default:
			return "", fmt.Errorf(
				"path parameter support is currently limited to strings and ints (%q not supported, path=%q method=%q)",
				pParam.Type,
				h.Path,
				h.Method,
			)
		}
		paramList = append(paramList, argName(pParam.Name))
	}

	if len(paramList) == 0 {
		return `"` + h.Path + `"`, nil
	}

	str := strings.Join(pieces, "/")
	str = fmt.Sprintf(`fmt.Sprintf("%s", %s)`, str, strings.Join(paramList, ", "))

	return str, nil
}

func (t TemplateData) SecurityArgs() string {
	var args []string
	for _, v := range t.Security {
		args = append(args, fmt.Sprintf("%s %s", argName(v.Name), typeName(v.Name)))
	}

	return strings.Join(args, ", ")
}

func (t TemplateData) Routes() []Route {
	var routes []Route
	for _, h := range t.Handlers {
		routes = append(routes, Route{
			Path:         h.Path,
			Handler:      h.UnexportedName(),
			Method:       h.Method,
			Security:     h.Security,
			SecurityArgs: h.SecurityArgs,
		})
	}

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	return routes
}

func (h Handler) TypeList() (string, error) {
	data := []string{"ctx context.Context"}
	for _, v := range h.Params {
		if v.Location != "path" {
			continue
		}
		data = append(data, fmt.Sprintf("%s %s", argName(v.Name), v.Type))
	}

	if h.RequestBodyType != "" {
		dataType := h.RequestBodyType
		if h.PkgModels != "" {
			dataType = "models." + dataType
		}
		data = append(data, "req "+dataType)
	}
	if h.Params.HasQuery() {
		data = append(data, fmt.Sprintf("qp %s", typeName(h.Name+"_params")))
	}

	return strings.Join(data, ", "), nil
}

func (h Handler) ValueList(contextFromRequest bool) (string, error) {
	var data []string
	if contextFromRequest {
		data = append(data, "r.Context()")
	} else {
		data = append(data, "ctx")
	}
	for _, v := range h.Params {
		switch v.Location {
		case "query":
			// TODO: this only works right now on strings. Will need to put type validation in
			//data = append(data, fmt.Sprintf("r.URL.Query().Get(`%s`)", v.Name))
		case "path":
			//data = append(data, fmt.Sprintf("chi.URLParam(r, `%s`)", v.Name))
			switch v.Type {
			case "int8":
				data = append(data, fmt.Sprintf("int8(%s)", argName(v.Name)))
			case "int16":
				data = append(data, fmt.Sprintf("int16(%s)", argName(v.Name)))
			case "int32":
				data = append(data, fmt.Sprintf("int32(%s)", argName(v.Name)))
			default:
				data = append(data, argName(v.Name))
			}
		case "body":
			data = append(data, "req")
		default:
			return "", fmt.Errorf("unknown location %q", v.Location)
		}
	}

	if h.RequestBodyType != "" {
		data = append(data, "req")
	}

	if h.Params.HasQuery() {
		data = append(data, "qp")
	}

	return strings.Join(data, ", "), nil
}

type Param struct {
	Name          string
	Type          string
	Location      string
	Required      bool
	RetrievalName string
}

//go:embed partials/param_int.txt
var partialParseInt string

func (p Param) Field() Field {
	return Field{
		Name:     typeName(p.Name),
		Type:     p.Type,
		Required: p.Required,
	}
}

func (p Param) PathAssignment() (string, error) {
	if p.Location != "path" {
		return "", errors.New("called PathAssignment on non path param")
	}

	name := p.Name
	if p.RetrievalName != "" {
		name = p.RetrievalName
	}

	switch p.Type {
	// TODO: support bool, floats
	case "string":
		return fmt.Sprintf("%s := chi.URLParam(r, `%s`)", argName(p.Name), name), nil
	case "int8":
		return fmt.Sprintf(partialParseInt, argName(p.Name), name, 8), nil
	case "int16":
		return fmt.Sprintf(partialParseInt, argName(p.Name), name, 16), nil
	case "int32":
		return fmt.Sprintf(partialParseInt, argName(p.Name), name, 32), nil
	case "int", "int64":
		return fmt.Sprintf(partialParseInt, argName(p.Name), name, 64), nil
	default:
		return "", fmt.Errorf("PathAssignment called with unsupported type %s", p.Type)
	}
}

func (p Param) FormattingFunc() (string, error) {
	str := "p." + typeName(p.Name)
	if !p.Required {
		str = "*" + str
	}

	switch p.Type {
	case "string":
		return str, nil
	case "int", "int8", "int16", "int32", "int64":
		return `fmt.Sprintf("%d", ` + str + `)`, nil
	case "float32", "float64":
		return `fmt.Sprintf("%f", ` + str + `)`, nil
	case "bool":
		return `fmt.Sprintf("%t", ` + str + `)`, nil
	default:
		return "", fmt.Errorf("Param.FormattingFunc called with unsupported type %s", p.Type)
	}
}

type Route struct {
	Path         string
	Method       string
	Handler      string
	Security     *Security
	SecurityArgs []string
}

func (r Route) GetRoute() (string, error) {
	if r.Security != nil {
		idx, err := r.Security.GetPermutationIndex(r.SecurityArgs)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf(
			"s.router.With(%sAuthzPerm%d.Then).%s(`%s`, s.%s)",
			argName(r.Security.Name),
			idx,
			methodFunc(r.Method),
			r.Path,
			r.Handler,
		), nil
	}

	return fmt.Sprintf(
			"s.router.%s(`%s`, s.%s)",
			methodFunc(r.Method),
			r.Path,
			r.Handler,
		),
		nil
}

func methodFunc(method string) string {
	return cases.Title(language.English).String(strings.ToLower(method))
}

/*
// lowerFirstChar returns the string with the first character lowercased.
func lowerFirstChar(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(string(s[0])) + s[1:]
}
*/
