package main

import (
	"context"
	"log"
	"strings"
	"time"

	discovery "github.com/gkarthiks/k8s-discovery"
	"golang.org/x/exp/slices"
	v1a "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func main() {

	var trackers []*TrackedDeployment

	k8s, _ := discovery.NewK8s()

	clientSet, err := kubernetes.NewForConfig(k8s.RestConfig)
	if err != nil {
		panic(err.Error())
	}

	ticker := time.NewTicker(30 * time.Second)

	go func() {
		for {
			allDeployments := getAnnotatedDeployments(clientSet)

			var keep []int

			for _, dep := range allDeployments {
				found := slices.IndexFunc(trackers, func(t *TrackedDeployment) bool {
					return t.deployment.Name == dep.Name && t.deployment.Namespace == dep.Namespace
				})
				if found == -1 {
					interval, unavail := extractAnnotations(dep)
					t := CreateTrackedDeployment(interval, unavail, &dep, clientSet)
					t.Start()
					log.Println("Started tracking", dep.Name)
					trackers = append(trackers, &t)
				} else {
					keep = append(keep, found)
				}

			}

			var keepTD []*TrackedDeployment

			// Filter out which to keep
			for _, keepIndex := range keep {
				keepTD = append(keepTD, trackers[keepIndex])
				trackers[keepIndex] = nil
			}

			// Now remove the rest
			for _, td := range trackers {
				if td != nil {
					log.Println("Stopped tracking", td.deployment.Name)
					td.Stop()
				}
			}

			trackers = keepTD

			<-ticker.C
		}
	}()

	select {} // Blocks forever

}

func getAnnotatedDeployments(clientSet *kubernetes.Clientset) []v1a.Deployment {

	allDeployments := []v1a.Deployment{}

	namespaces, _ := clientSet.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{})
	for _, namespace := range namespaces.Items {
		deployments, _ := clientSet.AppsV1().Deployments(namespace.Name).List(context.TODO(), v1.ListOptions{})
		for _, deployment := range deployments.Items {

			keep := false
			for k, _ := range deployment.Annotations {
				if strings.Contains(k, "koder/") {
					keep = true
					break
				}
			}
			if keep {
				allDeployments = append(allDeployments, deployment)
			}

		}
	}

	return allDeployments
}

func extractAnnotations(dep v1a.Deployment) (interval string, unavailable bool) {

	unavailable = false

	for k, v := range dep.Annotations {
		switch k {
		case "koder/restart-time":
			interval = v
		case "koder/restart-unavailable":
			unavailable = v == "true"
		default:
			continue
		}
	}

	return
}
