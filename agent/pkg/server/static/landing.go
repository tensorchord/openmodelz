package static

import (
	"bytes"
	_ "embed"
	"html/template"

	"github.com/tensorchord/openmodelz/agent/pkg/version"
)

//go:embed index.html
var htmlTemplate string

type htmlStruct struct {
	Version string
}

func RenderLoadingPage() (*bytes.Buffer, error) {
	tmpl, err := template.New("root").Parse(htmlTemplate)
	if err != nil {
		return nil, err
	}

	data := htmlStruct{
		Version: version.GetAgentVersion(),
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return nil, err
	}

	return &buffer, nil
}
