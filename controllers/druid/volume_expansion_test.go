package druid

import (
	"time"

	druidv1alpha1 "github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:docs-gen:collapse=Imports

/*
zookeeper_dep_mgmt_test
*/
var _ = Describe("Test volume expansion feature", func() {
	const (
		filePath = "testdata/volume-expansion.yaml"
		timeout  = time.Second * 45
		interval = time.Millisecond * 250
	)

	Context("When checking if volume expansion is enabled", func() {
		It("should error if storageClassName does not exists", func() {
			druid := &druidv1alpha1.Druid{}

			druidCR, err := readDruidClusterSpecFromFile(filePath)
			Expect(err).Should(BeNil())

			By("By setting storage class name to nil")
			druidCR.Spec.Nodes["historicals"].VolumeClaimTemplates[0].Spec.StorageClassName = nil

			Expect(druidCR.Spec.Nodes["historicals"].VolumeClaimTemplates[0].Spec.StorageClassName).Should(BeNil())

			By("By creating a new druidCR")
			Expect(k8sClient.Create(ctx, druidCR)).To(Succeed())

			By("By getting a newly created druidCR")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: druidCR.Name, Namespace: druidCR.Namespace}, druid)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("By getting the historicals nodeSpec")
			allNodeSpecs, err := getAllNodeSpecsInDruidPrescribedOrder(druid)
			Expect(err).Should(BeNil())

			nodeSpec := &druidv1alpha1.DruidNodeSpec{}
			for _, elem := range allNodeSpecs {
				if elem.key == "historicals" {
					nodeSpec = &elem.spec
				}
			}
			Expect(nodeSpec).ShouldNot(BeNil())

			By("By calling the expand volume function with storageClass nil")
			Expect(isVolumeExpansionEnabled(ctx, k8sClient, druid, nodeSpec, nil)).Error()
		})
	})
})
