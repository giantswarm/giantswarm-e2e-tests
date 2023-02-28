# giantswarm-e2e-tests

Ginkgo suites for e2e testing the Giantswarm Platform.

These tests are organized in folders that describe different testing scenarios that we want to cover.
Each of this scenarios is a ginkgo test suite that uses a common library to interact with a Management Cluster.
This structure is designed so that every test suite creates a workload cluster for the scenario that we want to test and all the tests inside the same suite target that workload cluster. 

There is also a folder for common tests that we want to share across different scenarios. Different suites that want to test different scenarios can invoke these shared tests.

All the code in this repository could keep living in the same repository, but if needed, code could be splitted in different repositories for the cluster fixture, the common tests, etc.

## Running the tests

Get the kubeconfig for the management cluster where the workload cluster will be created / is running

```
$ opsctl login $installation --self-contained mc-kubeconfig.yaml
```

### Creating a new workload cluster

Then run ginkgo

```
E2E_KUBECONFIG_PATH=$PWD/mc-kubeconfig.yaml ginkgo capa-public
```

### Using an existing workload cluster

You can run the tests against an already existing workload cluster

```
E2E_WC_KUBECONFIG_PATH="$PWD/e2elyek9m.kubeconfig" E2E_WC_NAME="e2elyek9m" E2E_WC_ORG_NAME="e2elyek9m" E2E_KUBECONFIG_PATH=$PWD/mc-kubeconfig.yaml ginkgo capa-public
```

## Logs

You can get logs on what's happening by running ginkgo in verbose mode

```
E2E_WC_KUBECONFIG_PATH="$PWD/workloadcluster.kubeconfig" E2E_WC_NAME="my-wc" E2E_WC_ORG_NAME="my-wc" E2E_KUBECONFIG_PATH=$PWD/mc-kubeconfig.yaml ginkgo -v capa-public
Running Suite: Clusters Suite - /home/jose/dev/src/github.com/giantswarm/giantswarm-e2e-tests/capa-public
=========================================================================================================
Random Seed: 1677595864

Will run 3 of 3 specs
------------------------------
[BeforeSuite] 
/home/jose/dev/src/github.com/giantswarm/giantswarm-e2e-tests/capa-public/clusters_suite_test.go:32
  2023-02-28T15:51:07+01:00     INFO    skipping workload cluster creation, using cluster targeted by 'E2E_WC_KUBECONFIG_PATH'
[BeforeSuite] PASSED [0.799 seconds]
------------------------------
CAPA cluster deploys the default apps [capa]
/home/jose/dev/src/github.com/giantswarm/giantswarm-e2e-tests/capa-public/capa_test.go:18
  2023-02-28T15:51:07+01:00     INFO    Waiting for default-apps-aws to be marked as 'deployed'.
  2023-02-28T15:51:08+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-aws-ebs-csi-driver"}
  2023-02-28T15:51:08+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-aws-pod-identity-webhook"}
  2023-02-28T15:51:08+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-capi-node-labeler"}
  2023-02-28T15:51:08+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-cert-exporter"}
  2023-02-28T15:51:08+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-cert-manager"}
  2023-02-28T15:51:08+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-cilium"}
  2023-02-28T15:51:08+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-coredns"}
  2023-02-28T15:51:08+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-external-dns"}
  2023-02-28T15:51:08+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-kiam"}
  2023-02-28T15:51:08+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-kube-state-metrics"}
  2023-02-28T15:51:08+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-metrics-server"}
  2023-02-28T15:51:08+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-net-exporter"}
  2023-02-28T15:51:08+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-node-exporter"}
  2023-02-28T15:51:08+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-observability-bundle"}
  2023-02-28T15:51:09+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-vertical-pod-autoscaler"}
  2023-02-28T15:51:09+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-vertical-pod-autoscaler-crd"}
• [1.444 seconds]
------------------------------
CAPA cluster nodes are all ready [capa]
/home/jose/dev/src/github.com/giantswarm/giantswarm-e2e-tests/common/common.go:20
• [0.055 seconds]
------------------------------
CAPA cluster when using PVCs supports using PVCs [capa]
/home/jose/dev/src/github.com/giantswarm/giantswarm-e2e-tests/common/common.go:57
• [5.344 seconds]
------------------------------
[AfterSuite] 
/home/jose/dev/src/github.com/giantswarm/giantswarm-e2e-tests/capa-public/clusters_suite_test.go:48
[AfterSuite] PASSED [0.000 seconds]
------------------------------

Ran 3 of 3 Specs in 7.643 seconds
SUCCESS! -- 3 Passed | 0 Failed | 0 Pending | 0 Skipped
PASS

Ginkgo ran 1 suite in 10.396325137s
Test Suite Passed
```
