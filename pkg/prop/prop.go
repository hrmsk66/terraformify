package prop

import (
	"strconv"

	"github.com/hrmsk66/terraformify/pkg/naming"
)

type TFBlock interface {
	GetType() string
	GetID() string
	GetIDforTFImport() string
	GetName() string
	GetNormalizedName() string
	GetRef() string
}

type VCLServiceResource struct {
	ID            string
	Name          string
	TargetVersion int
}

func NewVCLServiceResource(id, name string, targetversion int) *VCLServiceResource {
	return &VCLServiceResource{
		ID:            id,
		Name:          name,
		TargetVersion: targetversion,
	}
}
func (v *VCLServiceResource) GetType() string {
	return "fastly_service_vcl"
}
func (v *VCLServiceResource) GetID() string {
	return v.ID
}
func (v *VCLServiceResource) GetIDforTFImport() string {
	if v.TargetVersion != 0 {
		return v.GetID() + "@" + strconv.Itoa(v.TargetVersion)
	}
	return v.GetID()
}
func (v *VCLServiceResource) GetName() string {
	return v.Name
}
func (v *VCLServiceResource) GetNormalizedName() string {
	// Check if the name can be used as a Terraform resource name
	// If not, falling back to the default resource name
	name := naming.Normalize(v.GetName())

	if !naming.IsValid(name) {
		name = "service"
	}
	return name
}
func (v *VCLServiceResource) GetRef() string {
	return v.GetType() + "." + v.GetNormalizedName()
}

type WAFResource struct {
	*VCLServiceResource
	ID   string
	Name string
}

func NewWAFResource(id string, sr *VCLServiceResource) *WAFResource {
	return &WAFResource{
		VCLServiceResource: sr,
		ID:                 id,
		Name:               "waf",
	}
}
func (w *WAFResource) GetType() string {
	return "fastly_service_waf_configuration"
}
func (w *WAFResource) GetID() string {
	return w.ID
}
func (w *WAFResource) GetIDforTFImport() string {
	return w.GetID()
}
func (w *WAFResource) GetName() string {
	return w.Name
}
func (w *WAFResource) GetNormalizedName() string {
	return naming.Normalize(w.GetName())
}
func (w *WAFResource) GetRef() string {
	return w.GetType() + "." + w.GetNormalizedName()
}

type ACLResource struct {
	*VCLServiceResource
	ID   string
	Name string
	No   int
}

func NewACLResource(id, name string, sr *VCLServiceResource) *ACLResource {
	return &ACLResource{
		VCLServiceResource: sr,
		ID:                 id,
		Name:               name,
	}
}
func (a *ACLResource) GetType() string {
	return "fastly_service_acl_entries"
}
func (a *ACLResource) GetID() string {
	return a.ID
}
func (a *ACLResource) GetIDforTFImport() string {
	return a.VCLServiceResource.GetID() + "/" + a.ID
}
func (a *ACLResource) GetName() string {
	return a.Name
}
func (a *ACLResource) GetNormalizedName() string {
	return naming.Normalize(a.Name)
}
func (a *ACLResource) GetRef() string {
	return a.GetType() + "." + a.GetNormalizedName()
}

type DictionaryResource struct {
	*VCLServiceResource
	ID   string
	Name string
}

func NewDictionaryResource(id, name string, sr *VCLServiceResource) *DictionaryResource {
	return &DictionaryResource{
		VCLServiceResource: sr,
		ID:                 id,
		Name:               name,
	}
}
func (d *DictionaryResource) GetType() string {
	return "fastly_service_dictionary_items"
}
func (d *DictionaryResource) GetID() string {
	return d.ID
}
func (d *DictionaryResource) GetIDforTFImport() string {
	return d.VCLServiceResource.GetID() + "/" + d.ID
}
func (d *DictionaryResource) GetName() string {
	return d.Name
}
func (d *DictionaryResource) GetNormalizedName() string {
	return naming.Normalize(d.GetName())
}
func (d *DictionaryResource) GetRef() string {
	return d.GetType() + "." + d.GetNormalizedName()
}

type DynamicSnippetResource struct {
	*VCLServiceResource
	ID   string
	Name string
}

func NewDynamicSnippetResource(id, name string, sr *VCLServiceResource) *DynamicSnippetResource {
	return &DynamicSnippetResource{
		VCLServiceResource: sr,
		ID:                 id,
		Name:               name,
	}
}
func (ds *DynamicSnippetResource) GetType() string {
	return "fastly_service_dynamic_snippet_content"
}
func (ds *DynamicSnippetResource) GetID() string {
	return ds.ID
}
func (ds *DynamicSnippetResource) GetIDforTFImport() string {
	return ds.VCLServiceResource.GetID() + "/" + ds.ID
}
func (ds *DynamicSnippetResource) GetName() string {
	return ds.Name
}
func (ds *DynamicSnippetResource) GetNormalizedName() string {
	return naming.Normalize(ds.GetName())
}
func (ds *DynamicSnippetResource) GetRef() string {
	return ds.GetType() + "." + ds.GetNormalizedName()
}
