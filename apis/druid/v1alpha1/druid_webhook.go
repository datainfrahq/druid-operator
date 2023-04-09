/*
Copyright 2023.

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
	"fmt"
	"regexp"

	"github.com/friendsofgo/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var druidlog = logf.Log.WithName("druid-resource")

func (r *Druid) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-druid-apache-org-v1alpha1-druid,mutating=false,failurePolicy=fail,sideEffects=None,groups=druid.apache.org,resources=druids,verbs=create;update,versions=v1alpha1,name=vdruid.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Druid{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Druid) ValidateCreate() error {
	druidlog.Info("validate create", "name", r.Name)
	err := r.validateDruidSpec()
	return err
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Druid) ValidateUpdate(old runtime.Object) error {
	druidlog.Info("validate update", "name", r.Name)
	err := r.validateDruidSpec()
	return err
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Druid) ValidateDelete() error {
	druidlog.Info("validate delete", "name", r.Name)
	return nil
}

func (r *Druid) validateDruidSpec() error {
	for key, _ := range r.Spec.Nodes {
		if err := validateKubernetesResourceRegex(key); err != nil {
			return err
		}

		if err := r.validateNodeImage(key); err != nil {
			return err
		}
	}

	return nil
}

func (r *Druid) validateNodeImage(key string) error {
	if r.Spec.Image == "" && r.Spec.Nodes[key].Image == "" {
		errMsg := fmt.Sprintf("Image missing from Druid Cluster Spec and node %s", key)
		return errors.New(errMsg)
	}

	return nil
}

func validateKubernetesResourceRegex(key string) error {
	keyValidationRegex, err := regexp.Compile("[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*")
	if err != nil {
		return err
	}

	if !keyValidationRegex.MatchString(key) {
		errMsg := fmt.Sprintf("Node[%s] Key must match k8s resource name regex '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*'", key)
		return errors.New(errMsg)
	}

	return nil
}
