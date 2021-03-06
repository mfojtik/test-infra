/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubeadm

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

// ConfigData is supplied to the kubeadm config template, with values populated
// by the cluster package
type ConfigData struct {
	ClusterName       string
	KubernetesVersion string
	// UnifiedControlPlaneImage - optional
	UnifiedControlPlaneImage string
	// AutoDerivedConfigData is populated by DeriveFields()
	AutoDerivedConfigData
}

// AutoDerivedConfigData fields are automatically derived by
// ConfigData.DeriveFieldsif they are not specified / zero valued
type AutoDerivedConfigData struct {
	// DockerStableTag is automatically derived from KubernetesVersion
	DockerStableTag string
}

// DeriveFields automatically derives DockerStableTag if not specified
func (c *ConfigData) DeriveFields() {
	if c.DockerStableTag == "" {
		c.DockerStableTag = strings.Replace(c.KubernetesVersion, "+", "_", -1)
	}
}

// DefaultConfigTemplate is the default kubeadm config template used by kind
const DefaultConfigTemplate = `# config generated by kind
apiVersion: kubeadm.k8s.io/v1alpha2
kind: MasterConfiguration
clusterName: {{.ClusterName}}
# on docker for mac we have to expose the api server via port forward,
# so we need to ensure the cert is valid for localhost so we can talk
# to the cluster after rewriting the kubeconfig to point to localhost
apiServerCertSANs: [localhost]
kubernetesVersion: {{.KubernetesVersion}}
{{if ne .UnifiedControlPlaneImage ""}}
# optionally specify a unified control plane image
unifiedControlPlaneImage: {{.UnifiedControlPlaneImage}}:{{.DockerStableTag}}
{{end}}`

// Config returns a kubeadm config from the template and config data,
// if templateSource == "", DeafultConfigTemplate will be used instead
// ConfigData will be supplied to the template after conversion to ConfigTemplateData
func Config(templateSource string, data ConfigData) (config string, err error) {
	// load the template, using the default if not specified
	if templateSource == "" {
		templateSource = DefaultConfigTemplate
	}
	t, err := template.New("kubeadm-config").Parse(templateSource)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse config template")
	}
	// derive any automatic fields if not supplied
	data.DeriveFields()
	// execute the template
	var buff bytes.Buffer
	err = t.Execute(&buff, data)
	if err != nil {
		return "", errors.Wrap(err, "error executing config template")
	}
	return buff.String(), nil
}
