package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	v1a "k8s.io/api/apps/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type TrackedDeployment struct {
	interval           int
	restartUnavailable bool

	deployment *v1a.Deployment
	clientSet  *kubernetes.Clientset
	tracking   bool

	ticker *time.Ticker
	quit   chan bool
}

func CreateTrackedDeployment(interval string, unavailable bool, deployment *v1a.Deployment, clientSet *kubernetes.Clientset) TrackedDeployment {
	a := 0
	b := ""

	fmt.Fscanf(strings.NewReader(interval), "%d%s", &a, &b)

	switch b {
	case "m":
		a = a * 60
	case "h":
		a = a * 3600
	case "d":
		a = a * 3600 * 24
	}

	return TrackedDeployment{a, unavailable, deployment, clientSet, false, nil, make(chan bool)}
}

func (t *TrackedDeployment) Start() {

	t.ticker = time.NewTicker(time.Duration(t.interval) * time.Second)

	go t.loop()

	t.tracking = true
}

func (t *TrackedDeployment) loop() {
	for {
		<-t.ticker.C
		fetchedDeployment, _ := t.clientSet.AppsV1().Deployments(t.deployment.Namespace).Get(context.TODO(), t.deployment.Name, v1.GetOptions{})
		t.deployment = fetchedDeployment
		if t.deployment.Status.UnavailableReplicas > 0 || !t.restartUnavailable {
			t.restart()
		}

	}
}

func (t *TrackedDeployment) Stop() {
	t.ticker.Stop()
	t.quit <- true
	t.tracking = false
}

func (t *TrackedDeployment) restart() {
	log.Println("Restarting ", t.deployment.Name)

	patch := []byte(`{"spec": {"template": {"metadata": {"annotations": {"koder/restartedAt": "` + time.Now().UTC().Format("2006-01-02T15:04:05Z") + `"}}}}}`)
	dep, err := t.clientSet.AppsV1().Deployments(t.deployment.Namespace).Patch(context.TODO(), t.deployment.Name, types.StrategicMergePatchType, patch, v1.PatchOptions{})
	if err != nil {
		log.Fatalln(err.Error())
	} else {
		log.Println("Restart successful for", dep.Name)
	}
}
