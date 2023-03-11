package tfstate

import (
	"bytes"
	"fmt"
)

// query templates for gojq
const setActivateTemplate = `(.resources[] | select(.instances[].attributes.id == "{{.ServiceId}}") | .instances[].attributes.activate) |= true`
const setActivateWAFTemplate = `(.resources[] | select(.instances[].attributes.id == "{{.WafId}}") | .instances[].attributes.activate) |= true`
const setIndexKeyTmplate = `(.resources[] | select(.type == "{{.ResourceType}}") | select(.instances[].attributes.service_id == "{{.ServiceId}}") | select(.name == "{{.ResourceName}}") | .instances[]) += {index_key: "{{.Name}}"}`
const setSensitiveAttributeTemplate = `(.resources[] | select(.instances[].attributes.id == "{{.ServiceId}}") | .instances[].sensitive_attributes) += [[{type: "get_attr", value: "{{.BlockType}}"}]]`
const setManageAttributeTemplate = `(.resources[] | select(.type == "{{.ResourceType}}") | select(.instances[].attributes.service_id == "{{.ServiceId}}") | .instances[].attributes.{{.AttributeName}}) |= true`
const setServiceForceDestroyTemplate = `(.resources[] | select(.instances[].attributes.id == "{{.ServiceId}}") | .instances[].attributes.force_destroy) |= true`
const setACLForceDestroyTemplate = `(.resources[] | select(.instances[].attributes.id == "{{.ServiceId}}") | .instances[].attributes | .acl[].force_destroy) |= true`
const setDictionaryForceDestroyTemplate = `(.resources[] | select(.instances[].attributes.id == "{{.ServiceId}}") | .instances[].attributes | .dictionary[].force_destroy) |= true`
const setPackageFilenameTemplate = `(.resources[] | select(.instances[].attributes.id == "{{.ServiceId}}") | .instances[].attributes.package[]) += {filename: "{{.PackageFilename}}"}`

type SetActivateWAFTemplateParams struct {
	WafId string
}

type SetActivateTemplateParams struct {
	ServiceId string
}

type SetForceDestroyParams struct {
	ServiceId    string
	ResourceType string
}

type SetIndexKeyParams struct {
	ServiceId    string
	ResourceType string
	ResourceName string
	Name         string
}

type SetPackageFilenameParams struct {
	ServiceId       string
	PackageFilename string
}

type setSensitiveAttributeParams struct {
	ServiceId string
	BlockType string
}

type setManageAttributeParams struct {
	ServiceId     string
	ResourceType  string
	AttributeName string
}

func (s *TFState) SetActivateWAFAttribute(param SetActivateWAFTemplateParams) (*TFState, error) {
	var q bytes.Buffer

	st, err := s.AddTemplate(setActivateWAFTemplate)
	if err != nil {
		return nil, err
	}

	err = st.Execute(&q, param)
	if err != nil {
		return nil, fmt.Errorf("tfstate: invalid params: %w", err)
	}

	return st.TFState.Query(q.String())
}

func (s *TFState) SetActivateAttribute(param SetActivateTemplateParams) (*TFState, error) {
	var q bytes.Buffer

	st, err := s.AddTemplate(setActivateTemplate)
	if err != nil {
		return nil, err
	}

	err = st.Execute(&q, param)
	if err != nil {
		return nil, fmt.Errorf("tfstate: invalid params: %w", err)
	}

	return st.TFState.Query(q.String())
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

func (s *TFState) SetSensitiveAttributes(serviceId string, blockTypes map[string]struct{}) (*TFState, error) {
	for blockType := range blockTypes {
		var q bytes.Buffer

		st, err := s.AddTemplate(setSensitiveAttributeTemplate)
		if err != nil {
			return nil, err
		}

		err = st.Execute(&q, setSensitiveAttributeParams{
			ServiceId: serviceId,
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

func (s *TFState) SetManageAttributes(serviceId string) (*TFState, error) {
	params := []setManageAttributeParams{
		{serviceId, "fastly_service_dynamic_snippet_content", "manage_snippets"},
		{serviceId, "fastly_service_dictionary_items", "manage_items"},
		{serviceId, "fastly_service_acl_entries", "manage_entries"},
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
