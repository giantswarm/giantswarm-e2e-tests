package clusters_test

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	applicationv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	app2 "github.com/giantswarm/app/v6/pkg/app"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

var _ = AfterSuite(func() {
	clusterFixture.TearDown()
})

func GetApp(name, namespace string) func() *applicationv1alpha1.App {
	managementClusterClient := clusterFixture.GetManagementClusterKubeClient()
	return func() *applicationv1alpha1.App {
		app := app2.NewCR(app2.Config{
			Name:      name,
			Namespace: namespace,
		})
		err := managementClusterClient.Get(context.Background(), client.ObjectKeyFromObject(app), app)
		if !k8serrors.IsNotFound(err) {
			Expect(err).NotTo(HaveOccurred())
		}

		return app
	}
}
