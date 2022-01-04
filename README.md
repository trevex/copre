# `copre`

[![Go Report Card](https://goreportcard.com/badge/github.com/trevex/copre)](https://goreportcard.com/report/github.com/trevex/copre)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/trevex/copre)](https://pkg.go.dev/github.com/trevex/copre#section-documentation)
[![Github Actions](https://github.com/trevex/copre/actions/workflows/tests.yaml/badge.svg)](https://github.com/trevex/copre/actions)
[![codecov](https://codecov.io/gh/trevex/copre/branch/main/graph/badge.svg?token=BMKV7KD2M8)](https://codecov.io/gh/trevex/copre)
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B29362%2Fgithub.com%2Ftrevex%2Fcopre.svg?type=shield)](https://app.fossa.com/projects/custom%2B29362%2Fgithub.com%2Ftrevex%2Fcopre?ref=badge_shield)



`copre` is a small library for loading **co**nfiguration from multiple sources with a user-defined **pre**cedence and merging them. The sources include [`pflags`](https://github.com/spf13/pflag), environment-variables and files (bring your own file-format).

## Overview

With `copre` it is straightforward to express how your configuration should be loaded.

`copre` provides:

* One-way to populate a configuration `struct`
* Struct-tags to specify options for environment variables and flags
* Minimal defaults, opt-in to features using options instead (intentionally explicit)
* Flexible `Loader`-composition as many passes as required (see example Y)
* Easy to extend (see example X)

## Install

```
go get github.com/trevex/copre
```

## Quickstart

The main entrypoint to loading configuration is the [`Load`](https://pkg.go.dev/github.com/trevex/copre#Load)-function.
The first argument is the pointer to the struct you want to populate and the rest a variadic list of [`Loader`](https://pkg.go.dev/github.com/trevex/copre#Loader) to process.

A simple example could look like this:
```go
type Config struct {
    Foo string `env:"FOO" flag:"foo" yaml:"foo"`
    Bar string `env:"BAR" yaml:"bar"` // Can only be set by env or file
    Baz string `yaml:"baz"` // In this example, can not be set by env or flag
}

// ...
cfg := Config{ Foo: "myDefaultValue" }
err := copre.Load(&cfg,
    copre.File("./config.yaml", yaml.Unmarshal, copre.IgnoreNotFound()),
    copre.Flag(flags), // assuming flags were setup prior
    copre.Env(copre.WithPrefix("MYAPP")), // by default no prefix, so let's set it explicitly
)
```
As no advanced options (e.g. [`ComputeEnvKey`](https://pkg.go.dev/github.com/trevex/copre#ComputeEnvKey)) are used, `env` and `flag` struct-tags have to be explicitly set,
if a field should be populated from those sources. However if an environment variable is not set or a flag with the corresponding name does not exist or has an empty value (e.g. empty string), the field will remain untouched. Therefore if no `Loader` sets a specific field, a value set prior to loading will remain in place.
In the above example the configuration-file to be loaded is optional as `copre.IgnoreNotFound()` was set.

If you want to learn more about `copre`, checkout the examples below or the [API documentation](https://pkg.go.dev/github.com/trevex/copre#section-documentation).

## Examples

### Using options

This example shows off lots of options and hopefully illustrates how you can use
options to make `copre` the glue that composes your configuration:
```go
package main

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/pflag"
	"github.com/trevex/copre"
	"gopkg.in/yaml.v3"
)

type Config struct {
	// Copre at least aims to support the same types as pflag for environment variables
	ListenIP   net.IP `flag:"listen-ip" env:"LISTEN_IP" yaml:"listenIP"`
	ListenPort int    `yaml:"listenPort"`
	// The Data field will not use a prefix for its environment variable!
	// So will be set by DATA rather than EXAMPLE_DATA
	Data    []byte `env:"DATA,noprefix,base64" flag:"data" yaml:"data"`
	Default string `env:"DEFAULT" yaml:"default"`
	Special string `superenv:"SPECIAL" flag:"special"`
}

func main() {
	cfg := Config{Default: "default"}

	flags := pflag.NewFlagSet("", pflag.ContinueOnError)
	flags.IP("listen-ip", net.IPv4(127, 0, 0, 1), "")
	flags.Int("listen-port", 8080, "")
	flags.BytesBase64("data", []byte{}, "")
	flags.String("default", "", "")
	flags.String("special", "", "")

	// For this example we provide some input data ourselves:
	err := flags.Parse([]string{"--listen-port=9090", "--special=foo"})
	if err != nil {
		panic(err)
	}
	os.Setenv("DATA", "MQ==")
	os.Setenv("SPECIAL", "bar")

	// Let's load the config
	err = copre.Load(&cfg,
		// Okay, here is a little trick, we want to use the pflag defaults in our struct.
		// So we run our first pass over the flags with IncludeUnchanged and later without.
		copre.FlagSet(flags,
			copre.IncludeUnchanged(),
			// Compute flag names for fields without a "flag"-tag using kebab-case
			copre.ComputeFlagName(copre.KebabCase),
		),
		copre.File( // We need at least one file and the unmarshal function
			"./first.yaml", yaml.Unmarshal,
			// But we can add more files to check
			copre.AppendFilePaths("./second.yaml", "./third.yaml"),
			// By default the first will be unmarshalled, but we can also merge all available files
			copre.MergeFiles(),
			// We can provide the following option if no file is okay as well
			copre.IgnoreNotFound(),
		),
		copre.Env(
			// Prefix all environment variables retrieved with EXAMPLE unless noprefix is set in tag
			copre.WithPrefix("EXAMPLE"),
			// Compute environment variable names for fields without "env"-tag
			copre.ComputeEnvKey(copre.UpperSnakeCase),
		),
		copre.FlagSet(flags, copre.ComputeFlagName(copre.KebabCase)),
		copre.Env(
			// You can also change the tag used, to allow multiple sets of precedences
			// or avoid compatiblity issues with other libraries
			copre.OverrideEnvTag("superenv"), // NOTE: similar functionality exists for flags
		),
	)

	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", cfg)
	// Prints:
	// {ListenIP:127.0.0.1 ListenPort:9090 Data:[49] Default:default Special:bar}
}
```

### Validating configuration

### Custom `Loader`

### Using Kubernetes ConfigMap

## Q & A

### Why?

Depending on the application domain the precedence of loading configuration can differ.
For example a CLI tool might have a precendence such as `flags > env > file`.
However services run in a container might prefer a precendence similar to `env > file > flags`.

At the end of the day the Go ecosystem had plenty options to load configuration,
but not to compose its precendence, so hopefully this library accomodates that.
