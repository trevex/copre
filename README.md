# `copre`

`copre` is a small library for loading **co**nfiguration from multiple sources with a user-defined **pre**cedence. The sources include [`pflags`](https://github.com/spf13/pflag), environment-variables and files (bring your own file-format).

## Motivation

Depending on the application domain the precedence of loading configuration can differ.
For example a CLI tool might have a precendence such as `flags > env > file`.
However services run in a container might prefer a precendence similar to `env > file > flags`.

At the end of the day the Go ecosystem had plenty options to load configuration,
but not to compose its precendence, so hopefully this library accomodates that.
