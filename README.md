# Paketo Buildpack for Node Engine

## `docker.io/paketobuildpacks/node-engine`

The Node Engine CNB provides the Node binary distribution. The buildpack
installs the Node binary distribution onto the `$PATH` which makes it available
for subsequent buildpacks and in the final running container. Examples of
buildpacks that might use the Node binary distribution are the [NPM
CNB](https://github.com/paketo-buildpacks/npm) and [Yarn Install
CNB](https://github.com/paketo-buildpacks/yarn-install)

## Version support

node-engine will include Node.js versions which are supported as `LTS` in the
community as well as the active `current` release. When a Node.js version goes
End of Life (EOL) in the community it may be removed from node-engine
any time after that.

For more information on what versions are `LTS` and `current` refer to Node.js
projects [Release Schedule](https://github.com/nodejs/release#release-schedule).

## Integration

The Node Engine CNB provides `node` and `npm` as dependencies. Downstream buildpacks, like
[Yarn Install CNB](https://github.com/paketo-buildpacks/yarn-install) or
[NPM CNB](https://github.com/paketo-buildpacks/npm), can require the `node` dependency
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
  # specifying a semver constraint in the form of "15.*", "15.14.*", or even
  # "15.14.0".
  version = "15.14.0"

  # The Node Engine buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the Node Engine
    # depdendency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to run Node
    # during its build process, this flag should be set to true.
    build = true

    # Setting the launch flag to true will ensure that the Node Engine
    # dependency is available on the $PATH for the running application. If you are
    # writing an application that needs to run node at runtime, this flag should
    # be set to true.
    launch = true
```

Or they can require both `node` and `npm` using a Build Plan that looks like the following:

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
  # specifying a semver constraint in the form of "15.*", "15.14.*", or even
  # "15.14.0".
  version = "15.14.0"

  # The Node Engine buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the Node Engine
    # depdendency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to run Node
    # during its build process, this flag should be set to true.
    build = true

    # Setting the launch flag to true will ensure that the Node Engine
    # dependency is available on the $PATH for the running application. If you are
    # writing an application that needs to run node at runtime, this flag should
    # be set to true.
    launch = true

[[requires]]

  # The name of the Npm dependency is "npm". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "npm"
```

## Usage

To package this buildpack for consumption:

```
$ ./scripts/package.sh --version <version-number>
```

This will create a `buildpackage.cnb` file under the `build` directory which you
can use to build your app as follows:
`pack build <app-name> -p <path-to-app> -b build/buildpackage.cnb`

## Configurations

### Specifying a Node version

To specify the version of the Node that is installed, set the `$BP_NODE_VERSION`
environment variable at build time either directly (ex. `pack build my-app
--env BP_NODE_VERSION=~15`) or through a [`project.toml`
file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md)

```shell
$BP_NODE_VERSION="~15"
```

You can also specify a node version via an `.nvmrc` or `.node-version` file, also at the application directory root.

### Enabling memory optimization

To specify the use of memory optimization, set the `$BP_NODE_OPTIMIZE_MEMORY`
environment variable at build time either directly (ex. `pack build my-app
--env BP_NODE_OPTIMIZE_MEMORY=true`) or through a [`project.toml`
file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md)

```shell
$BP_NODE_OPTIMIZE_MEMORY="true"
```

### Specifying a project path

To specify a project subdirectory to be used as the root of the app, please use
the `BP_NODE_PROJECT_PATH` environment variable at build time either directly
(ex. `pack build my-app --env BP_NODE_PROJECT_PATH=./src/my-app`) or through a
[`project.toml`
file](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md).
This could be useful if your app is a part of a monorepo.

### Enabling Inspector for Remote Debugging

To enable the Inspector set the `BPL_DEBUG_ENABLED` environment variable at launch time. Optionally, you can specify the `BPL_DEBUG_PORT` environment variable to use a specific port.

```shell
$BPL_DEBUG_ENABLED="true"
$BPL_DEBUG_PORT="9009"
```

For more information on debugging, see [Official Documentation](https://nodejs.org/en/docs/guides/debugging-getting-started)

### Include python during build process

To require [python](https://github.com/paketo-buildpacks/cpython) during the build process, set the `BP_NODE_INCLUDE_BUILD_PYTHON` environment variable at build time. You can set it either directly:

```shell
pack build my-app --builder paketobuildpcks/builder-jammy-base \
  --env BP_NODE_INCLUDE_BUILD_PYTHON=true
```

or through a [`project.toml`](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md) file.

This is necessary for compiling native modules during `npm install` process, as `node-gyp` requires Python to complete this process.

Note that the `BP_NODE_INCLUDE_BUILD_PYTHON` variable is not required in the following cases:

- The [builder-jammy-full](https://github.com/paketo-buildpacks/builder-jammy-full) builder as python is already provided by the build image.
- The UBI builders ([ubi-8-builder](https://github.com/paketo-buildpacks/builder-ubi8-base), [ubi-9-builder](https://github.com/paketo-buildpacks/ubi-9-builder), etc.), as Python is being provided by the extension during build time.

## Run Tests

To run all unit tests, run:

```
./scripts/unit.sh
```

To run all integration tests, run:

```
/scripts/integration.sh
```
