# kayentactl

A CLI tool for running canary analysis using [Kayenta]().

## Prerequisites

In order to use `kayentactl` you must have an instance of Kayenta configured and running.

## Usage


### Canary config
`kayentactl` requires a user-defined canary config to run. This config is what Kayenta uses to determine which metrics
to consider when evaluating whether a canary is performing properly. When deciding which metrics to measure, you can
reference this [best practices guid](https://spinnaker.io/guides/user/canary/best-practices/).

By default, `kayentactl` reads the canary config from `canary.json`. Both JSON and YAML formats are supported so if you're
more comfortable with YAML, feel free to use it! Below is an example canary config that uses Datadog to measure IO, CPU,
and Memory utilization. 

<details><summary>Example Canary config</summary>
<p>

```yaml
executionRequest:
  lifetimeDurationMins: 10
  analysisIntervalMins: 1
  scopes:
  - controlScope: host:gke-armory-shared-ser-devops-gke-test-96cd73b4-zr3q.c.shared-services-pb.internal
    controlOffset: 0
    experimentScope: host:gke-armory-shared-ser-devops-gke-test-96cd73b4-10cz.c.shared-services-pb.internal
    extendedScopeParams:
      application: datadog-cn
    scopeName: default
    step: 300
  thresholds:
    marginal: "0"
    pass: "99"
canaryConfig:
  classifier:
    groupWeights:
      MEM: 40
      CPU: 35
      IO: 25
  configVersion: "1"
  judge:
    name: NetflixACAJudge-v1.0
  metrics:
  - groups:
    - MEM
    name: mem-rss
    query:
      metricName: max:docker.mem.rss
      serviceType: datadog
      type: datadog
    scopeName: default
  - groups:
    - MEM
    name: mem-in-use
    query:
      metricName: max:docker.mem.in_use
      serviceType: datadog
      type: datadog
    scopeName: default
  - groups:
    - CPU
    name: cpu-total
    query:
      metricName: avg:docker.cpu.usage
      serviceType: datadog
      type: datadog
    scopeName: default
  - groups:
    - CPU
    name: cpu-sys
    query:
      metricName: avg:docker.cpu.system
      serviceType: datadog
      type: datadog
    scopeName: default
  - groups:
    - CPU
    name: cpu-user
    query:
      metricName: avg:docker.cpu.user
      serviceType: datadog
      type: datadog
    scopeName: default
  - groups:
    - CPU
    name: cpu-threads
    query:
      metricName: max:docker.thread.count
      serviceType: datadog
      type: datadog
    scopeName: default
  - groups:
    - IO
    name: io-bytes-rcvd
    query:
      metricName: avg:docker.net.bytes_rcvd
      serviceType: datadog
      type: datadog
    scopeName: default
  - groups:
    - IO
    name: io-bytes-sent
    query:
      metricName: avg:docker.net.bytes_rcvd
      serviceType: datadog
      type: datadog
    scopeName: default
  name: democonfig
```

</p>
</details>

### Simple usage with default canary configuration
```shell
kayentactl analysis start --scope=kube_deployment:myappname --canary-config config.yaml
```

### Adding a duration allows you to determine the duration of the experiment 
```shell
kayentactl analysis start --scope=kube_deployment:spud-stories --lifetime-duration=2m --canary-config config.yaml
 ```

### Accessing an analysis result

If you've started an analysis but opted not to wait for it's completion (using the `--no-wait` flag), you can use the
analysis ID to get the analysis at your convenience.

```shell
kayentactl analysis get {ANALYSIS-ID} # add -o json for JSON output instead of the pretty report
```

### Help

```
Usage:
  kayentactl [command]

Available Commands:
  analysis    commands for interacting with canary analysis like starting or retrieving an analysis
  help        Help about any command

Flags:
  -h, --help                 help for kayentactl
  -u, --kayenta-url string   kayenta url (default "http://localhost:8090")
      --no-color             disable output colors
  -v, --verbosity string     log level (debug, info, warn, error, fatal, panic) (default "info")

Use "kayentactl [command] --help" for more information about a command.
```

## TODO

- [] Add documentation about how to install/configure Kayenta OR how to access hosted Kayenta.
- [] Add example/default canary config templates.
- [] Jenkins Plugin
- [] Github Action
- [] Think of more awesome things!


