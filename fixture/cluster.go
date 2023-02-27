package fixture

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	app2 "github.com/giantswarm/app/v6/pkg/app"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/giantswarm-e2e-tests/kubectl"
	. "github.com/giantswarm/giantswarm-e2e-tests/matchers"
)

type Cluster struct {
	managementClusterClient ctrl.Client
	workloadClusterClient   ctrl.Client

	managementClusterKubeConfigPath string

	workloadClusterManifestsPath string
	organizationManifestsPath    string

	workloadClusterName string
	organizationName    string
}

func NewClusterFixture(managementClusterKubeConfigPath string) Cluster {
	kubeConfigPath := os.Getenv("E2E_KUBECONFIG_PATH")
	if kubeConfigPath == "" {
		log.Fatal("E2E_KUBECONFIG_PATH env var not set")
	}

	mcClient, err := getManagementClusterK8sClient(managementClusterKubeConfigPath)
	Expect(err).NotTo(HaveOccurred())

	return Cluster{
		managementClusterClient:         mcClient,
		managementClusterKubeConfigPath: managementClusterKubeConfigPath,
	}
}

func (f *Cluster) SetUp(ctx context.Context, kubectlgsParams ...string) {
	workloadClusterKubeConfigPath := os.Getenv("E2E_WC_KUBECONFIG_PATH")
	if workloadClusterKubeConfigPath != "" {
		f.workloadClusterName = os.Getenv("E2E_WC_NAME")
		f.organizationName = os.Getenv("E2E_WC_ORG_NAME")
		return
	}

	name := generateName("e2e")
	f.createOrganization(ctx, f.GetManagementClusterKubeClient(), name)
	f.createWorkloadCluster(ctx, name, kubectlgsParams)
}

func (f *Cluster) createWorkloadCluster(ctx context.Context, name string, kubectlgsParams []string) {
	f.workloadClusterName = name
	clusterFile, err := os.CreateTemp("", fmt.Sprintf("kubectl-gs-cluster-%s-", name))
	Expect(err).NotTo(HaveOccurred())
	f.workloadClusterManifestsPath = clusterFile.Name()
	defer clusterFile.Close()

	params := []string{"template", "cluster"}
	params = append(params, kubectlgsParams...)
	params = append(params, "--organization", name, "--description", "e2e test", strings.ToLower(fmt.Sprintf("--name=%s", name)), f.getKubeconfigFlag(), "--output", clusterFile.Name())
	session := kubectl.GS(params...)
	Eventually(session, "15s").Should(gexec.Exit(0))

	session = kubectl.Kubectl(f.getKubeconfigFlag(), "apply", "-f", clusterFile.Name())
	Eventually(session, "10s").Should(gexec.Exit(0))

	Eventually(f.GetApp(ctx, name, fmt.Sprintf("org-%s", name)), "10s").Should(HaveAppStatus("deployed"))
}

func (f *Cluster) createOrganization(ctx context.Context, mcClient ctrl.Client, organizationName string) (string, string) {
	f.organizationName = organizationName
	orgFile, err := os.CreateTemp("", fmt.Sprintf("kubectl-gs-org-%s-", organizationName))
	Expect(err).NotTo(HaveOccurred())
	f.organizationManifestsPath = orgFile.Name()
	defer orgFile.Close()

	session := kubectl.GS("template", "organization", "--name", organizationName, "--output", orgFile.Name())
	Eventually(session, "15s").Should(gexec.Exit(0))

	session = kubectl.Kubectl(f.getKubeconfigFlag(), "apply", "-f", orgFile.Name())
	Eventually(session, "15s").Should(gexec.Exit(0))

	Eventually(func() error {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("org-%s", organizationName),
			},
		}
		err := mcClient.Get(ctx, ctrl.ObjectKeyFromObject(ns), ns)
		return err
	}).Should(Succeed())

	return organizationName, orgFile.Name()
}

func (f *Cluster) TearDown(ctx context.Context) {
	workloadClusterKubeConfigPath := os.Getenv("E2E_WC_KUBECONFIG_PATH")
	if workloadClusterKubeConfigPath != "" {
		return
	}

	kubeConfigFlag := f.getKubeconfigFlag()
	session := kubectl.Kubectl(kubeConfigFlag, "delete", "-f", f.workloadClusterManifestsPath)
	Eventually(session, "30s").Should(gexec.Exit(0))

	var getErr error
	Eventually(func() error {
		cluster := &capi.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      f.workloadClusterName,
				Namespace: f.GetOrganizationNamespace(),
			},
		}

		getErr = f.managementClusterClient.Get(ctx, ctrl.ObjectKeyFromObject(cluster), cluster)
		return getErr
	}, "15m").ShouldNot(Succeed())
	Expect(k8serrors.IsNotFound(getErr)).To(BeTrue())

	// The next patch it's a hack that we currently need when deleting clusters, it should go away eventually.
	patch := []byte(`{"metadata":{"finalizers":null}}`)
	err := f.managementClusterClient.Patch(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: f.GetOrganizationNamespace(),
			Name:      fmt.Sprintf("%s-bastion-ignition", f.GetWorkloadClusterName()),
		},
	}, ctrl.RawPatch(types.MergePatchType, patch))
	Expect(err).ToNot(HaveOccurred())

	session = kubectl.Kubectl(kubeConfigFlag, "delete", "-f", f.organizationManifestsPath)
	Eventually(session, "30s").Should(gexec.Exit(0))
}

func (f *Cluster) GetApp(ctx context.Context, name, namespace string) func() *appv1alpha1.App {
	return func() *appv1alpha1.App {
		app := app2.NewCR(app2.Config{
			Name:      name,
			Namespace: namespace,
		})
		err := f.managementClusterClient.Get(ctx, ctrl.ObjectKeyFromObject(app), app)
		if !k8serrors.IsNotFound(err) {
			Expect(err).NotTo(HaveOccurred())
		}

		return app
	}
}

func (f *Cluster) GetWorkloadClusterKubeClient() ctrl.Client {
	return f.workloadClusterClient
}

func (f *Cluster) GetManagementClusterKubeClient() ctrl.Client {
	return f.managementClusterClient
}

func (f *Cluster) GetWorkloadClusterName() string {
	return f.workloadClusterName
}

func (f *Cluster) GetOrganizationName() string {
	return f.organizationName
}

func (f *Cluster) GetOrganizationNamespace() string {
	return fmt.Sprintf("org-%s", f.organizationName)
}

func (f *Cluster) getKubeconfigFlag() string {
	return fmt.Sprintf("--kubeconfig=%s", f.managementClusterKubeConfigPath)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func generateName(prefix string) string {
	sequence := randSeq(10)
	return fmt.Sprintf("%s%s", prefix, sequence)[:9]
}

func getManagementClusterK8sClient(kubeConfigPath string) (ctrl.Client, error) {
	configBytes, err := os.ReadFile(kubeConfigPath)
	Expect(err).NotTo(HaveOccurred())

	config, err := clientcmd.RESTConfigFromKubeConfig(configBytes)
	Expect(err).NotTo(HaveOccurred())

	appv1alpha1.AddToScheme(scheme.Scheme)
	capi.AddToScheme(scheme.Scheme)
	return ctrl.New(config, ctrl.Options{Scheme: scheme.Scheme})
}
