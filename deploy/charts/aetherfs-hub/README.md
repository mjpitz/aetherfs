# AetherFS

An [AetherFS] hub deployment consists of a replicated set of nodes that provide the core interfaces as well as a web
interface. An ingress can expose the service to your own private network (or the world).

[aetherfs]: https://aetherfs.tech

## Introduction

This chart bootstraps an aetherfs-hub deployment on a [Kubernetes] cluster using the [Helm] package manager.

[kubernetes]: https://kubernetes.io
[helm]: https://helm.sh

Current chart version is `0.0.0`

## Source Code

- <https://github.com/mjpitz/aetherfs>

## Installing the Chart

To install the chart with the release name `aetherfs-hub`:

```bash
$ helm repo add aetherfs https://aetherfs.tech
$ helm install aetherfs-hub aetherfs/aetherfs-hub
```

The command deploys aetherfs-hub on the Kubernetes cluster using the default configuration.
The [Parameters](#parameters) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm search repo --versions`

## Uninstalling the Chart

To uninstall/delete the `aetherfs-hub` deployment:

```bash
$ helm delete aetherfs-hub
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Parameters

The following table lists the configurable parameters of the aetherfs-hub chart and their default values.

| Key                                        | Type   | Default                                       | Description                                                                                                            |
| ------------------------------------------ | ------ | --------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| affinity                                   | object | `{}`                                          | Provide scheduling preferences for pods when they are being considered for placement.                                  |
| autoscaling.enabled                        | bool   | `false`                                       | Whether to enable autoscaling for this deployment.                                                                     |
| autoscaling.maxReplicas                    | int    | `100`                                         | The maximum number of replicas to deploy.                                                                              |
| autoscaling.minReplicas                    | int    | `1`                                           | The minimum number of replicas to scale down to.                                                                       |
| autoscaling.targetCPUUtilizationPercentage | int    | `80`                                          | What CPU percentage we should scale up at.                                                                             |
| env                                        | list   | `[]`                                          | Specify environment variables used to configure other portions of the application.                                     |
| fullnameOverride                           | string | `""`                                          |                                                                                                                        |
| image.pullPolicy                           | string | `"IfNotPresent"`                              |                                                                                                                        |
| image.repository                           | string | `"ghcr.io/mjpitz/aetherfs"`                   |                                                                                                                        |
| image.tag                                  | string | `""`                                          |                                                                                                                        |
| imagePullSecrets                           | list   | `[]`                                          |                                                                                                                        |
| ingress.annotations                        | object | `{}`                                          | Annotations to add to the ingress class.                                                                               |
| ingress.enabled                            | bool   | `false`                                       | Whether to enable an ingress address for the hub.                                                                      |
| ingress.hosts                              | list   | `[{"host":"chart-example.local","paths":[]}]` | kubernetes.io/tls-acme: "true"                                                                                         |
| ingress.hosts[0].host                      | string | `"chart-example.local"`                       | The name for the service.                                                                                              |
| ingress.hosts[0].paths                     | list   | `[]`                                          | What paths are exposed.                                                                                                |
| ingress.tls                                | list   | `[]`                                          |                                                                                                                        |
| nameOverride                               | string | `""`                                          |                                                                                                                        |
| nfs.enabled                                | bool   | `false`                                       | Deploys the nfs server.                                                                                                |
| nodeSelector                               | object | `{}`                                          | Limit node selection to nodes with the matching set of labels.                                                         |
| podAnnotations                             | object | `{}`                                          | Annotations to provide directly to the pod.                                                                            |
| podSecurityContext                         | object | `{}`                                          | The security context to provide to the pod.                                                                            |
| replicaCount                               | int    | `1`                                           | The number of instances to run.                                                                                        |
| resources                                  | object | `{}`                                          |                                                                                                                        |
| securityContext                            | object | `{}`                                          | The security context to apply specifically the aetherfs-hub container.                                                 |
| service.clusterIP                          | string | `""`                                          | Request a specific service IP address for use in the cluster, or None for a headless mode.                             |
| service.port                               | int    | `80`                                          | The port to advertise the service on.                                                                                  |
| service.type                               | string | `"ClusterIP"`                                 | The type of service to create.                                                                                         |
| serviceAccount.annotations                 | object | `{}`                                          | Annotations to add to the service account.                                                                             |
| serviceAccount.create                      | bool   | `true`                                        | Specifies whether a service account should be created.                                                                 |
| serviceAccount.name                        | string | `""`                                          | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| tolerations                                | list   | `[]`                                          | Configure rules that allow pods to run on tainted nodes. For example, running on control plane nodes.                  |
