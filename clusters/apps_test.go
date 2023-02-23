package clusters_test

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/giantswarm/giantswarm-e2e-tests/matchers"
)

var _ = Describe("Apps", func() {
	It("deploys the default apps", func() {
		ctx := context.Background()
		clusterName := clusterFixture.GetWorkloadClusterName()
		orgNamespace := clusterFixture.GetOrganizationNamespace()
		// We need to wait for default-apps to be deployed before we can check all apps.
		Eventually(GetApp(fmt.Sprintf("%s-%s", clusterName, "default-apps"), orgNamespace), "30m").Should(HaveAppStatus("deployed"))

		managementClusterKubeClient := clusterFixture.GetManagementClusterKubeClient()
		appList := &v1alpha1.AppList{}
		err := managementClusterKubeClient.List(ctx, appList, ctrl.InNamespace(orgNamespace), ctrl.MatchingLabels{"giantswarm.io/managed-by": fmt.Sprintf("%s-%s", clusterName, "default-apps")})
		Expect(err).ShouldNot(HaveOccurred())

		for _, app := range appList.Items {
			Eventually(GetApp(app.Name, app.Namespace), "30m").Should(HaveAppStatus("deployed"))
		}
	})
})
