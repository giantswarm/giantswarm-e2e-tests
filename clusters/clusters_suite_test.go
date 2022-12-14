package clusters_test

import (
	"math/rand"
	"os"
	"testing"
	"time"

	config "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	kubeConfigPath string
	ctrlclient     client.Client
	err            error
)

func TestClusters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clusters Suite")
}

var _ = BeforeEach(func() {
	rand.Seed(time.Now().UnixNano())
	kubeConfigPath = os.Getenv("E2E_KUBECONFIG_PATH")
	if kubeConfigPath == "" {
		Fail("E2E_KUBECONFIG_PATH env var not set")
	}

	// ctrlclient, err = GetManagementClusterK8sClient()
})

func GetManagementClusterK8sClient() (client.Client, error) {
	return client.New(config.GetConfigOrDie(), client.Options{})
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
