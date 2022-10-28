package main

import (
	"context"
	"fmt"

	discovery "github.com/gkarthiks/k8s-discovery"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func main(){
	fmt.Println("Hello World");

	k8s, _ := discovery.NewK8s();

	clientSet, err := kubernetes.NewForConfig(k8s.RestConfig);
	if(err != nil){
		panic(err.Error())
	}

	fmt.Println(clientSet.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{}));
}