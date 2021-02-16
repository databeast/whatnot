# Usage

## Installation

`go get github.com/databeast/whatnot`



#### Selecting Compile-Time functionality with Optional Build Tags

_metrics_ - enabling this build tag will register Prometheus metrics from Whatnot into the prometheus metrics registry if
your application exports Prometheus metrics (or makes them available via HTTP by way of `promhttp`), you should see the
following additional metrics available amongst your existing ones:

_errortraces_ - enabling this build tag will switch the base `error` type returned, from GO's internal `errors` package
to stacktrace-enabled errors provided by the [github.com/pkg/errors](github.com/pkg/errors) package

### Instantiating a Manager and populating Namespaces

    NewNameSpaceManager() 

    RegisterNamespace()

    RegisterAbsolutePath() 

### Adding Individual Keys, Fetching Existing Ones

fetch the pathelement for an absolute path.

fetch the pathelement for a relative subpath of a Path Element






