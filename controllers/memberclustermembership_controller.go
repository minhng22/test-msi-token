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
	"crypto/tls"
	"crypto/x509"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/pkg/errors"
	v13 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
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
	l.Info("Current access token ", "token", (*t).Token)

	var demoCluster mngiov1.MemberClusterMembership
	if err := r.Get(ctx, req.NamespacedName, &demoCluster); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "fail getting demo kind")
	}
	l.Info("Creating kubeconfig", "hubServerUrl", hubServerUrl)

	// This ca cert comes from the field certificate-authority-data in the member cluster kubeconfig
	caCert := []byte("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZIVENDQXdXZ0F3SUJBZ0lRRCs4SEsrenlVRnkxcVVLczFoUGZpakFOQmdrcWhraUc5dzBCQVFzRkFEQU4KTVFzd0NRWURWUVFERXdKallUQWVGdzB5TWpBMU1UQXdNakV5TXpSYUZ3MHlOREExTVRBd01qSXlNelJhTURBeApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1SVXdFd1lEVlFRREV3eHRZWE4wWlhKamJHbGxiblF3CmdnSWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUNEd0F3Z2dJS0FvSUNBUUN5LzVtdHN0RG9ieFVvQkhnSktzQkMKK1pxOXBhSW53QjhiNlFwbWxTMDVaSngzSTVEcGxNQzlEeGQ0ajk4MlZLWFAwUUdvdUtiNEFHakQ0VkJSTy9pNwp1eXJJRkIwTDRJUFNxVTdaRXBJVW91WmhqcXRUbnJrUDZQcDJxVFgwRTltWDFJaHlpaGprTVB1MDZSSUZtUEVoClFaVzdxRDc4OWhjTVJTN3hjRS9QTEQzMkphYnd4ZURoa08yV2tnMm9lUWtLN0NHRXQxVWNnWFo2aXFWbzJZZnAKWGZCVmlXcnYzT3p3azM1VjhIcnBKTk9LeWo2Qk5qRXFJQ3BEZ1pqWjFBSCtBd1hMNWhtbXJWVTVrTmVyWVJIbwp3OEs2S0wzSkhjTDJiUkdZNzlHeHVjUTJvRkgwRWRkVFJsejNtcUx1Z1dic25jY1M3aGlZNm54V1MwYjByYnFtCkppelYzS0ROaHdTS3Z4a0ZQNEs2LzNHcFNRR2NVbk9EWSswSnIwdDJQZDhxOHluM0xUakx5ckdBOVBKVUpSYWsKTzBuVVRzOStyL2Fzbit4M05kQ3g1bS9JNXZ5clJwekU1aTJ1bXdzenVMZ1NIWUxGOE0rbjErdUJvYnJxb0pTNwptdnovLzBNOEhLKzEwajVyc0ZkdGxDSC9sTDgrcDJWV0JhbEZRTENKL1RQQXlNWlQrQnRMckFaODhsWUFHRnZjCmZTUWNLaVhmVlpwNitBSnd6UVR3T29IQXhkNHliMHRxVVZoR2E4NDFIQVZJU2tjaTJZQW5YWUs2NW93UU1tMDkKVHNHMVdBRUk3Sm4zNmp6SHdCamRieWRycE9VK0plV0ZPQ25NcTREWi9UemJObHhnckR5aHB6RFIwdzlvcU9mSApZd0hvK0pLemQrUWNQYStBejlNV29RSURBUUFCbzFZd1ZEQU9CZ05WSFE4QkFmOEVCQU1DQmFBd0V3WURWUjBsCkJBd3dDZ1lJS3dZQkJRVUhBd0l3REFZRFZSMFRBUUgvQkFJd0FEQWZCZ05WSFNNRUdEQVdnQlN1SldiS1lHY2UKcW8wTFVRY2p4SmJnSkZ2bWZUQU5CZ2txaGtpRzl3MEJBUXNGQUFPQ0FnRUFiUFY3MzdFbWptVjRPUDk0VCtRagp4QmRCZGIwQmRTOXJWT0FEOHd1YXlBUk9hcSt1YTdZUVVXQW5sK2NXVEl5SUFCbTZCS1V2VTJ4RE9sR0RscXdjCkZ2bnh1ZFNBVi9YOTl1QTlaNXZ2ZURBSU1EdlVwR2FQTHRsUmVReTJzM1p6dnoxajZhUGxOS3pwK3A0bkVmK1AKdGlIWWRlTXZHZDVibkM2RkRqS2hjamcxS0lNR2pjWCtiVzNWV29FRURCVmlyVTlDQ3pScXpHUjMxVEphUTRVeApKdmJFR1ErZEl3RWZWNDJ2SFFSblVDU2pzYmhqd0JJNUR3a2lDZ09tRVlzUEh3VDl5R1luYTFjM243WFJHZ3EwCnM1dHlBUWVZZER4V3I5OHQ2dUVJVkd6NEJWTk9VaEs3ZTRXWHJ4TVlPYTU3YWxtWlJiVTJSMnMzaTloK1FOb2YKWGk4MGxFYXBTc2Z3SDVKUnJ6WC9NTFl0b0U4RWx0ZWJxTm9zTml6K3VIRzdtQXdsTUkxR2s3WVVqNnVsYW9DcAo5N1JyczJRNGZnMTc1cGwrVHcvSFMxSUNVeVpuYmdPWUIrMjhnNGRTTmhzVEk3ZWJaald0TUVUMXhJbjVNL0ZuCkZqcE5MMThxckpTWHc4b0lZVExTODBEMzFaOGxhTUZiVG1ybnFGNUttekYxQmhvOWp6VmNNMDZhUHNzb2w2dU4KTVhkdWFPL0liTTJXQ2tJc3B6QXVJeGpWWXl3Q1RKMU13VTYzMDNFS3pEQmJTT1NjbC8wOSthSCs4VW9XMm1wQwpyWkRUWVNBOTNxV1Ywc2s3QWttbHdabjBXclA2ZEQ3cjBoRG1PMnB3Yzg0NFM4UHpYZjJzVTRwb1NmNDJiS1QwCldtQUJGbVMzZEZoYkJKdWxOUHpxOXE4PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==")
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "fail creating cert")
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	cf := rest.Config{
		BearerToken: (*t).Token,
		Host:        hubServerUrl,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	k8sClient, err := kubernetes.NewForConfig(&cf)

	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "fail creating client")
	}

	demoCluster.Spec.Foo = "updatedFoo"
	dPod := &v13.Pod{
		ObjectMeta: v12.ObjectMeta{Name: "demo-pod", Namespace: "default"},
		Spec: v13.PodSpec{
			Containers: []v13.Container{
				{Name: "demo-container", Image: "nginx"},
			},
		},
	}
	pod, err := k8sClient.CoreV1().Pods("default").Create(ctx, dPod, v12.CreateOptions{})
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "error updating hub cluster")
	}
	l.Info("Pod created", "pod", pod)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MemberClusterMembershipReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mngiov1.MemberClusterMembership{}).
		Complete(r)
}
