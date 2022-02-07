/*
Copyright 2021 The Tekton Authors
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

package v1alpha1

import (
	"encoding/json"
	"fmt"
	"github.com/tektoncd/pipeline/pkg/substitution"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/ptr"

	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	triggersv1beta1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genreconciler:krshapedlogic=false
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// Workflow represents a Workflow Custom Resource
type Workflow struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec holds the desired state of the Workflow from the client
	// +optional
	Spec WorkflowSpec `json:"spec,omitempty"`

	// +optional
	Status WorkflowStatus `json:"status,omitempty"`
}

// WorkflowSpec describes the desired state of the Workflow
type WorkflowSpec struct {
	// TODO: Repositories
	Secrets []Secret `json:"secrets,omitempty"`

	Triggers []Trigger `json:"triggers,omitempty"`

	// Params define the default values for params in the Pipeline that can be
	// overridden in a WorkflowRun or (in the future) from an incoming event.
	Params []pipelinev1beta1.ParamSpec `json:"params,omitempty"`

	// Pipeline is a reference to a Pipeline. Currently only an inline
	// pipelineSpec is supported
	Pipeline PipelineRef `json:"pipeline,omitempty"`

	// ServiceAccountName is the K8s service account that pipelineruns
	// generated from this workflow run as
	// +optional
	ServiceAccountName *string `json:"serviceAccountName,omitempty"`

	// Workspaces is a list of workspaces that the Pipeline needs
	// TODO: Auto-setup a Workspace across multiple
	Workspaces []WorkflowWorkspaceBinding `json:"workspaces"`

	// TODO: Timeout ?
	Timeout pipelinev1beta1.TimeoutFields `json:"timeout,omitempty"`
	// TODO: queue_ttl -> pending_timeout
}

// WorkflowStatus describes the observed state of the Workflow
type WorkflowStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkflowList contains a list of Workflows
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workflow `json:"items"`
}

// PipelineRef describes a pipeline
// Only one of the following must be provided
// TODO: Add validation
type PipelineRef struct {
	Spec pipelinev1beta1.PipelineSpec `json:"spec,omitempty"`
}

// WorkflowWorkspaceBinding maps a Pipeline's declared Workspaces
// to a Volume. Unlike a regular WorkspaceBinding, a WorkflowWorkspaceBinding
// will add additional magic to auto-propagate/generate PVCs
// TODO: Fluent Syntax for Binding
type WorkflowWorkspaceBinding struct {
	// TODO: Support Secret Syntax here
	Name                             string `json:"name"`
	Secret                           string `json:"secret,omitempty"`
	pipelinev1beta1.WorkspaceBinding `json:",inline"`
}

type Secret struct {
	Name string `json:"name"`
	Ref  string `json:"ref"`
}

type Trigger struct {
	// +listType=atomic
	Bindings []*triggersv1beta1.TriggerSpecBinding `json:"bindings"`
	// +optional
	Name string `json:"name,omitempty"`

	// TODO: Tackle simplified filters later
	// +listType=atomic
	Interceptors []*triggersv1beta1.TriggerInterceptor `json:"interceptors,omitempty"`
}

func makeWorkspaces(bindings []WorkflowWorkspaceBinding, secrets []Secret) []pipelinev1beta1.WorkspaceBinding {
	res := []pipelinev1beta1.WorkspaceBinding{}
	secretReplacements := map[string]string{}
	for _, s := range secrets {
		secretReplacements[fmt.Sprintf("secrets.%s", s.Name)] = s.Ref
	}

	for _, b := range bindings {
		if b.Secret != "" {
			// Assumes secret name is valid.
			// TODO: Add validation for secret name
			secretName := substitution.ApplyReplacements(b.Secret, secretReplacements)
			res = append(res, pipelinev1beta1.WorkspaceBinding{
				Name: b.Name,
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretName,
					Items:      nil,
					Optional:   ptr.Bool(false),
				},
			})
		} else {
			b.WorkspaceBinding.Name = b.Name
			res = append(res, b.WorkspaceBinding)
		}
	}
	return res
}

// ToPipelineRun converts a Workflow to a PipelineRun.
// Probably should be in its own pkg/resources folder so that it can be reused between
// EL and WorkflowRun, and Workflow reconcilers
func (w *Workflow) ToPipelineRun() (*pipelinev1beta1.PipelineRun, error) {
	saName := "default"
	if w.Spec.ServiceAccountName != nil && *w.Spec.ServiceAccountName != "" {
		saName = *w.Spec.ServiceAccountName
	}

	params := []pipelinev1beta1.Param{}
	for _, ps := range w.Spec.Params {
		params = append(params, pipelinev1beta1.Param{
			Name:  ps.Name,
			Value: *ps.Default,
		})
	}

	return &pipelinev1beta1.PipelineRun{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PipelineRun",
			APIVersion: pipelinev1beta1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-run-", w.Name),
			Namespace:    w.Namespace, // TODO: Do Runs generate from a Workflow always run in the same namespace

			// TODO: Propagate labels/annotations from Workflows as well?
		},
		Spec: pipelinev1beta1.PipelineRunSpec{
			PipelineSpec:       &w.Spec.Pipeline.Spec, // TODO: Apply transforms
			Params:             params,
			ServiceAccountName: saName,
			Timeouts:           &w.Spec.Timeout,
			Workspaces:         makeWorkspaces(w.Spec.Workspaces, w.Spec.Secrets), // TODO: Add workspaces
		},
	}, nil
}

func (w *Workflow) ToTriggerTemplate() (*triggersv1beta1.TriggerTemplate, error) {
	pr, err := w.ToPipelineRun()
	if err != nil {
		return nil, err
	}

	params := []triggersv1beta1.ParamSpec{}
	for _, p := range w.Spec.Params {
		// TODO: Triggers does not support array values
		if p.Type == pipelinev1beta1.ParamTypeArray {
			continue
		}

		params = append(params, triggersv1beta1.ParamSpec{
			Name:        p.Name,
			Description: p.Description,
			Default:     ptr.String(p.Default.StringVal),
		})
		for i, prp := range pr.Spec.Params {
			if prp.Name == p.Name {
				pr.Spec.Params[i].Value.StringVal = fmt.Sprintf("$(tt.params.%s)", prp.Name)
				pr.Spec.Params[i].Value.Type = pipelinev1beta1.ParamTypeString
			}
		}
	}

	// TODO: Once we add trigger-bindings, we need to match on binding param names
	// and replace the values with the ones from binding
	prJson, err := json.Marshal(pr)
	if err != nil {
		return nil, err
	}

	tt := &triggersv1beta1.TriggerTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TriggerTemplate",
			APIVersion: triggersv1beta1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("tt-%s", w.Name),
			Namespace: w.Namespace,
		},
		Spec: triggersv1beta1.TriggerTemplateSpec{
			Params: params,
			// Look in triggers code base for what this should look like
			ResourceTemplates: []triggersv1beta1.TriggerResourceTemplate{{
				RawExtension: runtime.RawExtension{
					Raw: prJson,
				},
			}},
		},
	}

	return tt, nil
}

// ToTriggers creates a new Trigger with inline bindings and template for each type
// TODO: Reuse same triggertemplate for efficiency?
func (w *Workflow) ToTriggers() ([]triggersv1beta1.Trigger, error) {
	tt, err := w.ToTriggerTemplate()
	if err != nil {
		return nil, err
	}
	triggers := []triggersv1beta1.Trigger{}
	for i, t := range w.Spec.Triggers {
		name := t.Name
		if name == "" {
			// FIXME: What if user re-orders triggers
			// Name field should always exist -> Add it in defautls
			name = string(i)
		}

		triggers = append(triggers, triggersv1beta1.Trigger{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Trigger",
				APIVersion: triggersv1beta1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-%s", w.Name, name),
				// Trigger is created in the namespace of the EL for easier RBAC
				// The SA needs roles to create in any repo though
				Labels: map[string]string{
					"managed-by": "tekton-workflows",
				},
			},
			Spec: triggersv1beta1.TriggerSpec{
				Bindings: t.Bindings,
				Template: triggersv1beta1.TriggerSpecTemplate{
					Spec: &tt.Spec,
				},
				Name:         name,
				Interceptors: t.Interceptors,
			},
		})
	}
	return triggers, nil
}