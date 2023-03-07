package common

import (
	"github.com/giantswarm/clustertest"
	"github.com/giantswarm/clustertest/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func Run(framework *clustertest.Framework, clusterName *string) {
	var wcClient *client.Client

	BeforeEach(func() {
		var err error

		wcClient, err = framework.WC(*clusterName)
		if err != nil {
			Fail(err.Error())
		}
	})

	It("should be able to connect to MC cluster", func() {
		Expect(framework.MC().CheckConnection()).To(BeNil())
	})

	It("should be able to connect to WC cluster", func() {
		Expect(wcClient.CheckConnection()).To(BeNil())
	})
}
