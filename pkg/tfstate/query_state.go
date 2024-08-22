package tfstate

import (
	"bytes"
	"fmt"
)

// query templates for gojq
const ServiceQueryTmplate = `.resources[] | select(.instances[].attributes.id == "{{.ServiceId}}") | .instances[].attributes.{{.NestedBlockName}}[] | select(.name == "{{.Name}}") | .{{.AttributeName}}`
const DsnippetQueryTmplate = `.resources[] | select(.type == "fastly_service_dynamic_snippet_content") | select(.name == "{{.ResourceName}}") | .instances[].attributes.content`
const ResourceNameQueryTmplate = `.resources[] | select(.type == "{{.ResourceType}}") | .instances[].attributes.{{.NestedBlockName}}[] | select(.{{.IDName}} == "{{.ID}}") | .name`
const RateLimiterContentQueryTemplate = `.resources[] | select(.instances[].attributes.id == "{{.ServiceId}}") | .instances[].attributes.rate_limiter[] | select(.name == "{{.Name}}") | .response[] | .content`

type ServiceQueryParams struct {
	ServiceId       string
	NestedBlockName string
	Name            string
	AttributeName   string
}

type DSnippetQueryParams struct {
	ResourceName string
}

type ResourceNameQueryParams struct {
	ResourceType    string
	NestedBlockName string
	IDName          string
	ID              string
}

type RateLimiterContentQueryParams struct {
	ServiceId string
	Name      string
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

func (s *TFStateWithTemplate) RateLimiterContentQuery(params RateLimiterContentQueryParams) (*TFState, error) {
	var q bytes.Buffer
	if err := s.Execute(&q, params); err != nil {
		return nil, fmt.Errorf("tfstate: invalid params: %w", err)
	}

	return s.TFState.Query(q.String())
}
