package main

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	discovery "github.com/gkarthiks/k8s-discovery"
	"github.com/samber/lo"
	v1a "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func main() {

	log.Println("Welcome to Koder :)")

	var trackers []*TrackedDeployment

	k8s, _ := discovery.NewK8s()

	clientSet, err := kubernetes.NewForConfig(k8s.RestConfig)
	if err != nil {
		panic(err.Error())
	}

	ticker := time.NewTicker(30 * time.Second)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			allDeployments := getAnnotatedDeployments(clientSet)

			var keep []int
			var newtrackers []*TrackedDeployment

			for _, dep := range allDeployments {
				_, found, _ := lo.FindIndexOf(trackers, func(t *TrackedDeployment) bool {
					return t.deployment.Name == dep.Name && t.deployment.Namespace == dep.Namespace
				})
				if found == -1 {
					interval, unavail := extractAnnotations(dep)
					depcopy := dep
					t := CreateTrackedDeployment(interval, unavail, &depcopy, clientSet)
					t.Start()
					log.Println("Started tracking", dep.Name, t.deployment.Name)
					newtrackers = append(newtrackers, &t)
				} else {
					keep = append(keep, found)
				}

			}

			var keepTD []*TrackedDeployment

			// Mark all for removal first
			for _, td := range trackers {
				td.MarkedForRemoval = true
			}

			// Filter out which to keep
			for _, keepIndex := range keep {
				keepTD = append(keepTD, trackers[keepIndex])
				trackers[keepIndex].MarkedForRemoval = false
			}

			// Now remove the rest
			for _, td := range trackers {
				if td.MarkedForRemoval {
					log.Println("Stopped tracking", td.deployment.Name)
					td.Stop()
				}
			}

			trackers = []*TrackedDeployment{}
			trackers = append(trackers, keepTD...)
			trackers = append(trackers, newtrackers...)

			<-ticker.C
		}
	}()

	wg.Wait()
}

func getAnnotatedDeployments(clientSet *kubernetes.Clientset) []v1a.Deployment {

	allDeployments := []v1a.Deployment{}

	namespaces, err := clientSet.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		log.Println("Trouble retrieving namespaces")
		panic(err.Error())
	}
	for _, namespace := range namespaces.Items {
		deployments, err := clientSet.AppsV1().Deployments(namespace.Name).List(context.TODO(), v1.ListOptions{})
		if err != nil {
			log.Println("Trouble retrieving deployments")
			panic(err.Error())
		}
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
