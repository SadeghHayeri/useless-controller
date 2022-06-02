/*
Copyright 2022.

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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var applicationlog = logf.Log.WithName("application-resource")

func (r *Application) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-application-sotoon-ir-v1alpha1-application,mutating=true,failurePolicy=fail,sideEffects=None,groups=application.sotoon.ir,resources=applications,verbs=create;update,versions=v1alpha1,name=mapplication.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Application{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Application) Default() {
	applicationlog.Info("default", "name", r.Name)

	if r.Spec.Replicas == nil {
		applicationlog.Info("(Defaulting) Set application replicas to one")
		one := int32(1)
		r.Spec.Replicas = &one
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-application-sotoon-ir-v1alpha1-application,mutating=false,failurePolicy=fail,sideEffects=None,groups=application.sotoon.ir,resources=applications,verbs=create;update,versions=v1alpha1,name=vapplication.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Application{}

func (r *Application) ValidateReplicas() *field.Error {
	if *r.Spec.Replicas > 5 {
		return field.Invalid(field.NewPath("spec", "replicas"), *r.Spec.Replicas, "replicas must be less than 5")
	}
	return nil
}

func (r *Application) ValidateCreateOrUpdate() field.ErrorList {
	var allErrs field.ErrorList

	if err := r.ValidateReplicas(); err != nil {
		allErrs = append(allErrs, err)
	}

	return allErrs
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Application) ValidateCreate() error {
	applicationlog.Info("validate create", "name", r.Name)
	var allErrs field.ErrorList
	for _, err := range r.ValidateCreateOrUpdate() {
		allErrs = append(allErrs, err)
	}

	//TODO: Create validation here

	if len(allErrs) == 0 {
		return nil
	}
	return errors.NewInvalid(schema.GroupKind{Group: "application.sotoon.ir", Kind: "Application"}, r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Application) ValidateUpdate(old runtime.Object) error {
	applicationlog.Info("validate update", "name", r.Name)
	var allErrs field.ErrorList
	for _, err := range r.ValidateCreateOrUpdate() {
		allErrs = append(allErrs, err)
	}

	//TODO: Update validation here

	if len(allErrs) == 0 {
		return nil
	}
	return errors.NewInvalid(schema.GroupKind{Group: "application.sotoon.ir", Kind: "Application"}, r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Application) ValidateDelete() error {
	applicationlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
