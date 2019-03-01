[![Go Report Card](https://goreportcard.com/badge/github.com/piontec/netperf-operator)](https://goreportcard.com/report/github.com/piontec/netperf-operator)
[![Build status](https://travis-ci.com/piontec/netperf-operator.svg?branch=master)](https://travis-ci.com/piontec/netperf-operator.svg?branch=master)
# Netperf operator
This is a simple Kubernetes Operator, that uses the legendary [netperf tool](https://hewlettpackard.github.io/netperf/) to run 2 pods in your cluster and measure real TCP connection throughput between them.
This project is an example kubernetes operator based on [Operator Framework SDK](https://github.com/operator-framework/operator-sdk). I created it for two reasons: first one, to learn how to build an operator using the SDK, second to solve a network testing problem. 
This README.md file is about how to develop, build and run the project. If you wanto to learn more about how it works and how to write Kubernetes operators/controllers, head to my [blog](https://www.tailored.cloud/kubernetes/write-a-kubernetes-controller-operator-sdk/).

## Installing
*Note: for installation for development, check [Developers guide](#dev-guide)*

You need to deploy the controller, its Custom Resource Definition and RBAC resources:
```bash
kubectl create -f deploy/crd.yaml
kubectl create -f deploy/rbac.yaml
kubectl create -f deploy/operator.yaml
```

## Users guide
The controller runs tests only in a single namespace, in which the controller is deployed.
In this namespace, you have to create the following resource:
```yaml
apiVersion: "app.example.com/v1alpha1"
kind: "Netperf"
metadata:
  name: "example"
spec:
  serverNode: "minikube"
  clientNode: "minikube"
```
Wait for the Netperf object to complete (`status: Done`) and check the measured throughput.

If you skip any of the `serverNode` or `clientNode` in `spec:`, they will be normally chosen and assigned by kube's scheduler. If you configure them, node affinity will be used to run on the specific node.

## <a name="dev-guide"></a> Developers guide
There are 2 ways you can build and run the operator:
* for rapid development and testing: run the operator process [outside of cluster](#dev-outside), on your development machine, with `kubectl` configured to access your cluster
* for deploying to cluster and testing full deployment and in cluster operation: you need to build a container image with the operator and deploy it to the cluster

### Common tasks
In both cases, you need to start with following steps:
* you need to have a functional golang build environment
* you need to have `dep` [installed](https://github.com/golang/dep)
* clone this repo
```bash
mkdir $GOPATH/src/github.com/piontec
cd $GOPATH/src/github.com/piontec
git clone https://github.com/piontec/netperf-operator.git
cd netperf-operator
```
* install dependencies
```bash
dep ensure
```
* create required Custom Resource Definition in the cluster
```bash
kubectl create -f deploy/crd.yaml
```

### <a name="dev-outside"></a> Running outside of cluster
To build the plugin on your machine, you have to build it with
```bash
go build cmd/netperf-operator/main.go
```

To run it, you have to set the following 2 environment variables that point to your `.kube/config` file and define the name of a Namespace to monitor:
```bash
export KUBERNETES_CONFIG=/your/path/.kube/config
export WATCH_NAMESPACE=default
```

Still, if you're using VS Code with go plugin (which I highly recommend), I incuded in the repo my `launch.json` file. Skip the manual steps above, open the project, set your environment variables in the `launch.json` file and hit `F5` in VS Code - you're ready to go.

### <a name="dev-incluster"></a> Running inside a cluster
The simplest way is through [installing the operator sdk](https://github.com/operator-framework/operator-sdk#quick-start). You also need an image repository, where you can store the ready image:
```bash
export IMAGE=my.repo.url/netperf-operator:v0.0.1
```
Then, build it and push to registry:
```bash
operator-sdk build $IMAGE
docker push $IMAGE
```
As last step, deploy the RBAC definition for the controller and a Deployment that will run it:
```bash
kubectl create -f deploy/rbac.yaml
kubectl create -f deploy/operator.yaml
```
