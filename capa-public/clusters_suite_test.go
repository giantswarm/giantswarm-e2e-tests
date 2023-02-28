package capa_public_test

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
	ctx            context.Context
	logger         logr.Logger
)

var promise = &fixture.Promise{}

func TestClusters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CAPA Clusters with default values suite")
}

var _ = BeforeSuite(func() {
	ctx = context.Background()
	opts := zap.Options{
		DestWriter:  GinkgoWriter,
		Development: true,
		TimeEncoder: zapcore.RFC3339TimeEncoder,
	}
	logger = zap.New(zap.UseFlagOptions(&opts))

	clusterFixture = fixture.NewClusterFixture(os.Getenv("E2E_KUBECONFIG_PATH"))
	clusterFixture.SetUp(ctx, logger, "--provider", "capa")

	// We need to set the cluster fixture into our promise so that tests can use it
	promise.Cluster = &clusterFixture
})

var _ = AfterSuite(func() {
	clusterFixture.TearDown(ctx, logger)
	promise.Cluster = nil
})
