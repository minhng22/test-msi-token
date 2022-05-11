package main

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	"k8s.io/client-go/rest"
)

var (
	HubServerApiAddress   = ""
	MemberClusterClientId = ""
	Namespace             = "member-a"
	Pod                   = "demo-msi"
)

const AKSScope = "6dae42f8-4368-4678-94ff-3960e28e3630"

func main() {

	clientID := azidentity.ClientID(MemberClusterClientId)
	opts := &azidentity.ManagedIdentityCredentialOptions{ID: clientID}
	managed, err := azidentity.NewManagedIdentityCredential(opts)
	if err != nil {
		fmt.Printf("\nerror creating the managed identity. err: %v", err.Error())
	}

	token, err := managed.GetToken(context.TODO(), policy.TokenRequestOptions{
		Scopes: []string{AKSScope},
	})
	if err != nil {
		fmt.Printf("\nerror getting the token. err: %v", err.Error())
	}

	cf := rest.Config{
		BearerToken: token.Token,
		Host:        HubServerApiAddress,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(&cf)
	if err != nil {
		fmt.Printf("\nerror creating clientset. err: %v", err.Error())
	}

	for {
		fmt.Printf("\nstart pod listing loop")
		// this should fail as access wasn't granted for namespace 'default'
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("\nerror listing pods. err: %v\n", err.Error())
		}

		// this should succeed
		pods, err = clientset.CoreV1().Pods(Namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("\nerror listing pods. err: %v", err.Error())
		}
		fmt.Printf("\nThere are %d pods in the cluster\n", len(pods.Items))

		_, err = clientset.CoreV1().Pods(Namespace).Get(context.TODO(), Pod, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("\npod %s in namespace %s not found\n", Pod, Namespace)
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("\nerror getting pod %s in namespace %s: %v\n",
				Pod, Namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("\nfound pod %s in namespace %s\n", Pod, Namespace)
		}
		fmt.Printf("\nend pod listing loop")
		time.Sleep(1 * time.Second)
	}
}
