GoWarcraft3/goop
===========
[![Build Status](https://travis-ci.org/nielsAD/gowarcraft3.svg?branch=master)](https://travis-ci.org/nielsAD/gowarcraft3)
[![Build status](https://ci.appveyor.com/api/projects/status/a5cecrpfo0pe14ux/branch/master?svg=true)](https://ci.appveyor.com/project/nielsAD/gowarcraft3)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

Goop (GO OPerator) is a BNCS Channel Operator.

Usage
-----

`./goop [config_file]`

Configuration
-------------
See [`config.toml.example`](/config.toml.example).


Download
--------

Official binaries for tools are [available](https://github.com/nielsAD/goop/releases/latest). Simply download and run.

_Note: additional dependencies may be required (see [build instructions](/README.md#build))._

Build
-----

```bash
# Linux dependencies
apt-get install --no-install-recommends -y build-essential cmake git golang-go libgmp-dev libbz2-dev zlib1g-dev

# OSX dependencies
brew install cmake git go gmp bzip2 zlib

# Windows dependencies (use MSYS2 -- https://www.msys2.org/)
pacman --needed --noconfirm -S git mingw-w64-x86_64-toolchain mingw-w64-x86_64-go mingw-w64-x86_64-cmake

# Download vendor submodules
git submodule update --init --recursive

# Run tests
make test

# Build release files in ./bin/
make release
```