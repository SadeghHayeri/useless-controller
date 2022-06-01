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

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	controllers "sotoon.ir/application/utils"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	applicationv1alpha1 "sotoon.ir/application/api/v1alpha1"
)

var HttpPortName = "http"

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var DefaultControllerLabels = map[string]string{
	"controller": "application-controller",
}

func (r *ApplicationReconciler) getResourceLabels(name string) map[string]string {
	return controllers.MergeMaps(
		DefaultControllerLabels,
		map[string]string{
			"controller": "applicationController",
			"app":        name,
		},
	)
}

func (r *ApplicationReconciler) GetIngress(namespaced types.NamespacedName, targetService *corev1.Service, domainSpecs []applicationv1alpha1.ApplicationDomainSpec) *v12.Ingress {
	labels := r.getResourceLabels(namespaced.Name)
	pathType := v12.PathTypePrefix

	var rules []v12.IngressRule
	var tls []v12.IngressTLS
	for _, domainSpec := range domainSpecs {
		rules = append(rules, v12.IngressRule{
			Host: domainSpec.Host,
			IngressRuleValue: v12.IngressRuleValue{
				HTTP: &v12.HTTPIngressRuleValue{
					Paths: []v12.HTTPIngressPath{
						{
							PathType: &pathType,
							Path:     domainSpec.Path,
							Backend: v12.IngressBackend{
								Service: &v12.IngressServiceBackend{
									Name: targetService.Name,
									Port: v12.ServiceBackendPort{
										Name: targetService.Spec.Ports[0].Name, //TODO: Refactor
									},
								},
							},
						},
					},
				},
			},
		})
		tls = append(tls, v12.IngressTLS{
			Hosts:      []string{domainSpec.Host},
			SecretName: domainSpec.Host, // SecretName same as hostname
		})
	}

	return &v12.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespaced.Name,
			Namespace: namespaced.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"cert-manager.io/cluster-issuer": "letsencrypt",
			},
		},
		Spec: v12.IngressSpec{
			IngressClassName: &namespaced.Name,
			Rules:            rules,
			TLS:              tls,
		},
	}
}

func (r *ApplicationReconciler) GetClusterIPService(namespaced types.NamespacedName, selector map[string]string, containerPort *corev1.ContainerPort) *corev1.Service {
	labels := r.getResourceLabels(namespaced.Name)

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespaced.Name,
			Namespace: namespaced.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports: []corev1.ServicePort{
				{
					Name:       HttpPortName,
					Port:       containerPort.ContainerPort, // Service port same as container
					TargetPort: intstr.IntOrString{Type: intstr.String, StrVal: HttpPortName},
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
}

func (r *ApplicationReconciler) GetDeployment(namespaced types.NamespacedName, image string, replicas int32, httpPort int32) *v1.Deployment {
	labels := r.getResourceLabels(namespaced.Name)

	return &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespaced.Name,
			Namespace: namespaced.Namespace,
			Labels:    labels,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									Name:          HttpPortName,
									ContainerPort: httpPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *ApplicationReconciler) ReconcileDeployment(deployment *v1.Deployment, logger *logr.Logger, ctx context.Context, req ctrl.Request) error {
	var currentDeployment v1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &currentDeployment); err != nil {
		logger.Info("Creating Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		if err := r.Create(ctx, deployment); err != nil {
			logger.Error(err, "Failed to create Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
			return err
		}
	}
	return nil
}

func (r *ApplicationReconciler) ReconcileService(service *corev1.Service, logger *logr.Logger, ctx context.Context, req ctrl.Request) error {
	var currentService corev1.Service
	if err := r.Get(ctx, req.NamespacedName, &currentService); err != nil {
		logger.Info("Creating Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		if err := r.Create(ctx, service); err != nil {
			logger.Error(err, "Failed to create Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			return err
		}
	}
	return nil
}

func (r *ApplicationReconciler) ReconcileIngress(ingress *v12.Ingress, logger *logr.Logger, ctx context.Context, req ctrl.Request) error {
	var currentIngress v12.Ingress
	if err := r.Get(ctx, req.NamespacedName, &currentIngress); err != nil {
		logger.Info("Creating Ingress", "Ingress.Namespace", ingress.Namespace, "Ingress.Name", ingress.Name)
		if err := r.Create(ctx, ingress); err != nil {
			logger.Error(err, "Failed to create Ingress", "Ingress.Namespace", ingress.Namespace, "Ingress.Name", ingress.Name)
			return err
		}
	}
	return nil
}

func (r *ApplicationReconciler) UpdateApplicationStatus(application *applicationv1alpha1.Application, logger *logr.Logger, ctx context.Context, req ctrl.Request) error {
	var applicationEndpoints string
	for _, domain := range application.Spec.Domains {
		if domain.SSL.Enabled {
			applicationEndpoints += fmt.Sprintf("https://%s:%d%s, ", domain.Host, application.Spec.HttpPort, domain.Path)
		} else {
			applicationEndpoints += fmt.Sprintf("http://%s:%d%s, ", domain.Host, application.Spec.HttpPort, domain.Path)
		}
	}
	application.Status.Endpoints = strings.TrimSuffix(applicationEndpoints, ", ")

	return r.Status().Update(ctx, application)
}

//+kubebuilder:rbac:groups=application.sotoon.ir,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=application.sotoon.ir,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=application.sotoon.ir,resources=applications/finalizers,verbs=update
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var application applicationv1alpha1.Application
	if err := r.Get(ctx, req.NamespacedName, &application); err != nil {
		logger.Error(err, "Unable to fetch application")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	deployment := r.GetDeployment(req.NamespacedName, application.Spec.Image, application.Spec.Replicas, application.Spec.HttpPort)
	err := r.ReconcileDeployment(deployment, &logger, ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}

	mainContainerPort := &deployment.Spec.Template.Spec.Containers[0].Ports[0]
	service := r.GetClusterIPService(req.NamespacedName, deployment.Spec.Template.Labels, mainContainerPort)
	err = r.ReconcileService(service, &logger, ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}

	ingress := r.GetIngress(req.NamespacedName, service, application.Spec.Domains)
	err = r.ReconcileIngress(ingress, &logger, ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.UpdateApplicationStatus(&application, &logger, ctx, req)
	if err != nil {
		logger.Error(err, "Failed to update application status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&applicationv1alpha1.Application{}).
		Complete(r)
}
