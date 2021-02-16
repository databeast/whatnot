# Usage

## Installation


## Selecting Compile-Time functionality with Optional Build Tags

_metrics_ - enabled this build tag will register Prometheus metrics from Whatnot into the prometheus metrics registry if
your application exports Prometheus metrics (or makes them available via HTTP by way of `promhttp`), you should see the
following additional metrics available amongst your existing ones:

_errortraces_




