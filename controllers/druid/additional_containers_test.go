package druid

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("Test Additional Containers", func() {
	Context("When adding cluster-level additional containers", func() {
		It("Should add the containers to the pod", func() {
			filePath := "testdata/additional-containers.yaml"
			druid, err := readDruidClusterSpecFromFile(filePath)
			Expect(err).Should(BeNil())

			err = k8sClient.Create(ctx, druid)
			Expect(err).Should(BeNil())

		})
	})
})
