package capa_public_test

import (
	"fmt"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/giantswarm-e2e-tests/common"
	. "github.com/giantswarm/giantswarm-e2e-tests/matchers"
)

var _ = Describe("CAPA cluster", Label("capa"), func() {
	common.CommonTests(promise)

	It("deploys the default apps", func() {
		clusterName := clusterFixture.GetWorkloadClusterName()
		orgNamespace := clusterFixture.GetOrganizationNamespace()
		defaultAppsAppName := fmt.Sprintf("%s-%s", clusterName, "default-apps")
		// We need to wait for default-apps to be deployed before we can check all apps.
		logger.Info("Waiting for default-apps-aws to be marked as 'deployed'")
		Eventually(clusterFixture.GetApp(ctx, defaultAppsAppName, orgNamespace), "30s").Should(HaveAppStatus("deployed"))

		managementClusterKubeClient := clusterFixture.GetManagementClusterKubeClient()
		appList := &v1alpha1.AppList{}
		err := managementClusterKubeClient.List(ctx, appList, ctrl.InNamespace(orgNamespace), ctrl.MatchingLabels{"giantswarm.io/managed-by": defaultAppsAppName})
		Expect(err).ShouldNot(HaveOccurred())

		for _, app := range appList.Items {
			logger.Info("Waiting for app to be marked as 'deployed'", "name", app.Name)
			Eventually(clusterFixture.GetApp(ctx, app.Name, app.Namespace), "30m", "1m").Should(HaveAppStatus("deployed"))
		}
	})
})
