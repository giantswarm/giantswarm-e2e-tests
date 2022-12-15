package clusters_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	kubeConfigPath string
	k8sClient      client.Client
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

	k8sClient, err = GetManagementClusterK8sClient()
	Expect(err).NotTo(HaveOccurred())
})

func GetManagementClusterK8sClient() (client.Client, error) {
	configBytes, err := os.ReadFile(kubeConfigPath)
	Expect(err).NotTo(HaveOccurred())

	config, err := clientcmd.RESTConfigFromKubeConfig(configBytes)
	Expect(err).NotTo(HaveOccurred())

	appv1alpha1.AddToScheme(scheme.Scheme)
	return client.New(config, client.Options{Scheme: scheme.Scheme})
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GenerateName(prefix string) string {
	sequence := randSeq(10)
	return fmt.Sprintf("%s%s", prefix, sequence)[:9]
}
