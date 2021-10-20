# AetherFS Architecture

* [Background](#background)
  * [Motivation](#motivation)
  * [Concepts](#concepts)
* [Overview](#overview)
  * [Requirements](#requirements)
* [Implementation](#implementation)
  * [Components](#components)
    * [AetherFS Hub](#aetherfs-hub)
    * [AetherFS Agent](#aetherfs-agent)
    * [AetherFS CLI](#aetherfs-cli)
  * [Interfaces](#interfaces)
    * [REST & gRPC](#rest--grpc)
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
everything in S3. While I have not used Hollow myself, I see the utility it provides to Java ecosystem.

[Indeed]: https://www.indeed.com
[RAD]: https://www.youtube.com/watch?v=lDXdf5q8Yw8
[Netflix]: https://netflix.com
[Hollow]: https://github.com/Netflix/hollow

### Motivation

Since leaving Indeed, I've often thought about what a modern take on this technology might look like. In addition to
this curiosity, I've found myself wanting a similar solution that can be used on edge or IoT devices where storage may
be limited or non-existent.

### Concepts

**Dataset**

At Indeed, we referred to these as "artifacts," but I often found the term to be too generic in conversation. In
AetherFS, we refer to collections of information as a _dataset_. Datasets can be tagged, similar to containers. This
allows publishers to manage their own history and channels for consumers.

For example, you might maintain a `stable` tag that contains the latest stable version of the dataset. To help insulate
consumers, you might also manage a `next` tag that contains the next version of the dataset. This allows consumers to
follow the `stable` tag in production, and the `next` tag in development.

You can also follow your standard [semantic][] or [calendar][] versioning tags to maintain a history of all versions of the
dataset. This is particularly useful should you need to rollback a change to a dataset.

[semantic]: https://semver.org/
[calendar]: https://calver.org/

_**NOTE:** In theory, this system could be used as a generalized package manager, in which the term "artifact" would be
appropriate. However, that is not my intent for this project._

**Blocks**

When clients push a dataset into AetherFS, its contents are broken up into fixed sized _blocks_. This allows smaller
files to be stored as a single block and larger ones to be broken up into multiple smaller ones. In the end, their goal
is to reduce the amount of data between versions and reduce the number of calls made to the backend.

Blocks are immutable, which allows them to be cached amongst your agents. This allows hot data to be read from your
peers instead of making a call to your underlying storage tier.

**BitTorrent**

Indeed's RAD ecosystem used the [BitTorrent][] protocol to replicate information around the world. This was done to
reduce the data load on the producer machine. However, for Indeed to leverage BitTorrent, they needed to modify the
torrent manifest to propagate the last modified times for a file. This adds a maintenance burden since we would then
need to maintain a [fork][]. Similarly, the academic community has latched onto this at [Florida State University][]
where they use BitTorrent to share large datasets between researchers.

While AetherFS does not use BitTorrent, we do lift some concepts from the protocol. For example, our dataset manifest
uses a similar structure to a BitTorrent manifest since we deal with similar structures. Similar to BitTorrent, AetherFS
chunks the data into blocks, optimized for storage in [AWS S3][] (or equivalent). When read from S3, we break blocks
down into smaller, cache optimized blocks. For better performance, we can tier the sizes of our caching layers. This
will be explained more in depth later on.

[Florida State University]: https://web.archive.org/web/20130402200554/https://www.hpc.fsu.edu/index.php?option=com_wrapper&view=wrapper&Itemid=80
[BitTorrent]: https://en.wikipedia.org/wiki/BitTorrent
[fork]: https://github.com/indeedeng/ttorrent
[AWS S3]: https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html

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

## Implementation

<a href="docs/assets/seen-stored-cached.png">
<img src="https://aetherfs.tech/assets/overview.png" align="right" width="60%"/>
</a>

AaetherFS deploys as a simple client-server architecture, with a few caveats. It's distributed as a single binary making
the full suite of components accessible to users.

### Components

#### AetherFS Hub

![role: server](https://img.shields.io/badge/role-server-white?style=for-the-badge)
![interfaces: grpc, file server, rest, web](https://img.shields.io/badge/interfaces-grpc,%20file%20server,%20rest,%20web-white?style=for-the-badge)

The AetherFS Hub is the primary component in AetherFS. It provides the core interfaces that are leverage by all other
components in AetherFS. The hub is responsible for managing the underlying storage tier and verifying the
authenticated clients have access to the desired dataset.

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

The AetherFS agent is an optional sidecar process. It provides an application level cache for block data and can also
manage a local file system path (if enabled). It provides a special `AgentAPI` that can publish datasets 
programmatically.

_Not yet implemented._

```
$ aetherfs run agent -h
```

#### AetherFS CLI

![role: client](https://img.shields.io/badge/role-client-white?style=for-the-badge)
![interfaces: command line, fuse](https://img.shields.io/badge/interfaces-command%20line,%20fuse-white?style=for-the-badge)

End users can interact with a command line interface (CLI) to push and pull datasets from AetherFS hubs. It's the
primary point of interaction for operators and engineers. It also contains the necessary code to run and operate your
own hub and agent processes.

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

### Interfaces

#### REST & gRPC

Each component of the architecture provides a REST and gRPC interface for communication. While primarily used by
internal components, these interfaces can be used by calling applications as well. However, AetherFS expects callers to
interact with one of our other interfaces as they abstract away the complexity of the underlying storage.

For the most part, AetherFS's interfaces are inspired by Docker and Git. All REST routes sit under the `/api` prefix.

**DatasetAPI**

The `DatasetAPI` allows callers to interact with various datasets stored within AetherFS. Dataset manifests contain a
complete list of files within the dataset, their sizes, and last modified timestamps. The manifests also contain a list 
of blocks that are required to construct the dataset. Using these components, clients can piece together the underlying
files.

**BlockAPI**

The `BlockAPI` gives callers direct access to the block data. You must use the `DatasetAPI` to obtain block references.
The `BlockAPI` does not provide callers with the ability to list blocks (intentional design decision). An entire block
can be read at a time, or just part of one.

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

#### Packing 

<a href="docs/assets/seen-stored-cached.png">
<img src="https://aetherfs.tech/assets/seen-stored-cached.png" align="right" width="40%"/>
</a>

When uploading files to AetherFS, we pack all files found in a target directory, zip, or tarball into a single 
contiguous blob. This large blob is broken into smaller blocks that are ideally sized for your storage layer. For 
example, Amazon Athena documentation suggests using S3 objects between 256MiB and 1GiB to optimize network bandwidth.
<!-- needs citation -->

Each dataset can choose their own block size, ideally striving to get the most reuse between versions. While producers 
have control over the size of the blocks that are stored in AetherFS, they do not control the size of the cacheable 
parts. This allows consumers of datasets to tune their usage based by adding more memory or disk where they need to.

#### Persistence

<!-- todo -->

#### Caching

_Not yet implemented._

### Security & Privacy

<!-- todo: add some pretext -->

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
