package template

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"github.com/iancoleman/strcase"
	version "github.com/jasonhancock/cobra-version"
	"github.com/jasonhancock/go-helpers"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

func (s *Security) GoName() string {
	return strcase.ToCamel(s.Name)
}

func (s *Security) ArgName() string {
	return firstToLower(s.GoName())
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

/*
func (s Security) HasAuthn() bool {
	if s.hasAuthn {
		return true
	}

	if !s.hasAuthn && s.NumArgs == 0 {
		// If they only have 0 args, then it should just be an authorization step
		return false
	}

}
*/

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
					Name:              op.OperationId,
					Path:              path,
					Method:            k,
					SuccessStatusCode: getStatusCode(op.Responses),
					ResponseType:      getResponseType(op),
					Params:            getParams(op),
					RequestBodyType:   getRequestBodyType(op),
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
		Name:        strcase.ToCamel(name),
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
		_, req := required[fieldName]
		m.Fields = append(m.Fields, Field{
			Name:      strcase.ToCamel(fieldName),
			Type:      modelType(v).Type(),
			StructTag: fieldName,
			Required:  req,
		})
	}

	sort.Slice(m.Fields, func(i, j int) bool { return m.Fields[i].Name < m.Fields[j].Name })

	return m, nil
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

func modelTypeName(str string) string {
	switch str {
	case "boolean":
		return "bool"
	default:
		return str
	}
}

type ModelType interface {
	Type() string
}

type BasicModelType string

func (b *BasicModelType) Type() string {
	return string(*b)
}

func newPrimitiveModelType(str string) *BasicModelType {
	b := BasicModelType(modelTypeName(str))
	return &b
}

func newObjectModelType(str string) *BasicModelType {
	b := BasicModelType(strcase.ToCamel(str))
	return &b
}

type SliceModelType string

func (b *SliceModelType) Type() string {
	return "[]" + string(*b)
}

func newSliceModelType(str string) *SliceModelType {
	b := SliceModelType(strcase.ToCamel(str))
	return &b
}

func modelType(schema *base.SchemaProxy) ModelType {
	if schema == nil {
		return newPrimitiveModelType("")
	}

	sch := schema.Schema()

	if sch.Type[0] == "object" {
		return newObjectModelType(strings.TrimPrefix(schema.GetReference(), "#/components/schemas/"))
	}

	if sch.Type[0] == "array" {
		return newSliceModelType(modelType(sch.Items.A).Type())
	}

	// TODO: will need to look into int32 stuff here
	return newPrimitiveModelType(sch.Type[0])
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
	t1, err := template.New("").Parse(tmpl)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := t1.Execute(&buf, data); err != nil {
		log.Fatal(err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		dest.Write(buf.Bytes())
		return fmt.Errorf("error while formatting code...wrote unformatted code to dest. error: %w", err)
	}

	dest.Write(formatted)
	return nil
}

type TemplateData struct {
	GeneratorInfo generatorInfo
	PackageName   string
	Handlers      []Handler
	Models        []Model
	Security      []Security
}

type Model struct {
	Name        string
	Fields      []Field
	Description string
}

type Field struct {
	Name      string
	Type      string
	StructTag string
	Required  bool
}

type Handler struct {
	Name              string
	Path              string
	Method            string
	SuccessStatusCode string
	ResponseType      string
	Params            []Param
	RequestBodyType   string
	Security          *Security
	SecurityArgs      []string
}

func (t TemplateData) SecurityArgs() string {
	var args []string
	for _, v := range t.Security {
		args = append(args, fmt.Sprintf("%s %s", v.ArgName(), v.GoName()))
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
		data = append(data, fmt.Sprintf("%s %s", v.Name, v.Type))
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
			data = append(data, fmt.Sprintf("chi.URLParam(r, `%s`)", v.Name))
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
			"s.router.With(%sPerm%d.Then).%s(`%s`, s.%s)",
			r.Security.ArgName(),
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

// https://stackoverflow.com/a/75989905
func firstToLower(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}
	lc := unicode.ToLower(r)
	if r == lc {
		return s
	}
	return string(lc) + s[size:]
}
