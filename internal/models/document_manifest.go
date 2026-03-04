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
}

type FieldName string

type Template struct {
	Fields map[FieldName]Fields `yaml:"fields"`
}
