package serviceadapter

import (
	"errors"

	"github.com/pivotal-cf/on-demand-service-broker-sdk/bosh"

	"gopkg.in/go-playground/validator.v8"
)

//go:generate counterfeiter -o fake_service_adapter/manifest_generator.go . ManifestGenerator
type ManifestGenerator interface {
	GenerateManifest(serviceDeployment ServiceDeployment, plan Plan, requestParams RequestParameters, previousManifest *bosh.BoshManifest, previousPlan *Plan) (bosh.BoshManifest, error)
}

//go:generate counterfeiter -o fake_service_adapter/binder.go . Binder
type Binder interface {
	CreateBinding(bindingID string, deploymentTopology bosh.BoshVMs, manifest bosh.BoshManifest, arbitraryParams map[string]interface{}) (Binding, error)
	DeleteBinding(bindingID string, deploymentTopology bosh.BoshVMs, manifest bosh.BoshManifest) error
}

type BindingAlreadyExistsError struct {
	error
}

type AppGuidNotProvidedError struct {
	error
}

func NewBindingAlreadyExistsError(e error) BindingAlreadyExistsError {
	return BindingAlreadyExistsError{e}
}

func NewAppGuidNotProvidedError(e error) AppGuidNotProvidedError {
	return AppGuidNotProvidedError{e}
}

type RequestParameters map[string]interface{}

func (s RequestParameters) ArbitraryParams() map[string]interface{} {
	if s["parameters"] == nil {
		return map[string]interface{}{}
	}
	return s["parameters"].(map[string]interface{})
}

var validate *validator.Validate

func init() {
	config := &validator.Config{TagName: "validate"}
	validate = validator.New(config)
}

type ServiceRelease struct {
	Name    string   `json:"name" validate:"required"`
	Version string   `json:"version" validate:"required"`
	Jobs    []string `json:"jobs" validate:"required,min=1"`
}

type ServiceReleases []ServiceRelease

type ServiceDeployment struct {
	DeploymentName string          `json:"deployment_name" validate:"required"`
	Releases       ServiceReleases `json:"releases" validate:"required"`
	Stemcell       Stemcell        `json:"stemcell" validate:"required"`
}

func (r ServiceReleases) Validate() error {
	if len(r) < 1 {
		return errors.New("no releases specified")
	}

	for _, serviceRelease := range r {
		if err := validate.Struct(serviceRelease); err != nil {
			return err
		}
	}

	return nil
}

type Stemcell struct {
	OS      string `json:"stemcell_os" validate:"required"`
	Version string `json:"stemcell_version" validate:"required"`
}

func (s ServiceDeployment) Validate() error {
	return validate.Struct(s)
}

type Properties map[string]interface{}

type Plan struct {
	Properties     Properties      `json:"properties"`
	InstanceGroups []InstanceGroup `json:"instance_groups" validate:"required,dive"`
}

func (p Plan) Validate() error {
	return validate.Struct(p)
}

type InstanceGroup struct {
	Name           string   `json:"name" validate:"required"`
	VMType         string   `yaml:"vm_type" json:"vm_type" validate:"required"`
	PersistentDisk string   `yaml:"persistent_disk,omitempty" json:"persistent_disk_type,omitempty"`
	Instances      int      `json:"instances" validate:"min=1"`
	Networks       []string `json:"networks" validate:"required"`
	AZs            []string `json:"azs" validate:"required,min=1"`
	Lifecycle      string   `yaml:"lifecycle,omitempty" json:"lifecycle,omitempty"`
}

type Binding struct {
	Credentials     map[string]interface{} `json:"credentials"`
	SyslogDrainURL  string                 `json:"syslog_drain_url,omitempty"`
	RouteServiceURL string                 `json:"route_service_url,omitempty"`
}
