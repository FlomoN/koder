# KODeR

**K**ubernetes **O**perator for **De**ployment **R**estarts

## Problem

Sometimes your Deployments get stuck in an unavailable state, where only a restart helps (and a readiness or liveness probe doesn't).
Or maybe you just want to restart an application in a regular interval.

Then KODeR may come to the rescue.

It simply tracks all deployments with `koder` annotations.

## Installation

Use and modify the manifests in `./deploy` to your needs.
This will deploy KODeR and the necessary service account in a namespace `koder`.

## Usage

There are two possible annotations that can be supplied to deployments:

- `koder/restart-time`: A value indicating the interval to restart the application OR the checking interval to restart on unavailability (`30s`, `20m`, `4h`, `3d`)
- `koder/restart-unavailable`: Restart a container that wont start properly (stuck in unavailable, assuming a restart solves the problem) (`true`)
