# giantswarm-e2e-tests

Ginkgo suites for e2e testing the Giantswarm Platform.

## Running the tests

Get the kubeconfig for the management cluster where the workload cluster will be created

```
$ opsctl login grizzly --self-contained mc-kubeconfig.yaml
```

Then run ginkgo

```
E2E_KUBECONFIG_PATH=$PWD/mc-kubeconfig.yaml ginkgo -p --nodes 4 --randomize-all --randomize-suites clusters
```
