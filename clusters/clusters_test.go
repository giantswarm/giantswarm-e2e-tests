package clusters_test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Clusters", func() {
	It("creates a cluster", func() {
		clusterName := strings.ToLower(randSeq(8))

		tmpFile, err := os.CreateTemp("", fmt.Sprintf("kubectl-gs-%s-", clusterName))
		Expect(err).NotTo(HaveOccurred())

		kubeConfigFlag := fmt.Sprintf("--kubeconfig=%s", kubeConfigPath)
		nameFlag := strings.ToLower(fmt.Sprintf("--name=t%s", clusterName))
		kubectlGSCommand := exec.Command("kubectl-gs", "template", "cluster", "--provider", "capa", "--organization", "giantswarm", "--description", "e2e test", nameFlag, kubeConfigFlag, "--output", tmpFile.Name())
		session, err := gexec.Start(kubectlGSCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "15s").Should(gexec.Exit(0))

		applyManifestsCommand := exec.Command("kubectl", kubeConfigFlag, "apply", "-f", tmpFile.Name())
		session, err = gexec.Start(applyManifestsCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "10s").Should(gexec.Exit(0))

		// 	Eventually(func() error {
		// 		var err error
		// 		app := app2.NewCR(app2.Config{
		// 			Name:      clusterName,
		// 			Namespace: "org-giantswarm",
		// 		})
		// 		err = ctrlclient.Get(context.Background(), client.ObjectKeyFromObject(app), app)
		// 		Expect(err).NotTo(HaveOccurred())
		//
		// 		if app.Status.Release.Status != "deployed" {
		// 			return errors.New("cluster app is not 'deployed' yet")
		// 		}
		//
		// 		return nil
		// 	}).Should(Succeed())
		// })
	})
