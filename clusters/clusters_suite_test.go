package clusters_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var kubeConfigPath string

func TestClusters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clusters Suite")
}

var _ = BeforeEach(func() {
	kubeConfigPath = os.Getenv("E2E_KUBECONFIG_PATH")
	if kubeConfigPath == "" {
		Fail("E2E_KUBECONFIG_PATH env var not set")
	}
})
