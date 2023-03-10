package tfstate

import (
	"bytes"
	"fmt"
)

// query and query templates for gojq
const setActivate = `(.resources[] | select(.type == "fastly_service_vcl" or .type == "fastly_service_compute" or .type == "fastly_service_waf_configuration") | .instances[].attributes.activate) |= true`
const setIndexKeyTmplate = `(.resources[] | select(.type == "{{.ResourceType}}") | select(.name == "{{.ResourceName}}") | .instances[]) += {index_key: "{{.Name}}"}`
const setSensitiveAttributeTemplate = `(.resources[] | select(.type == "{{.ResourceType}}") | .instances[].sensitive_attributes) += [[{type: "get_attr", value: "{{.BlockType}}"}]]`
const setManageAttributeTemplate = `(.resources[] | select(.type == "{{.ResourceType}}") | .instances[].attributes.{{.AttributeName}}) |= true`
const setServiceForceDestroyTemplate = `(.resources[] | select(.type == "{{.ResourceType}}") | .instances[].attributes.force_destroy) |= true`
const setACLForceDestroyTemplate = `(.resources[] | select(.type == "{{.ResourceType}}") | .instances[].attributes | .acl[].force_destroy) |= true`
const setDictionaryForceDestroyTemplate = `(.resources[] | select(.type == "{{.ResourceType}}") | .instances[].attributes | .dictionary[].force_destroy) |= true`
const setPackageFilenameTemplate = `(.resources[] | select(.type == "{{.ResourceType}}") | .instances[].attributes.package[]) += {filename: "{{.PackageFilename}}"}`

type SetForceDestroyParams struct {
	ResourceType string
}

type SetIndexKeyParams struct {
	ResourceType string
	ResourceName string
	Name         string
}

type SetPackageFilenameParams struct {
	ResourceType    string
	PackageFilename string
}

type setSensitiveAttributeParams struct {
	ResourceType string
	BlockType    string
}

type setManageAttributeParams struct {
	ResourceType  string
	AttributeName string
}

func (s *TFState) SetActivateAttributes() (*TFState, error) {
	q := setActivate
	return s.Query(q)
}

func (s *TFState) SetIndexKey(param SetIndexKeyParams) (*TFState, error) {
	var q bytes.Buffer

	st, err := s.AddTemplate(setIndexKeyTmplate)
	if err != nil {
		return nil, err
	}

	err = st.Execute(&q, param)
	if err != nil {
		return nil, fmt.Errorf("tfstate: invalid params: %w", err)
	}

	return st.TFState.Query(q.String())
}

func (s *TFState) SetPackageFilename(param SetPackageFilenameParams) (*TFState, error) {
	var q bytes.Buffer

	st, err := s.AddTemplate(setPackageFilenameTemplate)
	if err != nil {
		return nil, err
	}

	err = st.Execute(&q, param)
	if err != nil {
		return nil, fmt.Errorf("tfstate: invalid params: %w", err)
	}

	return st.TFState.Query(q.String())
}

func (s *TFState) SetSensitiveAttributes(resourceType string, blockTypes map[string]struct{}) (*TFState, error) {
	for blockType := range blockTypes {
		var q bytes.Buffer

		st, err := s.AddTemplate(setSensitiveAttributeTemplate)
		if err != nil {
			return nil, err
		}

		err = st.Execute(&q, setSensitiveAttributeParams{
			ResourceType: resourceType,
			BlockType:    blockType,
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

func (s *TFState) SetForceDestroy(param SetForceDestroyParams) (*TFState, error) {
	var q bytes.Buffer
	st, err := s.AddTemplate(setServiceForceDestroyTemplate)
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

	q.Reset()
	st, err = s.AddTemplate(setDictionaryForceDestroyTemplate)
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

	if param.ResourceType == "fastly_service_vcl" {
		q.Reset()
		st, err = s.AddTemplate(setACLForceDestroyTemplate)
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
