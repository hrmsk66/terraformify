package tfstate

import (
	"bytes"
	"fmt"
)

// query templates for gojq
const ServiceQueryTmplate = `.resources[] | select(.name == "{{.ResourceName}}") | .instances[].attributes.{{.NestedBlockName}}[] | select(.name == "{{.Name}}") | .{{.AttributeName}}`
const DsnippetQueryTmplate = `.resources[] | select(.name == "{{.ResourceName}}") | .instances[].attributes.content`
const ResourceNameQueryTmplate = `.resources[] | select(.type == "fastly_service_vcl") | .instances[].attributes.{{.NestedBlockName}}[] | select(.{{.IDName}} == "{{.ID}}") | .name`

type ServiceQueryParams struct {
	ResourceName  string
	NestedBlockName string
	Name          string
	AttributeName         string
}

type DSnippetQueryParams struct {
	ResourceName  string
}

type ResourceNameQueryParams struct {
	NestedBlockName string
	IDName        string
	ID            string
}

func (s *TFStateWithTemplate) ServiceQuery(params ServiceQueryParams) (*TFState, error) {
	var q bytes.Buffer
	if err := s.Execute(&q, params); err != nil {
		return nil, fmt.Errorf("tfstate: invalid params: %w", err)
	}

	return s.TFState.Query(q.String())
}

func (s *TFStateWithTemplate) DSnippetQuery(params DSnippetQueryParams) (*TFState, error) {
	var q bytes.Buffer
	if err := s.Execute(&q, params); err != nil {
		return nil, fmt.Errorf("tfstate: invalid params: %w", err)
	}

	return s.TFState.Query(q.String())
}

func (s *TFStateWithTemplate) ResourceNameQuery(params ResourceNameQueryParams) (*TFState, error) {
	var q bytes.Buffer
	if err := s.Execute(&q, params); err != nil {
		return nil, fmt.Errorf("tfstate: invalid params: %w", err)
	}

	return s.TFState.Query(q.String())
}