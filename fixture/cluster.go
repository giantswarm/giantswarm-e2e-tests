package fixture

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	appgs "github.com/giantswarm/app/v6/pkg/app"
	"github.com/go-logr/logr"
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

// SetUp will create an organization and a workload cluster using a random string.
// If `E2E_WC_KUBECONFIG_PATH`, `E2E_WC_NAME` and `E2E_WC_ORG_NAME` are passed, it will not create a workload cluster.
func (f *Cluster) SetUp(ctx context.Context, logger logr.Logger, kubectlgsParams ...string) {
	workloadClusterKubeConfigPath := os.Getenv("E2E_WC_KUBECONFIG_PATH")
	if workloadClusterKubeConfigPath != "" {
		f.workloadClusterName = os.Getenv("E2E_WC_NAME")
		f.organizationName = os.Getenv("E2E_WC_ORG_NAME")
		if f.workloadClusterName == "" || f.organizationName == "" {
			log.Fatal("When passing E2E_WC_KUBECONFIG_PATH env var, you need to also pass E2E_WC_NAME and E2E_WC_ORG_NAME")
		}

		logger.Info("skipping workload cluster creation, using cluster targeted by 'E2E_WC_KUBECONFIG_PATH'")
		wcClient, err := getWorkloadClusterK8sClient(workloadClusterKubeConfigPath)
		Expect(err).NotTo(HaveOccurred())
		f.workloadClusterClient = wcClient

		return
	}

	name := generateName("e2e")
	logger = logger.WithValues("name", name, "organization", name)
	f.createOrganization(ctx, logger, f.GetManagementClusterKubeClient(), name)
	f.createWorkloadCluster(ctx, logger, name, kubectlgsParams)
}

func (f *Cluster) createWorkloadCluster(ctx context.Context, logger logr.Logger, name string, kubectlgsParams []string) {
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

	logger.Info("Waiting for cluster app to be marked as 'deployed'.")
	Eventually(f.GetApp(ctx, name, f.GetOrganizationNamespace()), "10s").Should(HaveAppStatus("deployed"))

	Eventually(func() error {
		logger.Info("Trying to find workload cluster kubeconfig secret in management cluster")
		kubeconfigSecret := &corev1.Secret{}
		return f.managementClusterClient.Get(ctx, ctrl.ObjectKey{Name: fmt.Sprintf("%s-kubeconfig", f.workloadClusterName), Namespace: f.GetOrganizationNamespace()}, kubeconfigSecret)
	}, "5m", "30s").ShouldNot(HaveOccurred())

	kubeconfigSecret := &corev1.Secret{}
	err = f.managementClusterClient.Get(ctx, ctrl.ObjectKey{Name: fmt.Sprintf("%s-kubeconfig", f.workloadClusterName), Namespace: f.GetOrganizationNamespace()}, kubeconfigSecret)
	Expect(err).NotTo(HaveOccurred())

	workloadClusterKubeconfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigSecret.Data["value"])
	Expect(err).NotTo(HaveOccurred())

	Eventually(func() error {
		logger.Info("Trying to create workload cluster controller client")
		f.workloadClusterClient, err = ctrl.New(workloadClusterKubeconfig, ctrl.Options{Scheme: scheme.Scheme})
		if err != nil {
			return err
		}

		return f.managementClusterClient.List(ctx, &corev1.NodeList{})

	}, "10m", "1m").ShouldNot(HaveOccurred())

	logger.Info("Workload cluster has been created")
}

func (f *Cluster) createOrganization(ctx context.Context, logger logr.Logger, mcClient ctrl.Client, organizationName string) (string, string) {
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
				Name: f.GetOrganizationNamespace(),
			},
		}
		err := mcClient.Get(ctx, ctrl.ObjectKeyFromObject(ns), ns)
		return err
	}).Should(Succeed())

	return organizationName, orgFile.Name()
}

// TearDown will delete the workload cluster and the organization created during setup
func (f *Cluster) TearDown(ctx context.Context, logger logr.Logger) {
	workloadClusterKubeConfigPath := os.Getenv("E2E_WC_KUBECONFIG_PATH")
	if workloadClusterKubeConfigPath != "" {
		return
	}

	kubeConfigFlag := f.getKubeconfigFlag()
	session := kubectl.Kubectl(kubeConfigFlag, "delete", "-f", f.workloadClusterManifestsPath)
	Eventually(session, "30s").Should(gexec.Exit(0))

	logger.Info("waiting for workload cluster to be deleted")
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
	}, "20m").ShouldNot(Succeed())
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
		app := appgs.NewCR(appgs.Config{
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

func getWorkloadClusterK8sClient(kubeConfigPath string) (ctrl.Client, error) {
	configBytes, err := os.ReadFile(kubeConfigPath)
	Expect(err).NotTo(HaveOccurred())

	config, err := clientcmd.RESTConfigFromKubeConfig(configBytes)
	Expect(err).NotTo(HaveOccurred())

	return ctrl.New(config, ctrl.Options{Scheme: scheme.Scheme})
}
