package capa_test

import (
	"context"
	"path"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cr "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/clustertest"
	"github.com/giantswarm/clustertest/pkg/application"
	"github.com/giantswarm/clustertest/pkg/logger"
	"github.com/giantswarm/clustertest/pkg/organization"
	"github.com/giantswarm/clustertest/pkg/utils"
	"github.com/giantswarm/clustertest/pkg/wait"
)

var (
	framework *clustertest.Framework
	cluster   *application.Cluster
)

func TestCAPA(t *testing.T) {
	var err error
	ctx := context.Background()

	framework, err = clustertest.New()
	if err != nil {
		panic(err)
	}
	logger.LogWriter = GinkgoWriter

	cluster = application.NewClusterApp(utils.GenerateRandomName("t"), application.ProviderAWS).
		WithOrg(organization.NewRandomOrg()).
		WithAppVersions("", ""). // If not set, the latest is fetched
		WithAppValuesFile(path.Clean("./test_data/cluster_values.yaml"), path.Clean("./test_data/default-apps_values.yaml"))

	logger.Log("Workload cluster name: %s", cluster.Name)

	BeforeSuite(func() {
		applyCtx, cancelApplyCtx := context.WithTimeout(ctx, 20*time.Minute)
		defer cancelApplyCtx()

		client, err := framework.ApplyCluster(applyCtx, cluster)
		Expect(err).To(BeNil())

		Eventually(
			wait.IsNumNodesReady(ctx, client, 3, &cr.MatchingLabels{"node-role.kubernetes.io/control-plane": ""}),
			20*time.Minute,
			30*time.Second,
		).Should(BeTrue())
	})

	AfterSuite(func() {
		err := framework.DeleteCluster(ctx, cluster)
		Expect(err).To(BeNil())
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "CAPA Suite")
}
