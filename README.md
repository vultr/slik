# SLiK (SLurm in Kubernetes)

## Overview
An operator to deploy slurm in kubernetes.

Note: This project has been renamed from slinkee to slik; this was due to a name conflict.

## Usage
Everything is public, including the slurm images. You do not need any auth or secret sauce to use this. If you intend to use in a different cloud platform you may need to make tweaks to the mariadb statefulset. If you intend to deploy slurm on arm you'll need to build arm images.

You can deploy slik into your kubernetes cluster simply with: `helm install -f helm/slik/values.yaml slik ./helm/slik/`

You can then deploy a slurm cluster with one of the samples: `kubectl apply -f payloads/simple.yaml`

If you deploy "full" slurm cluster (with database) it can take awhile to initialize MariaDB.

You then interact with the slurm cluster through the toolbox pod: `kubectl exec --it toolbox -- bash`

Sample yaml:

```yaml
apiVersion: "ahmedtremo.com/v1"
kind: Slik
metadata:
  name: full
spec:
  namespace: default
  slurmdbd: true
  slurmrestd: true
  mariadb:
    storage_size: 5G
    storage_class: default
```

Update operations are not currently supported, you should rebuild the cluster instead. Delete the slurm deployment, then re-create it. If you use a PVC that is retained you should not lose any data.

You can list the slurm clusters: `kubectl get sliks`

You can delete slurm clusters: `kubectl delete slik <name>`

If you need to troubleshoot, check the logs for the operator: `kubectl logs slik-operator...`

## Contribution(s)
Please send any PRs for contributions/suggestions.

## Architecture
Below are some details on the architecture:
- `slurmabler`: Used to label the nodes in kubernetes so that it's easier to generate the `slurm.conf`. This provides a _guarantee_ that the generated configuration will work as it extracts `slurmd -C` and attaches the fields as labels. Deployed as DaemonSet.
- `munged`: Key is generated with HKDF in Go, then injected into all slurm services as a sidecar. Required for auth and doing anything in the cluster.
- `slurmctld`: Primary service that is interacted with.
- `slurmd`: Gets deployed as a Deployment per node. DaemonSet was not sufficient. A new type would be necessary that is between Deployment/DaemonSet. This is something that can be done with future work.
- `slurmdbd`: Job accounting history, uses MariaDB as the backend.
- `slurmrestd`: Deployed but has not been tested.

All the images are Ubuntu images using the Canonical built slurm.

You'll see various Services as well as various ConfigMaps. The ConfigMaps use Go's templating system to generate some config. The Services tie back all of the slurm services so that they can work properly.