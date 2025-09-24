# Dependency

Pre-compiled distributions of Python are provided for the Paketo stacks (i.e.
`io.buildpacks.stack.jammy` and `io.buildpacks.stacks.bionic`).

Source distributions of Python are provided for all other linux stacks.

This directory contains scripts and GitHub Actions to facilitate the following:
* Identifying when there is a new version of Python available
* Compiling Python against all supported stacks
* Packing the Python source for use in all other stacks (i.e. where we do not
provide pre-compiled binaries of python)

## Running locally

Running the steps locally can be useful for iterating on the compilation process
(e.g. changing compilation options) as well as debugging.

### Retrieval

Retrieve latest versions with:

```
cd ./retrieval

go run main.go \
  --buildpack-toml-path ../../buildpack.toml \
  --output /path/to/retrieved.json
```

See [retrieval/README.md](retrieval/README.md) for more details.

### Compilation

To compile on Ubuntu 22.04 (Jammy):

```
docker build \
  --tag cpython-compilation-jammy \
  --file ./actions/compile/jammy.Dockerfile \
  ./actions/compile

output_dir=$(mktemp -d)

docker run \
  --volume $output_dir:/tmp/compilation \
  cpython-compilation-jammy \
    --outputDir /tmp/compilation \
    --target jammy \
    --version 3.10.7
```

See [actions/compile/README.md](actions/compile/README.md) for more details.