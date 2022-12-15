package clusters_test

import (
	"math/rand"
	"os"
	"testing"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/giantswarm/giantswarm-e2e-tests/fixture"
)

var (
	clusterFixture fixture.Cluster
	err            error
)

func TestClusters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clusters Suite")
}

var _ = BeforeSuite(func() {
	rand.Seed(time.Now().UnixNano())
	kubeConfigPath := os.Getenv("E2E_KUBECONFIG_PATH")
	if kubeConfigPath == "" {
		Fail("E2E_KUBECONFIG_PATH env var not set")
	}

	clusterFixture = fixture.Cluster{}
	clusterFixture.SetUp(kubeConfigPath)
})
