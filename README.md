# `copre`

`copre` is a small library for loading **co**nfiguration from multiple sources with a user-defined **pre**cedence. The sources include [`pflags`](https://github.com/spf13/pflag), environment-variables and files (bring your own file-format).

## Why

The Go ecosystem already has plenty of options to load configuration, so why another library? 
Well, this library aims to act as glue to ensure a specific precedence in loading configuration from different sources, so it hopefully accomodates the existing ecosystem.

Why the focus on precedence?
There is no right precendence, so it is important to be able to compose the precendence.
For CLI-tools the precendence might look as follows: `flags > env > file`.
However for an application running in a container on Kubernetes it might instead make sense to ensure a precendence such as `env > file > flags`.
