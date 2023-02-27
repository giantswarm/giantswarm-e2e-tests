# giantswarm-e2e-tests

Ginkgo suites for e2e testing the Giantswarm Platform.

## Running the tests

Get the kubeconfig for the management cluster where the workload cluster will be created / is running

```
$ opsctl login $installation --self-contained mc-kubeconfig.yaml
```

### Creating a new workload cluster

Then run ginkgo

```
E2E_KUBECONFIG_PATH=$PWD/mc-kubeconfig.yaml ginkgo clusters
```

### Using an existing workload cluster

You can run the tests against an already existing workload cluster

```
E2E_WC_KUBECONFIG_PATH="$PWD/e2elyek9m.kubeconfig" E2E_WC_NAME="e2elyek9m" E2E_WC_ORG_NAME="e2elyek9m" E2E_KUBECONFIG_PATH=$PWD/mc-kubeconfig.yaml ginkgo clusters
```

## Logs

You can get logs on what's happening by running ginkgo in verbose mode

```
Running Suite: Clusters Suite - /home/jose/dev/src/github.com/giantswarm/giantswarm-e2e-tests/clusters
======================================================================================================
Random Seed: 1677506100

Will run 1 of 1 specs
------------------------------
[BeforeSuite] 
/home/jose/dev/src/github.com/giantswarm/giantswarm-e2e-tests/clusters/clusters_suite_test.go:28
  2023-02-27T14:55:02+01:00     INFO    Done with setup
[BeforeSuite] PASSED [0.580 seconds]
------------------------------
Apps deploys the default apps
/home/jose/dev/src/github.com/giantswarm/giantswarm-e2e-tests/clusters/apps_test.go:16
  2023-02-27T14:55:03+01:00     INFO    Waiting for default-apps-aws to be marked as 'deployed'.
  2023-02-27T14:55:03+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-aws-ebs-csi-driver"}
  2023-02-27T14:55:03+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-aws-pod-identity-webhook"}
  2023-02-27T14:55:03+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-capi-node-labeler"}
  2023-02-27T14:55:03+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-cert-exporter"}
  2023-02-27T14:55:03+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-cert-manager"}
  2023-02-27T14:55:03+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-cilium"}
  2023-02-27T14:55:03+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-coredns"}
  2023-02-27T14:55:03+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-external-dns"}
  2023-02-27T14:55:03+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-kiam"}
  2023-02-27T14:55:03+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-kube-state-metrics"}
  2023-02-27T14:55:03+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-metrics-server"}
  2023-02-27T14:55:03+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-net-exporter"}
  2023-02-27T14:55:04+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-node-exporter"}
  2023-02-27T14:55:04+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-observability-bundle"}
  2023-02-27T14:55:04+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-vertical-pod-autoscaler"}
  2023-02-27T14:55:04+01:00     INFO    Waiting for app to be marked as 'deployed'.     {"name": "e2elyek9m-vertical-pod-autoscaler-crd"}
• [1.453 seconds]
------------------------------
[AfterSuite] 
/home/jose/dev/src/github.com/giantswarm/giantswarm-e2e-tests/clusters/clusters_suite_test.go:42
[AfterSuite] PASSED [0.000 seconds]
------------------------------

Ran 1 of 1 Specs in 2.033 seconds
SUCCESS! -- 1 Passed | 0 Failed | 0 Pending | 0 Skipped
PASS

Ginkgo ran 1 suite in 4.634009236s
Test Suite Passed
```
