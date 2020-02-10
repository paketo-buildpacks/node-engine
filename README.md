# Node Engine Cloud Native Buildpack

## Integration

The Node Engine CNB provides node as a dependency. Downstream buildpacks, like Yarn or NPM, can require the node dependency by generating a Build Plan TOML file that looks like the following:

```toml
[[requires]]

  # The name of the Node Engine dependency is "node". This value is considered
  # part of the public API for the buildpack and will not change.
  name = "node"

  # The version of the Node Engine dependency is not required. In the case it
  # is not specified, the buildpack will provide the default version, which can
  # be seen in the buildpack.toml file.
  # If you wish to request a specific version, the buildpack supports
  # specifying a semver constraint in the form of "10.*", "10.15.*", or even
  # "10.15.1".
  version = "10.15.1"

  # The Node Engine buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the Node Engine
    # depdendency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to run Node
    # during its build process, this flag should be set to true.
    build = true

    # Setting the cache flag to true will enable caching of the Node Engine
    # dependency between builds. The benefits of caching include improved build
    # speeds at the cost of a higher storage requirement to store the cached
    # layer contents.
    cache = true
```

## Usage

To package this buildpack for consumption:

```
$ ./scripts/package.sh
```

This builds the buildpack's Go source using `GOOS=linux` by default. You can
supply another value as the first argument to `package.sh`.
