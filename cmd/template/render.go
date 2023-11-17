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

const (
	extensionGoType        = "x-go-type"
	extensionGoImport      = "x-go-import"
	extensionGoImportAlias = "x-go-import-alias"
)

type generatorInfo struct {
	Name    string
	Version string
}

type Security struct {
	Name    string
	NumArgs int

	hasAuthn bool

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
	"int32":   {},
	"int64":   {},
	"float32": {},
	"float64": {},
}

// typeName returns a camel cased typed name. Good for identifiers and types.
func typeName(str string) string {
	if _, isPrimitive := primitiveTypes[str]; isPrimitive {
		return str
	}

	return snaker.ForceCamelIdentifier(str)
}

// argName returns a lower cased version of an identifier, useful for unexported
// variable names and names of arguments to functions.
func argName(str string) string {
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
	//return firstToLower(typeName(str))
}

func (s *Security) AuthzArgs() string {
	if s.NumArgs == 0 {
		return ""
	}
	var args []string
	for i := 1; i <= s.NumArgs; i++ {
		args = append(args, fmt.Sprintf("arg%d", i))
	}

	return strings.Join(args, ", ") + " string"
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

func pp(data any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	enc.Encode(data)
}

func templateDataFrom(input *libopenapi.DocumentModel[v3high.Document], packageName string, info version.Info) (TemplateData, error) {
	data := TemplateData{
		PackageName: packageName,
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
					Params:             getParams(op),
					RequestBodyType:    getRequestBodyType(op),
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
							if len(secArgs) > 0 {
								sec.AddArgPermutation(secArgs)
							}

							h.Security = sec
							h.SecurityArgs = secArgs
						}
					}
				}

				data.Handlers = append(data.Handlers, h)
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

	sort.Slice(data.Handlers, func(i, j int) bool { return data.Handlers[i].Name < data.Handlers[j].Name })
	sort.Slice(data.Models, func(i, j int) bool { return data.Models[i].Name < data.Models[j].Name })
	sort.Slice(data.Security, func(i, j int) bool { return data.Security[i].Name < data.Security[j].Name })

	return data, nil
}

func getModel(name string, s *base.SchemaProxy) (Model, error) {
	schema := s.Schema()

	m := Model{
		Name:        typeName(name),
		Description: schema.Description,
	}

	if len(schema.Type) != 1 {
		return Model{}, errors.New("expected schema.Type to have len == 1")
	}

	if schema.Type[0] != "object" {
		// figure out what to do here.
		return Model{}, errors.New("Temmporary error: not an object")
	}

	required := make(map[string]struct{}, len(schema.Required))
	for _, fieldName := range schema.Required {
		required[fieldName] = struct{}{}
	}

	for fieldName, v := range schema.Properties {
		dataType := modelType(v).Type()
		goType, goImport := getGoTypeAndImport(v.Schema().Extensions)

		if goType != "" {
			dataType = goType
			m.AddImport(goImport)
		}

		_, req := required[fieldName]
		m.Fields = append(m.Fields, Field{
			Name:      typeName(fieldName),
			Type:      dataType,
			StructTag: fieldName,
			Required:  req,
		})
	}

	sort.Slice(m.Fields, func(i, j int) bool { return m.Fields[i].Less(m.Fields[j]) })

	return m, nil
}

