# monban 門番

Monban(門番) is simple managing locked deployments by Admission Webhooks in a namespace.

[![Build Status](https://travis-ci.org/koudaiii/monban.svg?branch=master)](https://travis-ci.org/koudaiii/monban)
[![Docker Repository on Quay](https://quay.io/repository/koudaiii/monban/status "Docker Repository on Quay")](https://quay.io/repository/koudaiii/monban)
[![GitHub release](https://img.shields.io/github/release/koudaiii/monban.svg)](https://github.com/koudaiii/monban/releases)

## Description

When you need to lock deployments. Monban(門番) can lock deployments in a namespace. Monban(門番) is valid at the time of the following situations.

- for Maintenance
- for Code-freeze
- for recovery operations in Production

Please refer to [Admission Webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#admission-webhooks) and the implementation of [the admission webhook server](https://github.com/kubernetes/kubernetes/tree/37555e6d24c2f951c40660ea59a80fa251982005/test/images/webhook).

## Table of Contents

- [monban 門番](#monban-%E9%96%80%E7%95%AA)
  - [Description](#description)
  - [Table of Contents](#table-of-contents)
  - [Requirements](#requirements)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Example: How to lock deployments in a namespace](#example-how-to-lock-deployments-in-a-namespace)
    - [Example: How to unlock deployment in a namespace](#example-how-to-unlock-deployment-in-a-namespace)
  - [in minikube](#in-minikube)
  - [Contribution](#contribution)
  - [Author](#author)
  - [License](#license)

## Requirements

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [cfssl](https://coreos.com/os/docs/latest/generate-self-signed-certificates.html)

```console
# for Mac
$ brew install cfssl # for make cert files
$ brew install kubernetes-cli # for deploy to kubernetes
```

## Installation

1. Setup RBAC (ex. https://docs.bitnami.com/kubernetes/how-to/configure-rbac-in-your-kubernetes-cluster/)
2. Monban(門番) deploy to k8s.

```console
$ make deploy
```

Check deploy

```console
$ kubectl get deployment monban -n default
NAME     DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
monban   1         1         1            1           21h

$ kubectl logs deployment/monban -f -n default
2018/12/19 05:26:03 Starting monban...
```

## Usage

### Example: How to lock deployments in a namespace

Monban enabled

```console
$ kubectl annotate namespace/default koudaiii/monban=enabled
namespace/default annotated
```

Check lock

```console
$ kubectl patch deployment/nginx-test -p "{\"spec\":{\"template\":{\"metadata\":{\"labels\":{\"date\":\"`date +'%s'`\"}}}}}"
Error from server: admission webhook "monban.default.service" denied the request: nginx-test is locked in default.
If you unlock, Please run command `kubectl annotate namespace/default koudaiii/monban-`
```

locking deployment :ok_hand:

### Example: How to unlock deployment in a namespace

Monban disabled

```console
$ kubectl annotate namespace/default koudaiii/monban-
namespace/default annotated
```

Check unlock

```console
$ kubectl patch deployment/nginx-test -p "{\"spec\":{\"template\":{\"metadata\":{\"labels\":{\"date\":\"`date +'%s'`\"}}}}}"
deployment.extensions/nginx-test patched

$ kubectl get po
nginx-test-56f766d96f-7qd4x            1/1     Running             0          2d
nginx-test-56f766d96f-8tgfn            1/1     Running             0          2d
nginx-test-56f766d96f-bwltr            0/1     Terminating         0          2d
nginx-test-56f766d96f-cfrpd            1/1     Running             0          2d
nginx-test-56f766d96f-k55jn            1/1     Running             0          2d
nginx-test-56f766d96f-rzd2j            1/1     Running             0          2d
nginx-test-56f766d96f-vvlb8            1/1     Running             0          2d
nginx-test-8595c7fdbd-642bn            1/1     Running             0          10s
nginx-test-8595c7fdbd-7m72g            1/1     Running             0          10s
nginx-test-8595c7fdbd-dgtqn            0/1     ContainerCreating   0          4s
nginx-test-8595c7fdbd-h6rqg            1/1     Running             0          6s
nginx-test-8595c7fdbd-hfml7            0/1     ContainerCreating   0          1s
```

unlocked deployment :ok_hand:

## in minikube

1. Setup [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/).
2. Clone this repository and build using `make`.

```shell-session
$ minikube start
$ minikube update-context
$ make deploy
```

3. Create User RBAC

Set context

```console
$ kubectl config set-credentials koudaiii --client-certificate=$HOME/.minikube/client.crt --client-key=$HOME/.minikube/client.key
$ kubectl config set-context koudaiii-context --cluster=minikube --namespace=default --user=koudaiii
# Check
$ kubectl --context=koudaiii-context get pods
```

Set RBAC

```console
$ kubectl apply -f example/user.yaml
```

4. Deploy sample app

```console
$ kubectl --context=koudaiii-context run --image nginx nginx-test
$ kubectl --context=koudaiii-context get pods
```

5. Check Monban(門番)

Reload

```console
$ kubectl --context=koudaiii-context patch deployment nginx-test -p "{\"spec\":{\"template\":{\"metadata\":{\"labels\":{\"date\":\"`date +'%s'`\"}}}}}"
deployment.extensions/nginx-test patched

$ kubectl get po
NAME                          READY   STATUS              RESTARTS   AGE
monban-84647c5bbc-p4ntj       1/1     Running             0          12m
nginx-test-5cb5969668-2j5qn   1/1     Running             0          1m
nginx-test-7499b7747-mvdf7    0/1     ContainerCreating   0          3s
```

Monban enabled

```console
$ kubectl --context=koudaiii-context annotate namespace/default koudaiii/monban=enabled
namespace/default annotated

$ kubectl --context=koudaiii-context patch deployment nginx-test -p "{\"spec\":{\"template\":{\"metadata\":{\"labels\":{\"date\":\"`date +'%s'`\"}}}}}"
Error from server: admission webhook "monban.default.service" denied the request: nginx-test is locked in default.
If you unlock, Please run command `kubectl annotate namespace/default koudaiii/monban-`
```

Monban disable

```console
$ kubectl annotate namespace/default koudaiii/monban-
namespace/default annotated

$ kubectl --context=koudaiii-context patch deployment nginx-test -p "{\"spec\":{\"template\":{\"metadata\":{\"labels\":{\"date\":\"`date +'%s'`\"}}}}}"
deployment.extensions/nginx-test patched
```

## Contribution

1. Fork ([https://github.com/koudaiii/monban/fork](https://github.com/koudaiii/monban/fork))
2. Create a feature branch
3. Commit your changes
4. Rebase your local changes against the master branch
5. Run test suite with the `go test ./...` command and confirm that it passes
6. Run `gofmt -s`
7. Create a new Pull Request

## Author

[koudaiii](https://github.com/koudaiii)

## License

[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)
