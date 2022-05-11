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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	mngiov1 "mng.io/test-msi/api/mng.io/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// MemberClusterMembershipReconciler reconciles a MemberClusterMembership object
type MemberClusterMembershipReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mng.io,resources=memberclustermemberships,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mng.io,resources=memberclustermemberships/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mng.io,resources=memberclustermemberships/finalizers,verbs=update

func (r *MemberClusterMembershipReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	scope := "https://graph.microsoft.com/.default"
	hubServerUrl := "https://demohub-caravel-dev-test-4be892-7e18dd01.hcp.eastus.azmk8s.io:443"

	l.Info("Starting reconciliation loop")
	cred, err := azidentity.NewClientSecretCredential("72f988bf-86f1-41af-91ab-2d7cd011db47", "12461fbc-e437-4ac0-b2c2-b55e4d0199b0", "d1Fe7RL2~1jFFXWWLcIpVCPvbEp-WCTY0F", nil)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "fail getting credentials")
	}
	l.Info("Azure credentials ", "cred", cred)

	t, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{scope},
	})
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "fail getting token")
	}
	l.Info("Current access token ", "token", t)

	var demoCluster mngiov1.MemberClusterMembership
	if err := r.Get(ctx, req.NamespacedName, &demoCluster); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "fail getting demo kind")
	}
	l.Info("Creating kubeconfig", "hubServerUrl", hubServerUrl)

	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "fail creating cert")
	}
	cf := rest.Config{
		BearerToken: t.Token,
		Host:        hubServerUrl,
	}

	k8sClient, err := client.New(&cf, client.Options{Scheme: r.Scheme})

	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "fail creating client")
	}

	demoCluster.Spec.Foo = "updatedFoo"
	if err := k8sClient.Update(ctx, &demoCluster); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "error updating hub cluster")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MemberClusterMembershipReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mngiov1.MemberClusterMembership{}).
		Complete(r)
}
