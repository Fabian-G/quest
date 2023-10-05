# Quest

[![Build status](https://img.shields.io/github/actions/workflow/status/Fabian-G/quest/test.yml)](https://github.com/Fabian-G/quest/actions)
[![Release](https://img.shields.io/github/downloads-pre/Fabian-G/quest/latest/total)](https://github.com/Fabian-G/quest/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/Fabian-G/quest)](https://goreportcard.com/report/github.com/Fabian-G/quest)

Note: This software is pre v1. Config file format, CLI or query language may change at any time without warning.

## Installation

```bash
go install github.com/Fabian-G/quest@latest
```
Make sure `"$(go env GOPATH)/bin` is in your `PATH`.

or download the precompiled binary from the [Release Page](https://github.com/Fabian-G/quest/releases).

## Documentation

Documentation can be found [here](https://fabian-g.github.io/quest)
## Basic Usage

![basic usage](examples/demo/basic.gif)

## Edit subsets

![edit](examples/demo/edit.gif)

## Define Views

```toml
# ~/.config/quest/config.toml
[views.important]
query = 'priority >= prioC'
sort = ["-priority", "project"]
clean = ["+ALL","@ALL"]
projection = ["line","priority","tag:do","projects","contexts","description"]
```

```bash
$ quest important
 #  Priority  Projects       Description         
 1  (A)       +destroy-ring  assemble fellowship
 3  (A)       +destroy-ring  Loose Gandalf to Balrog 
```
