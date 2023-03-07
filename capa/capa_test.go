package capa_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/giantswarm/giantswarm-e2e-tests/common"
)

var _ = Describe("Basic tests", func() {
	common.Run(framework, clusterName)
})
