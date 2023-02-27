package clusters_test

import (
	"context"
	"os"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/giantswarm/giantswarm-e2e-tests/fixture"
)

var (
	clusterFixture fixture.Cluster
	logger         logr.Logger
)

func TestClusters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clusters Suite")
}

var _ = BeforeSuite(func() {
	ctx := context.Background()
	opts := zap.Options{
		DestWriter:  GinkgoWriter,
		Development: true,
		TimeEncoder: zapcore.RFC3339TimeEncoder,
	}
	logger = zap.New(zap.UseFlagOptions(&opts))

	clusterFixture = fixture.NewClusterFixture(os.Getenv("E2E_KUBECONFIG_PATH"))
	clusterFixture.SetUp(ctx, "--provider", "capa")
})

var _ = AfterSuite(func() {
	ctx := context.Background()
	clusterFixture.TearDown(ctx)
})
