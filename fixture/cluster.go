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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/giantswarm-e2e-tests/kubectl"
)

type Cluster struct {
	managementClusterClient ctrl.Client
	workloadClusterClient   ctrl.Client

	workloadClusterName string
	organizationName    string
}

func (f *Cluster) SetUp(kubeConfigPath string) {
	mcClient, err := getManagementClusterK8sClient(kubeConfigPath)
	Expect(err).NotTo(HaveOccurred())

	f.managementClusterClient = mcClient
	f.workloadClusterName = GenerateName("e2e")
	f.organizationName = f.workloadClusterName

	clusterFile, err := os.CreateTemp("", fmt.Sprintf("kubectl-gs-cluster-%s-", f.workloadClusterName))
	Expect(err).NotTo(HaveOccurred())

	orgFile, err := os.CreateTemp("", fmt.Sprintf("kubectl-gs-org-%s-", f.workloadClusterName))
	Expect(err).NotTo(HaveOccurred())

	kubeConfigFlag := fmt.Sprintf("--kubeconfig=%s", kubeConfigPath)
	nameFlag := strings.ToLower(fmt.Sprintf("--name=%s", f.workloadClusterName))

	session := kubectl.GS("template", "organization", "--name", f.workloadClusterName, "--output", orgFile.Name())
	Eventually(session, "15s").Should(gexec.Exit(0))

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

	session = kubectl.GS("template", "cluster", "--provider", "capa", "--organization", f.organizationName, "--description", "e2e test", nameFlag, kubeConfigFlag, "--output", clusterFile.Name())
	Eventually(session, "15s").Should(gexec.Exit(0))

	session = kubectl.Kubectl(kubeConfigFlag, "apply", "-f", clusterFile.Name())
	Eventually(session, "10s").Should(gexec.Exit(0))
}

func (f *Cluster) TearDown() error {
	return nil
}

func (f *Cluster) GetWrokloadClusterKubeClient() ctrl.Client {
	return nil
}

func (f *Cluster) GetManagementClusterKubeClient() ctrl.Client {
	return f.managementClusterClient
}

func (f *Cluster) GetWorkloadClusterName() string {
	return ""
}

func (f *Cluster) GetOrganizationName() string {
	return ""
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

func getManagementClusterK8sClient(kubeConfigPath string) (ctrl.Client, error) {
	configBytes, err := os.ReadFile(kubeConfigPath)
	Expect(err).NotTo(HaveOccurred())

	config, err := clientcmd.RESTConfigFromKubeConfig(configBytes)
	Expect(err).NotTo(HaveOccurred())

	appv1alpha1.AddToScheme(scheme.Scheme)
	return ctrl.New(config, ctrl.Options{Scheme: scheme.Scheme})
}
