package druid

import (
	"fmt"
	"time"

	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("Test Additional Containers", func() {
	const (
		timeout  = time.Second * 45
		interval = time.Millisecond * 250
	)

	Context("When adding cluster-level additional containers", func() {
		It("Should add the containers to the pod", func() {
			By("By creating a Druid object")
			filePath := "testdata/additional-containers.yaml"
			druid, err := readDruidClusterSpecFromFile(filePath)
			Expect(err).Should(BeNil())

			Expect(k8sClient.Create(ctx, druid)).To(Succeed())

			existDruid := &v1alpha1.Druid{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: druid.Name, Namespace: druid.Namespace}, existDruid)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			brokerDeployment := &v1.Deployment{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Namespace: druid.Namespace,
					Name:      fmt.Sprintf("druid-%s-%s", druid.Name, "brokers"),
				}, brokerDeployment)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(brokerDeployment.Spec.Template.Spec.Containers).ShouldNot(BeNil())

			isClusterContainerExists := false
			isNodeContainerExists := false
			for _, container := range brokerDeployment.Spec.Template.Spec.Containers {
				if container.Name == "cluster-level" {
					isClusterContainerExists = true
					continue
				}
				if container.Name == "node-level" {
					isNodeContainerExists = true
					continue
				}
			}

			Expect(isClusterContainerExists).Should(BeTrue())
			Expect(isNodeContainerExists).Should(BeTrue())
		})
	})
})
