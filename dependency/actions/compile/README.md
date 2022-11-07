Running compilation locally:

1. Build the build environment:
```
docker build -t compilation -f <target>.Dockerfile dependency/actions/compile
```

2. Make the output directory:
```
mkdir <output dir>
```

3. Run compilation and use a volume mount to access it:
```
docker run -v <output dir>:$PWD --rm compilation --version <version> --output-dir $PWD --target <target>
```
