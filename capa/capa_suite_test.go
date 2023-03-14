package capa_test

import (
	"context"
	"path"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/giantswarm/clustertest"
	"github.com/giantswarm/clustertest/pkg/application"
	"github.com/giantswarm/clustertest/pkg/organization"
	"github.com/giantswarm/clustertest/pkg/utils"
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

	cluster = application.NewClusterApp(utils.GenerateRandomName("t"), application.ProviderAWS).
		WithOrg(organization.NewRandomOrg()).
		WithAppVersions("", ""). // If not set, the latest is fetched
		WithAppValuesFile(path.Clean("./test_data/cluster_values.yaml"), path.Clean("./test_data/default-apps_values.yaml"))

	framework.Log("Workload cluster name: %s\n", cluster.Name)

	BeforeSuite(func() {
		applyCtx, cancelApplyCtx := context.WithTimeout(ctx, 20*time.Minute)
		defer cancelApplyCtx()

		client, err := framework.ApplyCluster(applyCtx, cluster)
		Expect(err).To(BeNil())

		nodeCtx, cancelNodeCtx := context.WithTimeout(ctx, 20*time.Minute)
		defer cancelNodeCtx()

		err = client.WaitForControlPlane(nodeCtx, 3)
		Expect(err).To(BeNil())
	})

	AfterSuite(func() {
		err := framework.DeleteCluster(ctx, cluster)
		Expect(err).To(BeNil())
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "CAPA Suite")
}
