package kubectl

import (
	"os/exec"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func GS(args ...string) *gexec.Session {
	command := exec.Command("kubectl-gs", args...)
	session, err := gexec.Start(command, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}
