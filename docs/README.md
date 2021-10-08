AetherFS assists in the production, distribution, and replication of embedded databases and in-memory datasets. It
provides engineers with a platform to manage collections of files called datasets. AetherFS optimizes its use of the 
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

This project is under active development.

- Documentation
  - Architecture Document
- Features
  - HTTP file server for ease of interaction
  - REST and gRPC APIs for programmatic interaction
  - Optional agent that can manage a shared file systems
  - Efficiently persist and query information stored in [AWS S3][] (or compatible)
  - Authenticate using common schemes (such as OIDC)
  - Enforce access control around datasets
  - Encrypt data in transit and at rest
  - Built-in developer tools to help understand dataset performance and usage

[AWS S3]: https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html


### Expectations

This is a project I'm mostly iterating on in my free time. It's closed source, and I have no intent to open source. If
you're interested in learning more or getting updates, please sign up using the link below.


[![Project Interest Form][]](https://forms.gle/uCMy38ZLEchfNuka9)

[Project Interest Form]: https://img.shields.io/badge/-Project%20Interest%20Form-blue?style=for-the-badge
