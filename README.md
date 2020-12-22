# kayenta-jenkins-plugin

## Usage

### Simple usage with default canary configuration
```shell
$ kayentactl analysis start --scope=kube_deployment:myappname
```

### Adding a duration allows you to determine the duration of the experiment 
```shell
 kayentactl analysis start --scope=kube_deployment:spud-stories --lifetime-duration=2m
 ```

- [x] make an app that takes traffic and shows a difference
- [x] configure datadog
- [x] deploy that through argo
- [x] create and deploy that generates traffic
- [x] create new repo for kubernetes infrastructure
- [x] create automation to replace docker tag in deployment object
- [x] push kayentactl container to the cloud on CI! (registry)
- [x] docker container for kayentactl
- [x] update spud stories application to use command line args (isaac)
- [x] get canary config to use what we want it to use for the app above.
- [x] how do we want to communicate failure through the CLI? 
- [x] exit properly with proper error code
- [x] show pretty progress 
- [x] show pretty results
- [x] set rolling update for slower deploy (isaac)
- [] create demo using CLI (run end-to-end to tests)
- [] add CI with github actions
-----------------------------------
- [] jenkins plugin
- [] create demo using Jenkins


## DEMO
* we would explain the problem
  * for the customer - Typically what happens after you deploy is that you rush over to datadog dashboards to see if you notice a difference in metrics. But what metrics are important? how do you know if they have deviated enough to present a real problem? Should you continue moving forward or not? 
  * They want to automate canaries or have a safe deployment
  * There isn't enough intelligence in the system
  * if we simplify the barrier to entry to automate deployment verification, will people do it?
  * is there a way to automate 
  * we think there is a better way with Kayenta which is just the analysis engine from Spinnaker. It simply compares two datasets and we can use that to help inform a developer, operator or automated process on wether to continue to moving forward to scale out.
  * We've created a CLI that wraps the Kayenta API so we can invoke from our laptop, CI tool or workflow engine. This will give us the ability to scale up or scale down when we need to.
  * share the architecture 
  * describe the app we've created, i'll make a breaking change to spud-stories and watch that get deployed.
  * So what we've done is wrap 
  
* take the thing that was newly deployed compare it over 24hrs
* wexplain the proposed

