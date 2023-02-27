package clusters_test

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"github.com/giantswarm/giantswarm-e2e-tests/fixture"
)

var (
	clusterFixture fixture.Cluster
)

func TestClusters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clusters Suite")
}

var _ = BeforeSuite(func() {
	ctx := context.Background()
	clusterFixture = fixture.NewClusterFixture(os.Getenv("E2E_KUBECONFIG_PATH"))
	clusterFixture.SetUp(ctx, "--provider", "capa")
})

var _ = AfterSuite(func() {
	ctx := context.Background()
	clusterFixture.TearDown(ctx)
})
