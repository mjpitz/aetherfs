# Architecture

* [Background](#background)
* [Overview](#overview)
  * [Goals](#goals)
  * [Features](#features)
* [Implementation](#implementation)
  * [Components](#components)
    * [aetherfs-agent](#aetherfs-agent)
    * [aetherfs-server](#aetherfs-server)
  * [Security & Privacy](#security--privacy)

## Background

While working at [Indeed][], many of our systems leveraged a producer-consumer architecture. In this pattern, services
can load an artifact containing data into memory in order to better service requests. These artifacts could be consumed
by a single service or shared across many services. Eventually, this developed into a platform called [RAD][] (short for
resilient artifact distribution).

Sometime after Indeed developed RAD internally, we saw a similar system open sourced from [Netflix][] called [Hollow][].
Hollow is a Java library used to distribute in-memory datasets. Unlike RAD's file-system based approach, Hollow stored 
everything in S3. However, both of these approaches had their own benefits and trade-offs.

Since leaving, I've often thought about what a modern take on this technology might look like. After spending some time
digging through internals of `git` and `docker`, I made a first pass at this. In the end, I was not satisfied with how
it came out so back to the drawing board I went.

[Indeed]: https://www.indeed.com
[RAD]: https://www.youtube.com/watch?v=lDXdf5q8Yw8
[Netflix]: https://netflix.com
[Hollow]: https://github.com/Netflix/hollow

## Overview

For the most part, this document focuses on the design of an AP system similar to Indeed's RAD system. Instead of using
bittorrent to replicate data, we opt for a simpler, replicated architecture.

### Features

  - [ ] Ideal for small to medium datasets (KB, MB, and GB ; not TB or PB)
  - [ ] Data encrypted in transit
  - [ ] Authentication and authorization controls over datasets
  - [ ] Support for any file type in any language
  - [ ] Backed by Amazon [S3 API][] (AWS S3, [MinIO][], etc)
  - [ ] Caches hot data amongst cluster peers using [groupcache][]
  - [ ] Single agent process with minimal API 
  - [ ] [Prometheus][] / [Grafana][] for usage tracking

[S3 API]: https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html
[MinIO]: https://min.io/
[groupcache]: https://github.com/golang/groupcache
[Prometheus]: https://prometheus.io/
[Grafana]: https://grafana.com/

## Implementation

[![](https://mermaid.ink/img/eyJjb2RlIjoiZ3JhcGggVERcbiAgICBwcm9kdWNlclxuICAgIHByb2R1Y2VyLWFnZW50W2FldGhlcmZzLWFnZW50XVxuXG4gICAgY29uc3VtZXJcbiAgICBjb25zdW1lci1hZ2VudFthZXRoZXJmcy1hZ2VudF1cblxuICAgIHNlcnZlci0xW2FldGhlcmZzLXNlcnZlcl1cbiAgICBzZXJ2ZXItMlthZXRoZXJmcy1zZXJ2ZXJdXG4gICAgc2VydmVyLTNbYWV0aGVyZnMtc2VydmVyXVxuXG4gICAgYXdzLXMzW0FXIFMzXVxuXG4gICAgc3ViZ3JhcGggcHJvZHVjZXItcG9kXG4gICAgICAgIHByb2R1Y2VyIC0tIGFldGhlcmZzLmFnZW50LnYxLkFnZW50QVBJL1B1Ymxpc2ggLS0-IHByb2R1Y2VyLWFnZW50XG4gICAgZW5kXG5cbiAgICBzdWJncmFwaCBjb25zdW1lci1wb2RcbiAgICAgICAgY29uc3VtZXIgLS0gYWV0aGVyZnMuYWdlbnQudjEuQWdlbnRBUEkvU3Vic2NyaWJlIC0tPiBjb25zdW1lci1hZ2VudFxuICAgIGVuZFxuXG4gICAgcHJvZHVjZXItYWdlbnQgLS0gYWV0aGVyZnMuZGF0YXNldC52MS5EYXRhc2V0QVBJL1B1Ymxpc2ggLS0-IHNlcnZlci0xXG4gICAgcHJvZHVjZXItYWdlbnQgLS0gYWV0aGVyZnMuYmxvY2sudjEuQmxvY2tBUEkvVXBsb2FkIC0tPiBzZXJ2ZXItMlxuICAgIHByb2R1Y2VyLWFnZW50IC0tPiBzZXJ2ZXItM1xuXG4gICAgY29uc3VtZXItYWdlbnQgLS0-IHNlcnZlci0xXG4gICAgY29uc3VtZXItYWdlbnQgLS0gYWV0aGVyZnMuYmxvY2sudjEuQmxvY2tBUEkvRG93bmxvYWQgLS0-IHNlcnZlci0yXG4gICAgY29uc3VtZXItYWdlbnQgLS0gYWV0aGVyZnMuZGF0YXNldC52MS5EYXRhc2V0QVBJL1N1YnNjcmliZSAtLT4gc2VydmVyLTNcblxuICAgIHNlcnZlci0xIC0tPiBhd3MtczNcbiAgICBzZXJ2ZXItMiAtLT4gYXdzLXMzXG4gICAgc2VydmVyLTMgLS0-IGF3cy1zM1xuIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQifSwidXBkYXRlRWRpdG9yIjpmYWxzZSwiYXV0b1N5bmMiOnRydWUsInVwZGF0ZURpYWdyYW0iOmZhbHNlfQ)](https://mermaid-js.github.io/mermaid-live-editor/edit/#eyJjb2RlIjoiZ3JhcGggVERcbiAgICBwcm9kdWNlclxuICAgIHByb2R1Y2VyLWFnZW50W2FldGhlcmZzLWFnZW50XVxuXG4gICAgY29uc3VtZXJcbiAgICBjb25zdW1lci1hZ2VudFthZXRoZXJmcy1hZ2VudF1cblxuICAgIHNlcnZlci0xW2FldGhlcmZzLXNlcnZlcl1cbiAgICBzZXJ2ZXItMlthZXRoZXJmcy1zZXJ2ZXJdXG4gICAgc2VydmVyLTNbYWV0aGVyZnMtc2VydmVyXVxuXG4gICAgYXdzLXMzW0FXIFMzXVxuXG4gICAgc3ViZ3JhcGggcHJvZHVjZXItcG9kXG4gICAgICAgIHByb2R1Y2VyIC0tIGFldGhlcmZzLmFnZW50LnYxLkFnZW50QVBJL1B1Ymxpc2ggLS0-IHByb2R1Y2VyLWFnZW50XG4gICAgZW5kXG5cbiAgICBzdWJncmFwaCBjb25zdW1lci1wb2RcbiAgICAgICAgY29uc3VtZXIgLS0gYWV0aGVyZnMuYWdlbnQudjEuQWdlbnRBUEkvU3Vic2NyaWJlIC0tPiBjb25zdW1lci1hZ2VudFxuICAgIGVuZFxuXG4gICAgcHJvZHVjZXItYWdlbnQgLS0gYWV0aGVyZnMuZGF0YXNldC52MS5EYXRhc2V0QVBJL1B1Ymxpc2ggLS0-IHNlcnZlci0xXG4gICAgcHJvZHVjZXItYWdlbnQgLS0gYWV0aGVyZnMuYmxvY2sudjEuQmxvY2tBUEkvVXBsb2FkIC0tPiBzZXJ2ZXItMlxuICAgIHByb2R1Y2VyLWFnZW50IC0tPiBzZXJ2ZXItM1xuXG4gICAgY29uc3VtZXItYWdlbnQgLS0-IHNlcnZlci0xXG4gICAgY29uc3VtZXItYWdlbnQgLS0gYWV0aGVyZnMuYmxvY2sudjEuQmxvY2tBUEkvRG93bmxvYWQgLS0-IHNlcnZlci0yXG4gICAgY29uc3VtZXItYWdlbnQgLS0gYWV0aGVyZnMuZGF0YXNldC52MS5EYXRhc2V0QVBJL1N1YnNjcmliZSAtLT4gc2VydmVyLTNcblxuICAgIHNlcnZlci0xIC0tPiBhd3MtczNcbiAgICBzZXJ2ZXItMiAtLT4gYXdzLXMzdFxuICAgIHNlcnZlci0zIC0tPiBhd3MtczNcbiIsIm1lcm1haWQiOiJ7XG4gIFwidGhlbWVcIjogXCJkZWZhdWx0XCJcbn0iLCJ1cGRhdGVFZGl0b3IiOmZhbHNlLCJhdXRvU3luYyI6dHJ1ZSwidXBkYXRlRGlhZ3JhbSI6ZmFsc2V9)

### Components

#### aetherfs-agent

The `aetherfs-agent` process is responsible for managing the local file-system. In Kubernetes, this should be run as a 
sidecar to the main process. Operations can be performed programmatically. More often, consumers can simply watch the 
file system for when a new version of a dataset becomes available. 

#### aetherfs-server

The `aetherfs-server` process translates data stored in S3 to the client. It provides a `DatasetAPI` that allows callers
to resolve information about datasets the user has access to.

### Security & Privacy

#### Encryption at Rest

For the most part, AetherFS expects your small blob storage solution to provide this functionality. After an initial 
search, it seemed like most provide some form of encryption at rest.

#### Encryption in Transit

Where possible, our systems leverage TLS certificates to encrypt communication between processes.
