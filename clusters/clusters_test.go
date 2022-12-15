package clusters_test

import (
	"context"
	"errors"
	"fmt"

	app2 "github.com/giantswarm/app/v6/pkg/app"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Clusters", func() {
	It("creates a cluster", func() {
		clusterName := clusterFixture.GetWorkloadClusterName()
		managementClusterClient := clusterFixture.GetManagementClusterKubeClient()

		Eventually(func() error {
			var err error
			app := app2.NewCR(app2.Config{
				Name:      clusterName,
				Namespace: fmt.Sprintf("org-%s", clusterName),
			})
			err = managementClusterClient.Get(context.Background(), client.ObjectKeyFromObject(app), app)
			Expect(err).NotTo(HaveOccurred())

			if app.Status.Release.Status != "deployed" {
				return errors.New("cluster app is not 'deployed' yet")
			}

			return nil
		}, "1m").Should(Succeed())
	})
})
