# AetherFS Architecture

* [Background](#background)
  * [Motivation](#motivation)
  * [Concepts](#concepts)
* [Overview](#overview)
  * [Requirements](#requirements)
  * [Components](#components)
    * [AetherFS Agent](#aetherfs-agent)
    * [AetherFS Server](#aetherfs-server)
* [Implementation](#implementation)
  * [Interfaces](#interfaces)
    * [HTTP File Server](#http-file-server)
    * [Agent API](#agent-api)
    * [Block API](#block-api)
    * [Dataset API](#dataset-api)
  * [Configuration](#configuration)
    * [Clustering](#clustering)
    * [Persistence](#persistence)
    * [Caching](#caching)
  * [Security & Privacy](#security--privacy)
    * [Authentication](#authentication)
    * [Authorization](#authorization)
    * [Encryption at Rest](#encryption-at-rest)
    * [Encryption in Transit](#encryption-in-transit)

## Background

While working at [Indeed][], many of our systems leveraged a producer-consumer architecture. In this pattern, services
can load an artifact containing data into memory in order to better service requests. These artifacts could be consumed
by a single service or shared across many services. Eventually, this developed into a platform called [RAD][] (short for
resilient artifact distribution).

Sometime after Indeed developed RAD internally, we saw a similar system open sourced from [Netflix][] called [Hollow][].
Hollow is a Java library used to distribute in-memory datasets. Unlike RAD's file-system based approach, Hollow stored 
everything in S3. While I have not used Hollow myself, I can see the utility it provides to Java ecosystem.

[Indeed]: https://www.indeed.com
[RAD]: https://www.youtube.com/watch?v=lDXdf5q8Yw8
[Netflix]: https://netflix.com
[Hollow]: https://github.com/Netflix/hollow

### Motivation

Since leaving Indeed, I've often thought about what a modern take on this technology might look like. In addition to
this curiosity, I've found myself wanting a similar solution that can be used on edge or IoT devices where storage is 
limited or non-existent.

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

  - [ ] Efficiently use [AWS S3][] (or equivalent) to store dataset information.
  - [ ] Information should be encrypted in transit and at rest.
  - [ ] Authenticate clients (users and services) using common schemes (OIDC, Basic).
  - [ ] Enforce access controls around datasets.
  - [ ] Provide numerous interfaces to manage and access information in the system.
  - [ ] Build in developer tools to help artifact producers understand the performance of their datasets

### Components

#### AetherFS Agent

#### AetherFS Server

## Implementation

<!--
[![](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggVERcbiAgICBwcm9kdWNlclxuICAgIHByb2R1Y2VyLWFnZW50W2FldGhlcmZzLWFnZW50XVxuXG4gICAgY29uc3VtZXJcbiAgICBjb25zdW1lci1hZ2VudFthZXRoZXJmcy1hZ2VudF1cblxuICAgIHNlcnZlci0xW2FldGhlcmZzLXNlcnZlcl1cbiAgICBzZXJ2ZXItMlthZXRoZXJmcy1zZXJ2ZXJdXG4gICAgc2VydmVyLTNbYWV0aGVyZnMtc2VydmVyXVxuXG4gICAgYXdzLXMzW0FXIFMzXVxuXG4gICAgc3ViZ3JhcGggcHJvZHVjZXItcG9kXG4gICAgICAgIHByb2R1Y2VyIC0tIGFldGhlcmZzLmFnZW50LnYxLkFnZW50QVBJL1B1Ymxpc2ggLS0-IHByb2R1Y2VyLWFnZW50XG4gICAgZW5kXG5cbiAgICBzdWJncmFwaCBjb25zdW1lci1wb2RcbiAgICAgICAgY29uc3VtZXIgLS0gYWV0aGVyZnMuYWdlbnQudjEuQWdlbnRBUEkvU3Vic2NyaWJlIC0tPiBjb25zdW1lci1hZ2VudFxuICAgIGVuZFxuXG4gICAgcHJvZHVjZXItYWdlbnQgLS0gYWV0aGVyZnMuZGF0YXNldC52MS5EYXRhc2V0QVBJL1B1Ymxpc2ggLS0-IHNlcnZlci0xXG4gICAgcHJvZHVjZXItYWdlbnQgLS0gYWV0aGVyZnMuYmxvY2sudjEuQmxvY2tBUEkvVXBsb2FkIC0tPiBzZXJ2ZXItMlxuICAgIHByb2R1Y2VyLWFnZW50IC0tPiBzZXJ2ZXItM1xuXG4gICAgY29uc3VtZXItYWdlbnQgLS0-IHNlcnZlci0xXG4gICAgY29uc3VtZXItYWdlbnQgLS0gYWV0aGVyZnMuYmxvY2sudjEuQmxvY2tBUEkvRG93bmxvYWQgLS0-IHNlcnZlci0yXG4gICAgY29uc3VtZXItYWdlbnQgLS0gYWV0aGVyZnMuZGF0YXNldC52MS5EYXRhc2V0QVBJL1N1YnNjcmliZSAtLT4gc2VydmVyLTNcblxuICAgIHNlcnZlci0xIC0tPiBhd3MtczNcbiAgICBzZXJ2ZXItMiAtLT4gYXdzLXMzXG4gICAgc2VydmVyLTMgLS0-IGF3cy1zM1xuIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQifSwidXBkYXRlRWRpdG9yIjpmYWxzZSwiYXV0b1N5bmMiOnRydWUsInVwZGF0ZURpYWdyYW0iOmZhbHNlfQ)](https://mermaid-js.github.io/mermaid-live-editor/edit/#eyJjb2RlIjoiZ3JhcGggVERcbiAgICBwcm9kdWNlclxuICAgIHByb2R1Y2VyLWFnZW50W2FldGhlcmZzLWFnZW50XVxuXG4gICAgY29uc3VtZXJcbiAgICBjb25zdW1lci1hZ2VudFthZXRoZXJmcy1hZ2VudF1cblxuICAgIHNlcnZlci0xW2FldGhlcmZzLXNlcnZlcl1cbiAgICBzZXJ2ZXItMlthZXRoZXJmcy1zZXJ2ZXJdXG4gICAgc2VydmVyLTNbYWV0aGVyZnMtc2VydmVyXVxuXG4gICAgYXdzLXMzW0FXIFMzXVxuXG4gICAgc3ViZ3JhcGggcHJvZHVjZXItcG9kXG4gICAgICAgIHByb2R1Y2VyIC0tIGFldGhlcmZzLmFnZW50LnYxLkFnZW50QVBJL1B1Ymxpc2ggLS0-IHByb2R1Y2VyLWFnZW50XG4gICAgZW5kXG5cbiAgICBzdWJncmFwaCBjb25zdW1lci1wb2RcbiAgICAgICAgY29uc3VtZXIgLS0gYWV0aGVyZnMuYWdlbnQudjEuQWdlbnRBUEkvU3Vic2NyaWJlIC0tPiBjb25zdW1lci1hZ2VudFxuICAgIGVuZFxuXG4gICAgcHJvZHVjZXItYWdlbnQgLS0gYWV0aGVyZnMuZGF0YXNldC52MS5EYXRhc2V0QVBJL1B1Ymxpc2ggLS0-IHNlcnZlci0xXG4gICAgcHJvZHVjZXItYWdlbnQgLS0gYWV0aGVyZnMuYmxvY2sudjEuQmxvY2tBUEkvVXBsb2FkIC0tPiBzZXJ2ZXItMlxuICAgIHByb2R1Y2VyLWFnZW50IC0tPiBzZXJ2ZXItM1xuXG4gICAgY29uc3VtZXItYWdlbnQgLS0-IHNlcnZlci0xXG4gICAgY29uc3VtZXItYWdlbnQgLS0gYWV0aGVyZnMuYmxvY2sudjEuQmxvY2tBUEkvRG93bmxvYWQgLS0-IHNlcnZlci0yXG4gICAgY29uc3VtZXItYWdlbnQgLS0gYWV0aGVyZnMuZGF0YXNldC52MS5EYXRhc2V0QVBJL1N1YnNjcmliZSAtLT4gc2VydmVyLTNcblxuICAgIHNlcnZlci0xIC0tPiBhd3MtczNcbiAgICBzZXJ2ZXItMiAtLT4gYXdzLXMzdFxuICAgIHNlcnZlci0zIC0tPiBhd3MtczNcbiIsIm1lcm1haWQiOiJ7XG4gIFwidGhlbWVcIjogXCJkZWZhdWx0XCJcbn0iLCJ1cGRhdGVFZGl0b3IiOmZhbHNlLCJhdXRvU3luYyI6dHJ1ZSwidXBkYXRlRGlhZ3JhbSI6ZmFsc2V9)
-->

### Interfaces

#### HTTP File Server

#### Agent API

#### Block API

#### Dataset API

### Configuration

#### Clustering

<!-- how are clusters of nodes formed -->

#### Persistence

<!-- how and where is information stored -->

#### Caching

<!-- how and where is information cached -->

### Security & Privacy

#### Authentication

<!-- how are users and systems authenticated -->

#### Authorization

#### Encryption at Rest

For the most part, AetherFS expects your small blob storage solution to provide this functionality. After an initial 
search, it seemed like most solutions provide some form of encryption at rest.

#### Encryption in Transit

Where possible, our systems leverage TLS certificates to encrypt communication between processes.
