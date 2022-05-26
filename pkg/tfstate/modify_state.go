package tfstate

import (
	"bytes"
	"fmt"
)

// query templates for gojq
const setActivate = `(.resources[] | select(.type == "fastly_service_vcl" or .type == "fastly_service_waf_configuration") | .instances[].attributes.activate) |= true`
const setManageAttributeTemplate = `(.resources[] | select(.type == "{{.ResourceType}}") | .instances[].attributes.{{.AttributeName}}) |=true`
const setIndexKeyTmplate = `(.resources[] | select(.type == "{{.ResourceType}}") | select(.name == "{{.ResourceName}}") | .instances[]) += {index_key: "{{.Name}}"}`
const setSensitiveAttributeTemplate = `(.resources[] | select(.type == "{{.ResourceType}}") | .instances[].sensitive_attributes) += [[{type: "get_attr", value: "{{.BlockType}}"}]]`

type setManageAttributeParams struct {
	ResourceType string
	AttributeName string
}

type setSensitiveAttributeParams struct {
	ResourceType string
	BlockType string
}

type SetIndexKeyParams struct {
	ResourceType string
	ResourceName string
	Name         string
}

func (s *TFState) SetIndexKey(param SetIndexKeyParams) (*TFState, error) {
	var q bytes.Buffer

	st, err := s.AddTemplate(setIndexKeyTmplate)
	if err != nil {
		return nil, err
	}

	err = st.Execute(&q, SetIndexKeyParams{
		ResourceType: param.ResourceType,
		ResourceName: param.ResourceName,
		Name: param.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("tfstate: invalid params: %w", err)
	}

	return st.TFState.Query(q.String())
}

func (s *TFState) SetSensitiveAttributes(blockTypes map[string]struct{}) (*TFState, error) {
	for blockType := range blockTypes {
		var q bytes.Buffer

		st, err := s.AddTemplate(setSensitiveAttributeTemplate)
		if err != nil {
			return nil, err
		}

		err = st.Execute(&q, setSensitiveAttributeParams{
			ResourceType: "fastly_service_vcl",
			BlockType: blockType,
		})

		if err != nil {
			return nil, fmt.Errorf("tfstate: invalid params: %w", err)
		}

		s, err = st.TFState.Query(q.String())
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *TFState) SetManageAttributes() (*TFState, error) {
	params := []setManageAttributeParams{
		{"fastly_service_dynamic_snippet_content", "manage_snippets"},
		{"fastly_service_dictionary_items", "manage_items"},
		{"fastly_service_acl_entries", "manage_entries"},
	}

	for _, param := range params {
		var q bytes.Buffer

		st, err := s.AddTemplate(setManageAttributeTemplate)
		if err != nil {
			return nil, err
		}

		err = st.Execute(&q, param)
		if err != nil {
			return nil, fmt.Errorf("tfstate: invalid params: %w", err)
		}

		s, err = st.TFState.Query(q.String())
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *TFState) SetActivateAttributes() (*TFState, error) {
	q := setActivate
	return s.Query(q)
}
