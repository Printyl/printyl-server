package models

type DocumentManifest struct {
	Version     int      `yaml:"version"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Template    Template `yaml:"template"`
	TexFile     string   `yaml:"tex_file"`
}

type Type string

const (
	StringType  Type = "string"
	IntegerType Type = "integer"
	MathType    Type = "math"
)

type Fields struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Length      int    `yaml:"length"`
	Multiline   bool   `yaml:"multiline"`
	Type        Type   `yaml:"type"`
	Mandatory   bool   `yaml:"mandatory"`
}

type FieldName string

type Template struct {
	Fields map[FieldName]Fields `yaml:"fields"`
}

// FormResponse represents the response by the server
// It includes fields for every form entry.
type FormResponse struct {
	Fields map[FieldName]FieldsFormResponse `json:"fields"`
}

type FieldsFormResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Length      int    `json:"length"`
	Multiline   bool   `json:"multiline"`
	Type        Type   `json:"type"`
}

func FormResponseFromTemplate(template *Template) FormResponse {
	form := FormResponse{
		Fields: make(map[FieldName]FieldsFormResponse),
	}

	for fieldName, field := range template.Fields {
		form.Fields[fieldName] = FieldsFormResponse{
			Name:        field.Name,
			Description: field.Description,
			Length:      field.Length,
			Multiline:   field.Multiline,
			Type:        field.Type,
		}
	}

	return form
}

type GenerateRequest struct {
	Fields map[FieldName]FieldsFormResponse `json:"data"`
}
