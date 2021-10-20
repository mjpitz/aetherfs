# AetherFS Architecture

* [Background](#background)
  * [Motivation](#motivation)
  * [Concepts](#concepts)
* [Overview](#overview)
  * [Requirements](#requirements)
* [Implementation](#implementation)
  * [Components](#components)
    * [AetherFS CLI](#aetherfs-cli)
    * [AetherFS Hub](#aetherfs-hub)
    * [AetherFS Agent](#aetherfs-agent)
  * [Interfaces](#interfaces)
    * [gRPC](#grpc)
    * [REST](#rest)
    * [HTTP File Server](#http-file-server)
    * [FUSE File System](#fuse-file-system)
    * [Web](#web)
  * [Data Management](#data-management)
    * [Packing](#packing)
    * [Persistence](#persistence)
    * [Caching](#caching)
  * [Security & Privacy](#security--privacy)
    * [Authentication](#authentication)
    * [Authorization](#authorization)
    * [Encryption at Rest](#encryption-at-rest)
    * [Encryption in Transit](#encryption-in-transit)
* [Milestones](#milestones)
  * [Deliverables](#deliverables)

## Background

While working at [Indeed][], many of our systems leveraged a producer-consumer architecture. In this pattern, services
can load an artifact containing data into memory in order to better service requests. These artifacts could be consumed
by a single service or shared across many services. Eventually, this developed into a platform called [RAD][] (short for
resilient artifact distribution).

Sometime after Indeed developed RAD internally, we saw a similar system open sourced from [Netflix][] called [Hollow][].
Hollow is a Java library used to distribute in-memory datasets. Unlike RAD's file-system based approach, Hollow stored 
everything in S3. While I have not used Hollow myself, I see the utility it provides to the Java ecosystem.

[Indeed]: https://www.indeed.com
[RAD]: https://www.youtube.com/watch?v=lDXdf5q8Yw8
[Netflix]: https://netflix.com
[Hollow]: https://github.com/Netflix/hollow

### Motivation

As I've run more workloads on my Raspberry Pis, I've found myself wanting a different type of file system. Since my pis
don't have a lot of disk capacity, I want to be able to distribute parts of my total data across the cluster. In 
addition to that, I want to deduplicate blocks of data that might repeat between versions of that to help keep the
footprint low.

### Concepts

**Dataset**

At Indeed, we referred to these as "artifacts," but I often found the term to be too generic in conversation. In
AetherFS, we refer to collections of information as a _dataset_. Datasets can be tagged, similar to containers. This
allows publishers to manage their own history and channels for consumers.

_**NOTE:** In theory, this system could be used as a generalized package manager, in which the term "artifact" would be
appropriate. However, that is not my intent for this project._

**Tag**

A Tag in AetherFS is a pointer to a set of blocks that make up the associated dataset. They can be used to mark concrete
versions of a dataset (like a [semantic][] or [calendar][] version) or a floating pointer (like `stable` or `next`).
Using a floating pointer allows consumers to stay up to date on the latest version of the dataset without requiring
redeploy.

[semantic]: https://semver.org/
[calendar]: https://calver.org/

**Blocks**

When a publisher pushes a dataset into AetherFS, its contents are broken up into fixed sized _blocks_. This allows
smaller files to be stored as a single block and larger ones to be broken up into multiple smaller ones. In the end,
their goal is to reduce the amount of data between versions and reduce the number of calls made to the backend.

Blocks are immutable, which allows them to be cached amongst agents in your clusters. This allows hot data to be read
from your peers instead of making a call to your underlying storage tier.

**Signature**

Each block stored in S3 is given a unique, cryptographic signature that represents the contents of the block (i.e. a 
cryptographic hash). Signatures allow clients to check if a block already exists, to download a block, and to upload a 
block.

## Overview

This document focuses on the design of a highly available, partition-tolerant virtual file system for small to medium
datasets.

### Requirements

  - Efficiently store and query information in [AWS S3][] (or equivalent).
  - Information should be encrypted in transit and at rest.
  - Authenticate clients (users and services) using common schemes (OIDC, Basic).
  - Enforce access controls around datasets.
  - Provide numerous interfaces to manage and access information in the system.
  - Built in developer tools to help producers understand the performance of their datasets.

[AWS S3]: https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html

## Implementation

<a href="docs/assets/overview.png">
<img src="https://aetherfs.tech/assets/overview.png" align="right" width="40%"/>
</a>

AaetherFS deploys as a simple client-server architecture, with a few caveats.

The AetherFS Agent process can cluster into a mesh of nodes, allowing it to share data amongst its peers. This allows
agents to reduce calls to the database by re-using blocks of data that are likely already within the cluster.

### Components

While there are a few moving components to AetherFS, we distribute everything as part of a single binary. 

#### AetherFS CLI

![role: client](https://img.shields.io/badge/role-client-white?style=for-the-badge)
![interfaces: command line, fuse](https://img.shields.io/badge/interfaces-command%20line,%20fuse-white?style=for-the-badge)

The command line interface (CLI) is the primary mechanism for interacting with datasets stored in AetherFS. Operators
use it to run the various infrastructure components and engineers use it to download datasets on demand. It can even be 
run as an init container to initialize a file system for a Kubernetes Pod.

```
$ aetherfs -h

NAME:
   aetherfs - A virtual file system for small to medium sized datasets (MB or GB, not TB or PB).

USAGE:
   aetherfs [options] <command>

COMMANDS:
   pull     Pulls a dataset from AetherFS
   push     Pushes a dataset into AetherFS
   run      Run the various AetherFS processes
   version  Print the binary version information

GLOBAL OPTIONS:
   --log_level value   adjust the verbosity of the logs (default: "info") [$LOG_LEVEL]
   --log_format value  configure the format of the logs (default: "json") [$LOG_FORMAT]
   --help, -h          show help (default: false)

COPYRIGHT:
   Copyright 2021 The AetherFS Authors - All Rights Reserved

```

#### AetherFS Hub

![role: server](https://img.shields.io/badge/role-server-white?style=for-the-badge)
![interfaces: grpc, file server, rest, web](https://img.shields.io/badge/interfaces-grpc,%20file%20server,%20rest,%20web-white?style=for-the-badge)

AetherFS hubs host datasets for download. They're horizontally scalable making it easy to scale up when additional CPU
or memory is needed. Hubs are responsible for managing the underlying storage tier (S3) and verifying authenticated
clients have access to datasets. AetherFS itself does not implement an identity provider, but will eventually work with
[dex](https://dexidp.io) or other OIDC providers.

```
$ aetherfs run hub -h

NAME:
   aetherfs run hub - Runs the AetherFS Hub process

USAGE:
   aetherfs run hub [options]

DESCRIPTION:
   The aetherfs-hub process is responsible for collecting and hosting datasets.

OPTIONS:
   --port value                               which port the HTTP server should be bound to (default: 8080) [$PORT]
   --tls_cert_path value                      where to locate certificates for communication [$TLS_CERT_PATH]
   --storage_driver value                     configure how information is stored (default: "s3") [$STORAGE_DRIVER]
   --storage_s3_endpoint value                location of s3 endpoint (default: "s3.amazonaws.com") [$STORAGE_S3_ENDPOINT]
   --storage_s3_tls_cert_path value           where to locate certificates for communication [$STORAGE_S3_TLS_CERT_PATH]
   --storage_s3_access_key_id value           the access key id used to identify the client [$STORAGE_S3_ACCESS_KEY_ID]
   --storage_s3_secret_access_key value       the secret access key used to authenticate the client [$STORAGE_S3_SECRET_ACCESS_KEY]
   --storage_s3_region value                  the region where the bucket exists [$STORAGE_S3_REGION]
   --storage_s3_bucket value                  the name of the bucket to use [$STORAGE_S3_BUCKET]
   --help, -h                                 show help (default: false)

```

#### AetherFS Agent

![role: client, server](https://img.shields.io/badge/role-client,%20server-white?style=for-the-badge)
![interfaces: grpc, file server, fuse, rest](https://img.shields.io/badge/interfaces-grpc,%20file%20server,%20fuse,%20rest-white?style=for-the-badge)

AetherFS agents support a variety of use cases. At the end of the day, it's goal is to serve information from AetherFS
as quickly as possible through a variety of mechanisms. In addition to the File Server and FUSE interfaces, it can
manage a local file system. This allows applications to interact with files on disk just as if they were packaged
locally.

_Not yet implemented._

```
$ aetherfs run agent -h
```

### Interfaces

This project exposes numerous interfaces for interaction. Why? For the most part, everyone has their own preference. In
this case, a lot of it is offered out of convenience. For example, instead of implementing a `FileAPI`, I implemented a
[file server](#http-file-server) instead. Beyond the initial HTTP File System, this core set of logic can be reused to
write the FUSE interface later on.

#### gRPC

gRPC is the primary form of communication between processes. While it has some sharp edges, supporting communication
through an HTTP ingress is a requirement for end users to be able to download data. gRPC provides a convenient way to
write stream based APIs which are critical to large datasets. AetherFS uses gRPC to communicate between components.  

**DatasetAPI**

The `DatasetAPI` allows callers to interact with various datasets stored within AetherFS. Dataset manifests contain a
complete list of files within the dataset, their sizes, and last modified timestamps. The manifests also contain a list 
of blocks that are required to construct the dataset. Using these components, clients can piece together the underlying
files.

**BlockAPI**

The `BlockAPI` gives callers direct access to the block data. You must use the `DatasetAPI` to obtain block references.
The `BlockAPI` does not provide callers with the ability to list blocks (intentional design decision). An entire block
can be read at a time, or just part of one.

#### REST

AetherFS provides a REST interface via grpc-gateway. It's offered primarily for end users to interact with if no gRPC
client has been generated or gRPC is unavailable. For example, the web interface is built against the REST interface 
since gRPC isn't* available in browsers. Because our API contains streaming calls, support over REST is difficult to do.

All REST endpoints are under the `/api` prefix.

#### HTTP File Server

Using [Golang's http.FileServer][], AetherFS was able to provide a quick prototype of a FileSystem implementation. Using
[HTTP range requests][], callers are able to read segments of large files that may be too large to fit in memory all at
once. Agents can still make use of caching to keep the data local to the process.

This currently resides under the `/fs` prefix.

[Golang's http.FileServer]: https://pkg.go.dev/net/http#FileServer
[HTTP range requests]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Range_requests

#### FUSE File System

_Not yet implemented._
 
#### Web

A web interface is available under the `/ui` prefix. It allows you to explore datasets, their tags, and even files
within the dataset from a graphical interface. If you include a `README.md` file in the root of your dataset, we're
able to render inline in the browser. Eventually, it will even be able to provide insight into how well your dataset 
performs between versions.

In addition to providing a way to explore datasets, it serves a Swagger UI for engineers to explore the API in more 
detail.

### Data Management

<a href="docs/assets/seen-stored-cached.png">
<img src="https://aetherfs.tech/assets/seen-stored-cached.png" align="right" width="40%"/>
</a>

There isn't a ton of bells and whistles to how data is managed within the AetherFS architecture. We expect the storage
provider to offer durability guarantees. Caching will play a more important role in the next release and will require
some detail.

#### Packing 

When uploading files to AetherFS, we pack all files found in a target directory, zip, or tarball into a single 
contiguous blob. This large blob is broken into smaller blocks that are ideally sized for your storage layer. For 
example, Amazon Athena documentation suggests using S3 objects between 256MiB and 1GiB to optimize network bandwidth.
<!-- needs citation -->

Each dataset can choose their own block size, ideally striving to get the most reuse between versions. While producers 
have control over the size of the blocks that are stored in AetherFS, they do not control the size of the cacheable 
parts. This allows consumers of datasets to tune their usage based by adding more memory, disk, or peers where they need
to. To help explain this a little more, let us consider the following example manifest.

```json
{
    "dataset": {
        "files": [{
            "name": "HYP_HR_SR_W_DR.VERSION.txt",
            "size": "5",
            "lastModified": "2012-11-08T06:34:01Z"
        }, {
            "name": "HYP_HR_SR_W_DR.prj",
            "size": "147",
            "lastModified": "2012-09-19T13:56:32Z"
        }, {
            "name": "HYP_HR_SR_W_DR.tfw",
            "size": "173",
            "lastModified": "2009-12-22T20:48:30Z"
        }, {
            "name": "HYP_HR_SR_W_DR.tif",
            "size": "699969826",
            "lastModified": "2012-07-16T16:23:30Z"
        }, {
            "name": "README.html",
            "size": "30763",
            "lastModified": "2012-11-08T06:34:01Z"
        }],
        "blockSize": 268435456,
        "blocks": [
            "43fbwtgpsbh6naab7cevphlcxvp6xi3etkwataxtvxgkltcpky4q====",
            "feufuim3ogmo34aqv5htwkww4ovll5weurt3nc6p233irbjlwjwq====",
            "7egqp74hvvfdau5vy6tyynvzs6mlxoikit5bc4fdxeiuwwclcd4q===="
        ]
    }
}
```

This manifest is based on [BitTorrent][] and the modifications Indeed need to make to support RAD. While not immediately
obvious, it contains all the information needed to reconstruct the original file system AetherFS took a snapshot of.
For example, the snippet of pseudocode below can be used to identify which blocks need to be read in order to
reconstruct a specific file. 

[BitTorrent]: https://en.wikipedia.org/wiki/BitTorrent

```
offset = 0
size = 0
for file in files {
    if file.name == desiredFile {
        size = file.size
        break;
    }
    offset = offset + sile.size
}

block = blocks[offset / blockSize]
blockOffset = offset % blockSize

// read block starting at blockOffset for size
// note size > blockSize ;-)
```

Keep in mind, that blocks can contain many files, and a single file can require many blocks. This is an important detail
when reconstructing data.

#### Persistence

AetherFS persists data in an S3 compatible solution. Internally, it uses the MinIO Golang client to communicate with the
S3 API. We support reading common AWS environment variables, MinIO environment variables, or through command line
options. Similar to Git's object store and Dockers blob store, the AetherFS block store persists blocks in a directory
prefixed by the first two letters of the signature. Meanwhile, datasets are in a separate key space that allows hubs to
list datasets, and their tags without the need for a metadata file (i.e. through prefix scans.) For example:

```
blocks/{sha[:2]}/{sha[2:]}
blocks/1e/0d0f7ab123836cd69efb28cc62908c03002a11
blocks/1e/67d8b5474d6141c12da7798e494681e3d87440
blocks/1e/d939b2a6636c5a6085d0e885df0ac997c0d0a7
blocks/1e/fc7af3a2e5dfdd23163bd9e8c5b97059f56a37

datasets/{name}/{tag}
datasets/test/v21.10
datasets/test/latest
```

Since we store this information separately, implementing expiration and cleaning up older blocks is relatively easy to 
implement. Some datasets, like Maxmind, have terms of use that say new versions need to be replaced in a timely manner
and older copies are deleted upon request.

#### Caching

_Not yet implemented._

### Security & Privacy

This is an initial release. We'll add more on this later. Generally speaking, I like to take a reasonable approach
toward security. To me, that means that we ship security features as we need them instead of trying to force them in
right out the gate. That being said, while our initial release has no intent on supporting authentication, we're looking
at including it in our second.

So stay tuned until then... or not. I'm sure you know how these things work... 

#### Authentication

_Not yet implemented._

#### Authorization

_Not yet implemented._

#### Encryption at Rest

For the most part, AetherFS expects your small blob storage solution to provide this functionality. After an initial 
search, it seemed like most solutions provide some form of encryption at rest. Later on, we may add end-to-end
encryption support (assuming interest).

#### Encryption in Transit

Where possible, our systems leverage TLS certificates to encrypt communication between processes.
