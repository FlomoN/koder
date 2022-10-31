# koder

Kubernetes Operator for Deployment Restarts

## Annotations

- `koder/restart-time`: A value indicating the time (`30s`, `20m`, `4h`, `3d`)
- `koder/restart-unavailable`: Restart a container that wont start properly (stuck in unavailable, assuming a restart solves the problem) (`true`)
