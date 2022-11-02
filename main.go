package main

import (
	"context"
	"strings"

	discovery "github.com/gkarthiks/k8s-discovery"
	v1a "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func main(){
	k8s, _ := discovery.NewK8s();

	clientSet, err := kubernetes.NewForConfig(k8s.RestConfig);
	if(err != nil){
		panic(err.Error())
	}

	allDeployments := getAnnotatedDeployments(clientSet);

	for _, dep := range allDeployments {
		interval, unavail := extractAnnotations(dep)
		t := CreateTrackedDeployment(interval, unavail, dep);

		t.Start();
	}

	select {} // Blocks forever

}


func getAnnotatedDeployments(clientSet *kubernetes.Clientset) []v1a.Deployment {

	allDeployments := []v1a.Deployment{};

	namespaces, _ := clientSet.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{});
	for _, namespace := range namespaces.Items {
		deployments, _ := clientSet.AppsV1().Deployments(namespace.Name).List(context.TODO(), v1.ListOptions{});
		for _, deployment := range deployments.Items {

			keep := false;
			for k, _ := range deployment.Annotations {
				if(strings.Contains(k, "koder/")){
					keep = true;
					break
				}
			}
			if keep { 
				allDeployments = append(allDeployments, deployment);
			}
			
		}
	}

	return allDeployments;
}


func extractAnnotations(dep v1a.Deployment) (interval string, unavailable bool) {

	for k, v := range dep.Annotations {
		switch k {
			case "koder/restart-time":
				interval = v;
			case "koder/restart-unavailable":
				unavailable = v == "true";
			default:
				continue
		}
	}

	return
}