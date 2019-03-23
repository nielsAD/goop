Build From Source
=================

If you would like to run the development version, you will have to compile the project from source. Goop is built with [`Go`](https://golang.org/), but requires an additional `C++` compiler for external libraries. All dependencies are shipped with the source as git submodules and the build process is managed by a Makefile.

Requirements
------------

  * `go` 1.10+
  * `gcc`/`clang`
  * `Make`, `CMake`
  * `libgmp`, `libbzip2`, `zlib`


#### Ubuntu / Debian
```
apt install build-essential cmake git golang-go libgmp-dev libbz2-dev zlib1g-dev
```

#### macOS

Using [Homebrew](https://brew.sh/).

```
brew install cmake git go gmp bzip2 zlib
```

#### Windows

Using [MSYS2](https://www.msys2.org/).

```
pacman -S git mingw-w64-x86_64-toolchain mingw-w64-x86_64-go mingw-w64-x86_64-cmake
```

Build
-----

After all requirements are installed, the build process is managed with `make`.  
Binaries will be placed in the `./bin/` directory.

```shell
# Download source
git clone https://github.com/nielsAD/goop.git; cd goop
git submodule update --init --recursive

# Run tests
make test

# Build release files in ./bin/
make release
```

?> **TIP:** After making code changes, run `make test` to ensure all tests pass and everything is still functional.
