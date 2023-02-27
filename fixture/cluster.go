package fixture

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
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

func (f *Cluster) SetUp(ctx context.Context, kubeConfigPath string, kubectlgsParams ...string) {
	mcClient, err := getManagementClusterK8sClient(kubeConfigPath)
	Expect(err).NotTo(HaveOccurred())
	f.managementClusterKubeConfigPath = kubeConfigPath
	f.managementClusterClient = mcClient

	name := generateName("e2e")
	f.organizationName, f.organizationManifestsPath = createOrganization(ctx, mcClient, name, f.getKubeconfigFlag())
	f.workloadClusterName, f.workloadClusterManifestsPath = createWorkloadCluster(ctx, mcClient, name, f.getKubeconfigFlag(), kubectlgsParams)
}

func createWorkloadCluster(ctx context.Context, mcClient ctrl.Client, name, kubeconfigFlag string, kubectlgsParams []string) (string, string) {
	clusterFile, err := os.CreateTemp("", fmt.Sprintf("kubectl-gs-cluster-%s-", name))
	Expect(err).NotTo(HaveOccurred())
	defer clusterFile.Close()

	params := []string{"template", "cluster"}
	params = append(params, kubectlgsParams...)
	params = append(params, "--organization", name, "--description", "e2e test", strings.ToLower(fmt.Sprintf("--name=%s", name)), kubeconfigFlag, "--output", clusterFile.Name())
	session := kubectl.GS(params...)
	Eventually(session, "15s").Should(gexec.Exit(0))

	session = kubectl.Kubectl(kubeconfigFlag, "apply", "-f", clusterFile.Name())
	Eventually(session, "10s").Should(gexec.Exit(0))

	Eventually(func() error {
		clusterApp := &appv1alpha1.App{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: fmt.Sprintf("org-%s", name),
			},
		}
		err := mcClient.Get(ctx, ctrl.ObjectKeyFromObject(clusterApp), clusterApp)
		if err != nil {
			return err
		}

		if clusterApp.Status.Release.Status != "deployed" {
			return errors.New("cluster app is not 'deployed' yet")
		}

		return nil
	}, "10s").Should(Succeed())

	return name, clusterFile.Name()
}

func createOrganization(ctx context.Context, mcClient ctrl.Client, organizationName, kubeconfigFlag string) (string, string) {
	orgFile, err := os.CreateTemp("", fmt.Sprintf("kubectl-gs-org-%s-", organizationName))
	Expect(err).NotTo(HaveOccurred())
	defer orgFile.Close()

	session := kubectl.GS("template", "organization", "--name", organizationName, "--output", orgFile.Name())
	Eventually(session, "15s").Should(gexec.Exit(0))

	session = kubectl.Kubectl(kubeconfigFlag, "apply", "-f", orgFile.Name())
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
	kubeConfigFlag := f.getKubeconfigFlag()
	session := kubectl.Kubectl(kubeConfigFlag, "delete", "-f", f.workloadClusterManifestsPath)
	Eventually(session, "15s").Should(gexec.Exit(0))

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
	Eventually(session, "10s").Should(gexec.Exit(0))
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
