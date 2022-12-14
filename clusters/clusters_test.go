package clusters_test

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Clusters", func() {
	It("creates a cluster", func() {
		kubeConfigFlag := fmt.Sprintf("--kubeconfig=%s", kubeConfigPath)
		kubectlGSCommand := exec.Command("kubectl-gs", "template", "cluster", "--provider", "capa", "--organization", "$organization", "--description", "e2e test", "--name", "e2etests", kubeConfigFlag, "--output", "/tmp/cluster.yaml")
		session, err := gexec.Start(kubectlGSCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session.Out, "10s").Should(gbytes.Say("memphis-kiam"))
	})
})
