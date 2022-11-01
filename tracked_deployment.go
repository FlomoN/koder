package main

import (
	"fmt"
	"strings"
	"time"

	v1a "k8s.io/api/apps/v1"
)

type TrackedDeployment struct {
	interval           int
	restartUnavailable bool

	deployment v1a.Deployment
	tracking bool

	ticker *time.Ticker;
}

func CreateTrackedDeployment(interval string, unavailable bool, deployment v1a.Deployment) TrackedDeployment {
	a := 0;
	b := "";



	fmt.Fscanf(strings.NewReader(interval), "%d%s", &a, &b);

	switch b {
		case "m":
			a = a * 60;
		case "h":
			a = a * 3600;
		case "d":
			a = a * 3600 * 24;
	}

	return TrackedDeployment{a, unavailable, deployment, false, nil}
}

func (t *TrackedDeployment) Start() {
	t.ticker = time.NewTicker(10 * time.Second);
	
}