func (f Field) Less(other Field) bool {
	if f.Name == "ID" {
		return true
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
		// TODO: log a warning that a x-go-type was specified, but not the import
		return "", Import{}
	}
	impStr, ok := imp.(string)
	if !ok {
		// TODO: log a warning that a x-go-type was specified, but not the import
		return "", Import{}
	}

	var impAlias string
	impA, ok := extensions[extensionGoImportAlias]
	if ok {
		impAlias, ok = impA.(string)
		if !ok {
			// TODO: log a warning that x-go-import-alias was specified, but not a string
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
		params = append(params, Param{
			Name:     v.Name,
			Type:     modelType(v.Schema).Type(),
			Location: v.In,
		})

	}

	return params
}

type ModelType interface {
	Type() string
}

type BasicModelType string

func (b *BasicModelType) Type() string {
	return string(*b)
}

func newPrimitiveModelType(str string) *BasicModelType {
	b := BasicModelType(str)
	return &b
}

func newObjectModelType(str string) *BasicModelType {
	b := BasicModelType(typeName(str))
	return &b
}

type SliceModelType string

func (b *SliceModelType) Type() string {
	return "[]" + string(*b)
}

func newSliceModelType(str string) *SliceModelType {
	b := SliceModelType(typeName(str))
	return &b
}

type MapModelType string

func newMapModelType(str string) *MapModelType {
	b := MapModelType(typeName(str))
	return &b
}

func (b *MapModelType) Type() string {
	return "map[string]" + string(*b)
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
			return newMapModelType(sch.AdditionalProperties.A.Schema().Type[0])
		}
		return newObjectModelType(strings.TrimPrefix(schema.GetReference(), "#/components/schemas/"))
	}

	if sch.Type[0] == "array" {
		return newSliceModelType(modelType(sch.Items.A).Type())
	}

	switch sch.Type[0] {
	case "boolean":
		return newPrimitiveModelType("bool")
	case "integer":
		switch sch.Format {
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

func getResponseType(op *v3high.Operation) string {
	if op.Responses == nil {
		return ""
	}

	for code, r := range op.Responses.Codes {
		if !strings.HasPrefix(code, "2") {
			continue
		}

		j, ok := r.Content["application/json"]
		if !ok {
			return ""
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

func renderTemplate(tmpl string, data TemplateData, dest io.Writer) error {
	funcs := sprig.FuncMap()
	funcs["typename"] = typeName
	funcs["argname"] = argName

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

type TemplateData struct {
	GeneratorInfo generatorInfo
	PackageName   string
	Handlers      []Handler
	Models        Models
	Security      []Security
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

// TODO: we may want to consider how to handle import aliases and stuff.
func (m *Model) AddImport(imp Import) {
	if m.imports == nil {
		m.imports = make(map[Import]struct{})
	}

	m.imports[imp] = struct{}{}
}

type Field struct {
	Name      string
	Type      string
	StructTag string
	Required  bool
}

type Handler struct {
	op                 *v3high.Operation
	Name               string
	Path               string
	Method             string
	SuccessStatusCode  string
	SuccessContentType string
	ResponseType       string
	Params             []Param
	RequestBodyType    string
	Security           *Security
	SecurityArgs       []string
}

func (h Handler) Description() string {
	return h.op.Description
}

func (h Handler) ExportedName() string {
	return typeName(h.Name)
}

func (h Handler) ParameterizedURI() (string, error) {
	pathParams := make(map[string]Param)

	for _, p := range h.Params {
		if p.Location != "path" {
			continue
		}

		pathParams[fmt.Sprintf(`{%s}`, p.Name)] = p
	}

	pieces := strings.Split(h.Path, "/")
	var paramList []string
	for i := range pieces {
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
		case "int", "int32", "int64":
			pieces[i] = `%d`
		case "bool":
			pieces[i] = `%t`
		default:
			return "", errors.New("path parameter support is currently limited to strings and ints")
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
			Handler:      h.Name,
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
		data = append(data, fmt.Sprintf("%s %s", argName(v.Name), v.Type))
	}

	if h.RequestBodyType != "" {
		data = append(data, "req "+h.RequestBodyType)
	}

	return strings.Join(data, ", "), nil
}

func (h Handler) ValueList() (string, error) {
	data := []string{"r.Context()"}
	for _, v := range h.Params {
		switch v.Location {
		case "query":
			// TODO: this only works right now on strings. Will need to put type validation in
			data = append(data, fmt.Sprintf("r.URL.Query().Get(`%s`)", v.Name))
		case "path":
			//data = append(data, fmt.Sprintf("chi.URLParam(r, `%s`)", v.Name))
			if v.Type == "int32" {
				data = append(data, fmt.Sprintf("int32(%s)", argName(v.Name)))
			} else {
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

	return strings.Join(data, ", "), nil
}

type Param struct {
	Name     string
	Type     string
	Location string
}

//go:embed partials/param_int.txt
var partialParseInt string

func (p Param) PathAssignment() (string, error) {
	if p.Location != "path" {
		return "", errors.New("called PathAssignment on non path param")
	}

	switch p.Type {
	case "string":
		return fmt.Sprintf("%s := chi.URLParam(r, `%s`)", argName(p.Name), p.Name), nil
	case "int32":
		return fmt.Sprintf(partialParseInt, argName(p.Name), p.Name, 32), nil
	case "int", "int64":
		return fmt.Sprintf(partialParseInt, argName(p.Name), p.Name, 64), nil
	default:
		return "", fmt.Errorf("PathAssignment called with unsupported type %s", p.Type)
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
