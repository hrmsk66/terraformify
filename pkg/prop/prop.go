package prop

import (
	"errors"
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

type MutatableTfBlock interface {
	TFBlock
	MutateType() error
}

type ComputeServiceResource struct {
	ID            string
	Name          string
	TargetVersion int
}

func NewComputeServiceResource(id, name string, targetVersion int) *ComputeServiceResource {
	return &ComputeServiceResource{
		ID:            id,
		Name:          name,
		TargetVersion: targetVersion,
	}
}
func (c *ComputeServiceResource) GetType() string {
	return "fastly_service_compute"
}
func (c *ComputeServiceResource) GetID() string {
	return c.ID
}
func (c *ComputeServiceResource) GetIDforTFImport() string {
	if c.TargetVersion != 0 {
		return c.GetID() + "@" + strconv.Itoa(c.TargetVersion)
	}
	return c.GetID()
}
func (c *ComputeServiceResource) GetName() string {
	return c.Name
}
func (c *ComputeServiceResource) GetNormalizedName() string {
	// Check if the name can be used as a Terraform resource name
	// If not, falling back to the default resource name
	name := naming.Normalize(c.GetName())

	if !naming.IsValid(name) {
		name = "service"
	}
	return name
}
func (c *ComputeServiceResource) GetRef() string {
	return c.GetType() + "." + c.GetNormalizedName()
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
	ServiceResource TFBlock
	ID              string
	Name            string
}

func NewWAFResource(id string, sr TFBlock) *WAFResource {
	return &WAFResource{
		ServiceResource: sr,
		ID:              id,
		Name:            "waf",
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
	ServiceResource TFBlock
	ID              string
	Name            string
	No              int
}

func NewACLResource(id, name string, sr TFBlock) *ACLResource {
	return &ACLResource{
		ServiceResource: sr,
		ID:              id,
		Name:            name,
	}
}
func (a *ACLResource) GetType() string {
	return "fastly_service_acl_entries"
}
func (a *ACLResource) GetID() string {
	return a.ID
}
func (a *ACLResource) GetIDforTFImport() string {
	return a.ServiceResource.GetID() + "/" + a.ID
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
	ServiceResource TFBlock
	ID              string
	Name            string
}

func NewDictionaryResource(id, name string, sr TFBlock) *DictionaryResource {
	return &DictionaryResource{
		ServiceResource: sr,
		ID:              id,
		Name:            name,
	}
}
func (d *DictionaryResource) GetType() string {
	return "fastly_service_dictionary_items"
}
func (d *DictionaryResource) GetID() string {
	return d.ID
}
func (d *DictionaryResource) GetIDforTFImport() string {
	return d.ServiceResource.GetID() + "/" + d.ID
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
	ServiceResource TFBlock
	ID              string
	Name            string
}

func NewDynamicSnippetResource(id, name string, sr TFBlock) *DynamicSnippetResource {
	return &DynamicSnippetResource{
		ServiceResource: sr,
		ID:              id,
		Name:            name,
	}
}
func (ds *DynamicSnippetResource) GetType() string {
	return "fastly_service_dynamic_snippet_content"
}
func (ds *DynamicSnippetResource) GetID() string {
	return ds.ID
}
func (ds *DynamicSnippetResource) GetIDforTFImport() string {
	return ds.ServiceResource.GetID() + "/" + ds.ID
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

type LinkedResource struct {
	ServiceResource TFBlock
	ID              string
	Name            string
	Type            string
}

var ErrNoMoreResourceType = errors.New("no more linked resource type")
var ErrUnknownResourceType = errors.New("unknown linked resource type")
var ErrNoEntriesToImport = errors.New("no entries to import")

func NewLinkedResource(id, name string, sr TFBlock) *LinkedResource {
	return &LinkedResource{
		ServiceResource: sr,
		ID:              id,
		Name:            name,
		Type:            "fastly_configstore",
	}
}
func (l *LinkedResource) GetType() string {
	return l.Type
}
func (l *LinkedResource) GetID() string {
	return l.ID
}
func (l *LinkedResource) GetIDforTFImport() string {
	return l.ID
}
func (l *LinkedResource) GetName() string {
	return l.Name
}
func (l *LinkedResource) GetNormalizedName() string {
	return naming.Normalize(l.GetName())
}
func (l *LinkedResource) GetRef() string {
	return l.GetType() + "." + l.GetNormalizedName()
}
func (l *LinkedResource) SetDataStoreType(t string) {
	l.Type = t
}
func (l *LinkedResource) MutateType() error {
	switch l.Type {
	case "fastly_configstore":
		l.Type = "fastly_secretstore"
	case "fastly_secretstore":
		l.Type = "fastly_kvstore"
	case "fastly_kvstore":
		return ErrNoMoreResourceType
	default:
		return ErrUnknownResourceType
	}

	return nil
}
func (l *LinkedResource) CloneForEntriesImport() (*LinkedResource, error) {
	switch l.Type {
	case "fastly_configstore":
		return &LinkedResource{
			ServiceResource: l.ServiceResource,
			ID:              l.ID + "/entries",
			Name:            l.Name,
			Type:            l.Type + "_entries",
		}, nil
	default:
		return nil, ErrNoEntriesToImport
	}
}
