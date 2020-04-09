# Node Engine Cloud Native Buildpack

## Integration

The Node Engine CNB provides node as a dependency. Downstream buildpacks, like
[Yarn](https://github.com/paketo-buildpacks/yarn-install-cnb) or
[NPM](https://github.com/paketo-buildpacks/npm-cnb), can require the node dependency
by generating a [Build Plan
TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the Node Engine dependency is "node". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
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
```

## Usage

To package this buildpack for consumption:

```
$ ./scripts/package.sh
```

This builds the buildpack's Go source using `GOOS=linux` by default. You can
supply another value as the first argument to `package.sh`.

