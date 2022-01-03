# AetherFS

The [AetherFS] deployment consists of a replicated set of nodes that provide the core interfaces as well as a web
interface. An ingress exposes the service to your own private network (or the world).

[aetherfs]: https://aetherfs.tech

## Introduction

This chart bootstraps an aetherfs deployment on a [Kubernetes] cluster using the [Helm] package manager.

[kubernetes]: https://kubernetes.io
[helm]: https://helm.sh

Current chart version is `0.0.0`

## Source Code

- <https://github.com/mjpitz/aetherfs>

## Installing the Chart

To install the chart with the release name `aetherfs`:

```bash
$ helm repo add aetherfs https://aetherfs.tech
$ helm install aetherfs aetherfs/aetherfs
```

The command deploys aetherfs on the Kubernetes cluster using the default configuration.
The [Parameters](#parameters) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm search repo --versions`

## Uninstalling the Chart

To uninstall/delete the `aetherfs` deployment:

```bash
$ helm delete aetherfs
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Requirements

| Repository               | Name  | Version |
| ------------------------ | ----- | ------- |
| https://charts.dexidp.io | dex   | 0.5.0   |
| https://helm.min.io/     | minio | 8.0.9   |

## Parameters

The following table lists the configurable parameters of the aetherfs chart and their default values.

| Key                                        | Type   | Default                     | Description                                                                                                            |
| ------------------------------------------ | ------ | --------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| affinity                                   | object | `{}`                        | Provide scheduling preferences for pods when they are being considered for placement.                                  |
| autoscaling.enabled                        | bool   | `false`                     | Whether to enable autoscaling for this deployment.                                                                     |
| autoscaling.maxReplicas                    | int    | `100`                       | The maximum number of replicas to deploy.                                                                              |
| autoscaling.minReplicas                    | int    | `1`                         | The minimum number of replicas to scale down to.                                                                       |
| autoscaling.targetCPUUtilizationPercentage | int    | `80`                        | What CPU percentage we should scale up at.                                                                             |
| dex.enabled                                | bool   | `false`                     | Weather dex is enabled for authentication.                                                                             |
| env                                        | list   | `[]`                        | Additional environment variables.                                                                                      |
| envFrom                                    | list   | `[]`                        | Additional environment variables mounted from secrets.                                                                 |
| fullnameOverride                           | string | `""`                        |                                                                                                                        |
| image.pullPolicy                           | string | `"IfNotPresent"`            |                                                                                                                        |
| image.repository                           | string | `"ghcr.io/mjpitz/aetherfs"` |                                                                                                                        |
| image.tag                                  | string | `""`                        |                                                                                                                        |
| imagePullSecrets                           | list   | `[]`                        |                                                                                                                        |
| kind                                       | string | `"Deployment"`              | How AetherFS should be deployed. Either as a Deployment or DaemonSet.                                                  |
| minio.enabled                              | bool   | `false`                     | Weather minio is enabled for storage.                                                                                  |
| nameOverride                               | string | `""`                        |                                                                                                                        |
| nodeSelector                               | object | `{}`                        | Limit node selection to nodes with the matching set of labels.                                                         |
| podAnnotations                             | object | `{}`                        | Annotations to provide directly to the pod.                                                                            |
| podSecurityContext                         | object | `{}`                        | The security context to provide to the pod.                                                                            |
| priorityClassName                          | string | `""`                        | Specify a priority class name to set pod priority.                                                                     |
| replicaCount                               | int    | `1`                         | The number of instances to run.                                                                                        |
| resources                                  | object | `{}`                        |                                                                                                                        |
| securityContext                            | object | `{}`                        | The security context to apply specifically the aetherfs-hub container.                                                 |
| service.annotations                        | object | `{}`                        | Annotations to be added to the service.                                                                                |
| service.clusterIP                          | string | `""`                        | Request a specific service IP address for use in the cluster, or None for a headless mode.                             |
| service.http.nodePort                      | int    | `nil`                       | http node port (when applicable)                                                                                       |
| service.http.port                          | int    | `8080`                      | HTTP service port                                                                                                      |
| service.nfs.enabled                        | bool   | `false`                     | Enable the AetherFS NFS server.                                                                                        |
| service.nfs.nodePort                       | int    | `nil`                       | NFS node port (when applicable)                                                                                        |
| service.nfs.port                           | int    | `2049`                      | NFS service port                                                                                                       |
| service.type                               | string | `"ClusterIP"`               | The type of service to create.                                                                                         |
| service.ui.enabled                         | bool   | `false`                     | Enable the web UI.                                                                                                     |
| serviceAccount.annotations                 | object | `{}`                        | Annotations to add to the service account.                                                                             |
| serviceAccount.create                      | bool   | `true`                      | Specifies whether a service account should be created.                                                                 |
| serviceAccount.name                        | string | `""`                        | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| tolerations                                | list   | `[]`                        | Configure rules that allow pods to run on tainted nodes. For example, running on control plane nodes.                  |
| volumeMounts                               | list   | `[]`                        | Additional volume mounts.                                                                                              |
| volumes                                    | list   | `[]`                        | Additional storage volumes.                                                                                            |
