# Deploying SLiK

This guide covers deploying the SLiK operator and creating a Slurm cluster from the included sample manifests.

## Prerequisites

- Kubernetes v1.28 or newer.
- The `SidecarContainers` feature gate enabled. It is enabled by default in Kubernetes v1.29+.
- `kubectl` configured for the target cluster.
- Helm 3.
- Worker nodes that can schedule the `slurmabler` DaemonSet. Nodes with `NoSchedule` or `NoExecute` taints are ignored by SLiK.
- A storage class if deploying the full Slurm cluster with MariaDB accounting.

## Install The Operator

From the repository root:

```sh
helm install -f helm/slik/values.yaml slik ./helm/slik/
```

Check that the operator is running:

```sh
kubectl get pods
kubectl logs deploy/slik-operator
```

The Helm chart installs the CRD, service account, config map, and operator deployment.

## Deploy A Simple Slurm Cluster

The simple payload deploys Slurm without `slurmdbd`, `slurmrestd`, or MariaDB:

```sh
kubectl apply -f payloads/simple.yaml
```

Check the custom resource and generated pods:

```sh
kubectl get sliks
kubectl get pods -n default
```

The cluster is ready when the `Slik` resource reaches `ACTIVE` and the Slurm pods are running.

## Deploy A Full Slurm Cluster

The full payload enables MariaDB-backed accounting and `slurmrestd`:

```sh
kubectl apply -f payloads/full.yaml
```

Before applying it outside Vultr, update `payloads/full.yaml` to use a storage class that exists in your cluster:

```yaml
mariadb:
  storage_size: 50G
  storage_class: your-storage-class
```

MariaDB can take a few minutes to initialize on first boot.

## Access Slurm

Find the toolbox pod:

```sh
kubectl get pods -n default -l app=slurm-toolbox
```

Open a shell in it:

```sh
kubectl exec -n default -it deploy/test-slurm-toolbox -- bash
```

For the full sample, the deployment name is based on the `Slik` resource name:

```sh
kubectl exec -n default -it deploy/full-slurm-toolbox -- bash
```

From inside the toolbox, run Slurm commands such as:

```sh
sinfo
scontrol show nodes
```

## Upgrade Or Recreate A Cluster

SLiK does not currently support in-place updates to a Slurm cluster spec. Delete and recreate the `Slik` resource instead:

```sh
kubectl delete slik test
kubectl apply -f payloads/simple.yaml
```

If your MariaDB PVC uses a retained reclaim policy, accounting data can survive cluster recreation.

## Delete SLiK

Delete Slurm clusters first:

```sh
kubectl delete slik test
kubectl delete slik full
```

Then uninstall the operator:

```sh
helm uninstall slik
```

If you deployed a full cluster, check for retained PVCs before deleting storage manually:

```sh
kubectl get pvc -n default
```

## Troubleshooting

Check operator logs:

```sh
kubectl logs deploy/slik-operator
```

Check the `Slik` status:

```sh
kubectl get sliks
kubectl describe slik test
```

Check whether nodes were labeled by `slurmabler`:

```sh
kubectl get nodes --show-labels | grep slik.vultr.com
```

If a cluster stays pending, verify at least one schedulable worker node can run `slurmabler`. Control-plane nodes and nodes with `NoSchedule` or `NoExecute` taints are skipped.

If the full cluster fails validation, confirm MariaDB storage is at least `45G` and the storage class exists.

If `kubectl exec` cannot find the toolbox deployment, confirm the deployment name uses the `Slik` resource name as its prefix, for example `test-slurm-toolbox` or `full-slurm-toolbox`.
