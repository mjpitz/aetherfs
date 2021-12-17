# AetherFS Datasets

This chart makes it easy to mount an [AetherFS hub] deployment using an NFS volume mount. Since data provided over the
NFS interface is read only, any number of pods can mount this volume.

[aetherfs hub]: https://github.com/mjpitz/aetherfs/tree/main/deploy/helm/aetherfs-hub

## Introduction

This chart bootstraps an aetherfs-datasets deployment on a [Kubernetes] cluster using the [Helm] package
manager.

[kubernetes]: https://kubernetes.io
[helm]: https://helm.sh

Current chart version is `0.0.0`

## Source Code

- <https://github.com/mjpitz/aetherfs>

## Installing the Chart

While this chart can be installed directly with helm, it's intended to be added as a dependency:

```yaml
# In requirements.yaml for v1 charts
# In Chart.yaml for v2 charts
dependencies:
  - repository: http://aetherfs.tech
    name: aetherfs-datasets
    version: 0.0.0
    condition: aetherfs-datasets.enabled
```

This will automatically create a `ReadOnlyMany` [PersistentVolume] and [PersistentVolumeClaim] that your Deployment,
DaemonSet, or StatefulSet can mount.

[persistentvolume]: #tbd
[persistentvolumeclaim]: #tbd

## Uninstalling the Chart

To uninstall, simply remove the dependency and Helm should automatically clean itself up.

## Parameters

The following table lists the configurable parameters of the aetherfs-datasets chart and their default
values.

| Key        | Type   | Default | Description                            |
| ---------- | ------ | ------- | -------------------------------------- |
| nfs.port   | int    | `2049`  | The port the nfs server is bound to.   |
| nfs.server | string | `""`    | The ip or dns name for the nfs server. |
