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
)

var (
	framework   *clustertest.Framework
	clusterName *string
	namespace   *string
)

func TestCAPA(t *testing.T) {
	var err error
	ctx := context.Background()

	framework, err = clustertest.New()
	if err != nil {
		panic(err)
	}

	clusterName = clustertest.StringToPointer(clustertest.GenerateRandomName("t"))
	framework.Log("Workload cluster name: %s\n", *clusterName)
	namespace = clustertest.StringToPointer("org-giantswarm")

	BeforeSuite(func() {
		timeoutCtx, cancelTimeout := context.WithTimeout(ctx, 20*time.Minute)
		defer cancelTimeout()

		_, err := framework.ApplyCluster(
			timeoutCtx,
			application.NewClusterApp(*clusterName, application.ProviderAWS).
				WithNamespace(*namespace).
				WithAppVersions("", ""). // If not set, the latest is fetched
				WithAppValuesFile(path.Clean("./test_data/cluster_values.yaml"), path.Clean("./test_data/default-apps_values.yaml")),
		)
		Expect(err).To(BeNil())
	})

	AfterSuite(func() {
		err := framework.DeleteCluster(ctx, *clusterName, *namespace)
		Expect(err).To(BeNil())
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "CAPA Suite")
}
