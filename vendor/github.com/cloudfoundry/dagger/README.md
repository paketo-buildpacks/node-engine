# A Go library for working with V3 buildpacks

# Setup:

1. install the [pack cli tool](from https://github.com/buildpack/pack)
1. install docker
1. docker pull the following images
    ```
    docker pull cfbuildpacks/cflinuxfs3-cnb-experimental:build
    docker pull cfbuildpacks/cflinuxfs3-cnb-experimental:run
    ```
1. add the needed cflinuxfs3 stacks 
    ```
    pack add-stack org.cloudfoundry.stacks.cflinuxfs3 \
          --build-image cfbuildpacks/cflinuxfs3-cnb-experimental:build \
          --run-image cfbuildpacks/cflinuxfs3-cnb-experimental:run
    ```
1. Verify all tests pass using 
    ```
    go test -v ./...
    ```
