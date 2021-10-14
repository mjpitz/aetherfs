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
  * [Interfaces](#interfaces)
    * [REST & gRPC](#rest--grpc)
    * [HTTP File Server](#http-file-server)
    * [FUSE File System](#fuse-file-system)
    * [Web](#web)
  * [Configuration](#configuration)
    * [Clustering](#clustering)
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

<!-- todo: add some pretext here -->

### Components

AetherFS is distributed as a single binary. Each component provides both a REST and gRPC interface. Since we leverage
streaming APIs, not all gRPC calls are available on the REST interface. Additionally, the REST interface provides an
[HTTP file server](#http-file-server) where files can be read directly.

#### AetherFS Hub

The AetherFS Hub is the primary component in AetherFS. It provides the core interfaces that are leverage by all other
components in AetherFS. The hub is responsible for managing the underlying storage tier and verifying the
authenticated clients have access to the desired dataset.

#### AetherFS Agent

The AetherFS agent is an optional sidecar process. It provides an application level cache for block data and can also
manage a local file system path (if enabled). It provides a special `AgentAPI` that can publish datasets 
programmatically.

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

<!-- todo, linux only feature (until OSX Fuse works again) -->

#### Web

A web interface is available under the `/ui` prefix. 

### Configuration

<!-- todo: add some pretext -->

#### Clustering

<!-- todo -->

#### Persistence

<!-- todo -->

#### Caching

<!-- todo -->

### Security & Privacy

<!-- todo: add some pretext -->

#### Authentication

<!-- todo -->

#### Authorization

<!-- todo -->

#### Encryption at Rest

For the most part, AetherFS expects your small blob storage solution to provide this functionality. After an initial 
search, it seemed like most solutions provide some form of encryption at rest. Later on, we may add end-to-end
encryption support (assuming interest).

#### Encryption in Transit

Where possible, our systems leverage TLS certificates to encrypt communication between processes.
