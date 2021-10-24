AetherFS assists in the production, distribution, and replication of embedded databases and in-memory datasets.
You can think of it like [Docker][], but for data.

[Docker]: https://docker.com

AetherFS provides engineers with a platform to manage collections of files called datasets. It optimizes its use of the 
underlying blob store (AWS S3 or equivalent) to reduce cost to operators and improve performance for end users.

_Why not use S3 directly or a file server?_

While this is an option, there are several problems that arise with this solution. For example, to produce two 
references to the same dataset, you must upload the same set of files twice. If you want to produce three references, 
then three times (and so on). This comes at a cost of additional time in your pipeline and storage costs.

Instead, producers tag datasets in AetherFS. A tag can refer to a specific version ([semantic][] or [calendar][]) or a
channel that consumers can subscribe to (`latest`, `stable`, etc.). Instead of storing entire snapshots of datasets
in each version, AetherFS removes duplicated blocks between them. This allows clients to re-use blocks of data and only
download new or updated portions.

[semantic]: https://semver.org
[calendar]: https://calver.org

## Status

This project is under active development. The lists below detail aspirational features and documentation.

- Documentation
  - [Architecture Document](https://github.com/mjpitz/aetherfs/wiki/Architecture)
  - [General Wiki](https://github.com/mjpitz/aetherfs/wiki)
- Features
  - HTTP file server for ease of interaction
  - REST and gRPC APIs for programmatic interaction
  - Optional agent that can manage a shared or FUSE file system
  - Efficiently persist and query information stored in [AWS S3][]
  - Authenticate using common schemes (such as OIDC)
  - Enforce access control around datasets
  - Encrypt data in transit and at rest
  - Built-in developer tools to help understand dataset performance and usage

[AWS S3]: https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html


## Expectations & Roadmap

Since I'm mostly iterating on this project in my free time, I plan on using [calendar versioning][]. Bugfixes and minor
features can be introduced in any patch version but any major feature should wait for the next release. Releases happen 
in October, February, and June (every 4 months). Any security issues will be addressed in a timely manner, regardless of
release schedule.

[calendar versioning]: https://calver.org

### v21.10

This will be the initial release of AetherFS. It includes the "essentials".

- Single binary containing all components.
- Command to run an AetherFS data hub.
- Command to upload to and tag datasets in AetherFS.
- Command to download tagged datasets from AetherFS.
- Minimal web interface.

### v22.02

As the second major release of the AetherFS system, this will include additional security measures and helps simplify
interaction for end users (provided there's interest in the system).

- Command to run an agent process with a FUSE file system.
- Block caching to improve performance and usage of S3.
- Command to authenticate clients.
- Enforce access controls around datasets.
- Data encrypted in transit.
