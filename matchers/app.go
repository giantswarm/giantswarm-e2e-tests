package matchers

import (
	"fmt"

	applicationv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

type haveAppStatus struct {
	expected string
}

func HaveAppStatus(expected string) types.GomegaMatcher {
	return &haveAppStatus{expected: expected}
}

func (m *haveAppStatus) Match(actual interface{}) (bool, error) {
	if actual == nil {
		return false, nil
	}

	actualApp, isApp := actual.(*applicationv1alpha1.App)
	if !isApp {
		return false, fmt.Errorf("%#v is not an App", actual)
	}

	return Equal(actualApp.Status.Release.Status).Match(m.expected)
}

func (m *haveAppStatus) FailureMessage(actual interface{}) (message string) {
	return format.Message(
		actual,
		fmt.Sprintf("to be an App with release status: %s", m.expected),
	)
}

func (m *haveAppStatus) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(
		actual,
		fmt.Sprintf("not to be an App with release status: %s", m.expected),
	)
}
