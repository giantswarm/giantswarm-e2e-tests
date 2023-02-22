package clusters_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/giantswarm/giantswarm-e2e-tests/matchers"
)

var _ = Describe("Apps", func() {
	It("deploys the default apps", func() {
		clusterName := clusterFixture.GetWorkloadClusterName()
		orgNamespace := clusterFixture.GetOrganizationNamespace()
		ciliumApp := fmt.Sprintf("%s-cilium", clusterName)

		Eventually(GetApp(clusterName, orgNamespace), "30m").Should(HaveAppStatus("deployed"))
		Eventually(GetApp(ciliumApp, orgNamespace), "30m").Should(HaveAppStatus("deployed"))
	})
})
