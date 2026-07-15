# SLiK (SLurm in Kubernetes)

## Overview
An operator to deploy slurm in kubernetes.

Note: This project has been archived in favor of the [Slinky](https://github.com/SlinkyProject) project.

## Requirements
- Kubernetes v1.36+.
- Go 1.26+ if building from source.

## Usage
Everything is public, including the Slurm images. You do not need any auth or secret sauce to use this. If you intend to use a different cloud platform you may need to tweak the MariaDB storage class.

Add the Helm repository and deploy SLiK into your Kubernetes cluster:

```sh
helm repo add slik https://vultr.github.io/slik
helm repo update
helm install slik slik/slik
```

From a local checkout, you can also install the chart with: `helm install -f helm/slik/values.yaml slik ./helm/slik/`

You can then deploy a slurm cluster with one of the samples: `kubectl apply -f payloads/simple.yaml`

For a complete walkthrough, see [Deploying SLiK](docs/deployment.md).

If you deploy "full" slurm cluster (with database) it can take awhile to initialize MariaDB.

You then interact with the Slurm cluster through the toolbox pod: `kubectl exec -n default -it deploy/test-slurm-toolbox -- bash`

Sample yaml:

```yaml
apiVersion: "hpc.vultr.com/v1"
kind: Slik
metadata:
  name: full
spec:
  namespace: default
  slurmdbd: true
  slurmrestd: true
  mariadb:
    storage_size: 50G
    storage_class: vultr-block-storage-hdd-retain
```

You can update a Slurm cluster by editing and re-applying the `Slik` resource. The operator reconciles owned Deployments, DaemonSets, Services, ConfigMaps, optional `slurmdbd`/`slurmrestd`/MariaDB components, and MariaDB PVC expansion when the storage class allows it. The generated `munge.key` is preserved across updates.

You can list the slurm clusters: `kubectl get sliks`

You can delete slurm clusters: `kubectl delete slik <name>`

If you need to troubleshoot, check the logs for the operator: `kubectl logs slik-operator...`

## Contribution(s)
Please send any PRs for contributions/suggestions.

## Helm Chart Releases

Helm chart releases are published to GitHub Pages at `https://vultr.github.io/slik` by the `Release Helm Chart` workflow.

To release a chart, merge or push a commit to `master` with a message like:

```text
Release helm-0.0.1 #patch
```

The workflow packages `helm/slik`, creates a GitHub release named `helm-0.0.1`, and updates the `gh-pages` branch with the Helm repository index. Use `#major`, `#minor`, or `#patch` to document the intended release type.

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
