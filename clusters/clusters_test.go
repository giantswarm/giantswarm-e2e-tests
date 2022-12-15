package clusters_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	app2 "github.com/giantswarm/app/v6/pkg/app"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/giantswarm-e2e-tests/kubectl"
)

var _ = Describe("Clusters", func() {
	It("creates a cluster", func() {
		clusterName := GenerateName("e2e")

		clusterFile, err := os.CreateTemp("", fmt.Sprintf("kubectl-gs-cluster-%s-", clusterName))
		Expect(err).NotTo(HaveOccurred())

		orgFile, err := os.CreateTemp("", fmt.Sprintf("kubectl-gs-org-%s-", clusterName))
		Expect(err).NotTo(HaveOccurred())

		kubeConfigFlag := fmt.Sprintf("--kubeconfig=%s", kubeConfigPath)
		nameFlag := strings.ToLower(fmt.Sprintf("--name=%s", clusterName))

		session := kubectl.GS("template", "organization", "--name", clusterName, "--output", orgFile.Name())
		Eventually(session, "15s").Should(gexec.Exit(0))

		session = kubectl.Kubectl(kubeConfigFlag, "apply", "-f", orgFile.Name())
		Eventually(session, "10s").Should(gexec.Exit(0))

		Eventually(func() error {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("org-%s", clusterName),
				},
			}
			err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(ns), ns)
			return err
		}).Should(Succeed())

		session = kubectl.GS("template", "cluster", "--provider", "capa", "--organization", clusterName, "--description", "e2e test", nameFlag, kubeConfigFlag, "--output", clusterFile.Name())
		Eventually(session, "15s").Should(gexec.Exit(0))

		session = kubectl.Kubectl(kubeConfigFlag, "apply", "-f", clusterFile.Name())
		Eventually(session, "10s").Should(gexec.Exit(0))

		Eventually(func() error {
			var err error
			app := app2.NewCR(app2.Config{
				Name:      clusterName,
				Namespace: fmt.Sprintf("org-%s", clusterName),
			})
			err = k8sClient.Get(context.Background(), client.ObjectKeyFromObject(app), app)
			Expect(err).NotTo(HaveOccurred())

			if app.Status.Release.Status != "deployed" {
				return errors.New("cluster app is not 'deployed' yet")
			}

			return nil
		}, "1m").Should(Succeed())
	})
})
