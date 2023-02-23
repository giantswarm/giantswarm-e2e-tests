package fixture

import (
	"context"
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

func (f *Cluster) SetUp(kubeConfigPath string, kubectlgsParams ...string) {
	f.managementClusterKubeConfigPath = kubeConfigPath

	mcClient, err := getManagementClusterK8sClient(kubeConfigPath)
	Expect(err).NotTo(HaveOccurred())

	f.managementClusterClient = mcClient
	f.workloadClusterName = generateName("e2e")
	f.organizationName = f.workloadClusterName

	clusterFile, err := os.CreateTemp("", fmt.Sprintf("kubectl-gs-cluster-%s-", f.workloadClusterName))
	Expect(err).NotTo(HaveOccurred())
	defer clusterFile.Close()
	f.workloadClusterManifestsPath = clusterFile.Name()

	orgFile, err := os.CreateTemp("", fmt.Sprintf("kubectl-gs-org-%s-", f.workloadClusterName))
	Expect(err).NotTo(HaveOccurred())
	defer orgFile.Close()
	f.organizationManifestsPath = orgFile.Name()

	nameFlag := strings.ToLower(fmt.Sprintf("--name=%s", f.workloadClusterName))

	session := kubectl.GS("template", "organization", "--name", f.workloadClusterName, "--output", orgFile.Name())
	Eventually(session, "15s").Should(gexec.Exit(0))

	kubeConfigFlag := f.getKubeconfigFlag()
	session = kubectl.Kubectl(kubeConfigFlag, "apply", "-f", orgFile.Name())
	Eventually(session, "10s").Should(gexec.Exit(0))

	Eventually(func() error {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("org-%s", f.organizationName),
			},
		}
		err := mcClient.Get(context.Background(), ctrl.ObjectKeyFromObject(ns), ns)
		return err
	}).Should(Succeed())

	params := []string{"template", "cluster"}
	params = append(params, kubectlgsParams...)
	params = append(params, "--organization", f.organizationName, "--description", "e2e test", nameFlag, kubeConfigFlag, "--output", clusterFile.Name())
	session = kubectl.GS(params...)
	Eventually(session, "15s").Should(gexec.Exit(0))

	session = kubectl.Kubectl(kubeConfigFlag, "apply", "-f", clusterFile.Name())
	Eventually(session, "10s").Should(gexec.Exit(0))
}

func (f *Cluster) TearDown() {
	ctx := context.Background()
	kubeConfigFlag := f.getKubeconfigFlag()
	session := kubectl.Kubectl(kubeConfigFlag, "delete", "-f", f.workloadClusterManifestsPath)
	Eventually(session, "10s").Should(gexec.Exit(0))

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
