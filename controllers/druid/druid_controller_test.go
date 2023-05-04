package druid

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"

	druidv1alpha1 "github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
)

// +kubebuilder:docs-gen:collapse=Imports

/*
testDruidOperator
*/
var _ = Describe("Druid Operator", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		filePath = "testdata/druid-smoke-test-cluster.yaml"
		timeout  = time.Second * 45
		interval = time.Millisecond * 250
	)
	druid := &druidv1alpha1.Druid{}

	Context("When testing Druid Operator", func() {
		druidCR, err := readDruidClusterSpecFromFile(filePath)
		Expect(err).Should(BeNil())

		It("should create druidCR - testDruidOperator", func() {
			By("By creating a new druidCR")
			Expect(k8sClient.Create(context.TODO(), druidCR)).To(Succeed())

			// Get CR and match ConfigMaps
			expectedConfigMaps := []string{
				fmt.Sprintf("druid-%s-brokers-config", druidCR.Name),
				fmt.Sprintf("druid-%s-coordinators-config", druidCR.Name),
				fmt.Sprintf("druid-%s-historicals-config", druidCR.Name),
				fmt.Sprintf("druid-%s-routers-config", druidCR.Name),
				fmt.Sprintf("%s-druid-common-config", druidCR.Name),
			}

			By("By getting a newly created druidCR")
			Eventually(func() bool {
				err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: druidCR.Name, Namespace: druidCR.Namespace}, druid)
				if !areStringArraysEqual(druid.Status.ConfigMaps, expectedConfigMaps) {
					return false
				}
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Match ConfigMaps
			By("By matching ConfigMaps")
			Expect(druid.Status.ConfigMaps).Should(ConsistOf(expectedConfigMaps))

			// Match Services
			By("By matching Services")
			expectedServices := []string{
				fmt.Sprintf("druid-%s-brokers", druidCR.Name),
				fmt.Sprintf("druid-%s-coordinators", druidCR.Name),
				fmt.Sprintf("druid-%s-historicals", druidCR.Name),
				fmt.Sprintf("druid-%s-routers", druidCR.Name),
			}
			Expect(druid.Status.Services).Should(ConsistOf(expectedServices))

			// Match StatefulSets
			By("By matching StatefulSets")
			expectedStatefulSets := []string{
				fmt.Sprintf("druid-%s-coordinators", druidCR.Name),
				fmt.Sprintf("druid-%s-historicals", druidCR.Name),
				fmt.Sprintf("druid-%s-routers", druidCR.Name),
			}
			Expect(druid.Status.StatefulSets).Should(ConsistOf(expectedStatefulSets))

			// Match Deployments
			By("By matching Deployments")
			expectedDeployments := []string{
				fmt.Sprintf("druid-%s-brokers", druidCR.Name),
			}
			Expect(druid.Status.Deployments).Should(ConsistOf(expectedDeployments))

			// Match PDBs
			By("By matching PDBs")
			expectedPDBs := []string{
				fmt.Sprintf("druid-%s-brokers", druidCR.Name),
			}
			Expect(druid.Status.PodDisruptionBudgets).Should(ConsistOf(expectedPDBs))

			// Match HPAs
			By("By matching HPAs")
			expectedHPAs := []string{
				fmt.Sprintf("druid-%s-brokers", druidCR.Name),
			}
			Expect(druid.Status.HPAutoScalers).Should(ConsistOf(expectedHPAs))

			// Match Ingress
			By("By matching Ingress")
			expectedIngress := []string{
				fmt.Sprintf("druid-%s-routers", druidCR.Name),
			}
			Expect(druid.Status.Ingress).Should(ConsistOf(expectedIngress))

		})

		It("Test broker deployment", func() {
			createdDeploy := &appsv1.Deployment{}
			brokerDeployment := fmt.Sprintf("druid-%s-brokers", druidCR.Name)
			depNamespacedName := types.NamespacedName{Name: brokerDeployment, Namespace: druidCR.Namespace}

			// Match Deployment replicas
			By("Get deployment and check replicas")

			Eventually(func() bool {
				err := k8sClient.Get(context.TODO(), depNamespacedName, createdDeploy)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(*createdDeploy.Spec.Replicas).To(Equal(druidCR.Spec.Nodes["brokers"].Replicas))

			Eventually(func() bool {
				err := k8sClient.Get(context.TODO(), types.NamespacedName{Name: druidCR.Name, Namespace: druidCR.Namespace}, druid)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Update broker deployment replicas")
			replicaCount := 2
			if druidRep, ok := druid.Spec.Nodes["brokers"]; ok {
				druidRep.Replicas = int32(replicaCount)
				druid.Spec.Nodes["brokers"] = druidRep
			}
			// updating CR
			Expect(k8sClient.Update(context.TODO(), druid)).Should(Succeed())

			// Fetch druid CR and check replicas
			Eventually(func() bool {
				k8sClient.Get(context.TODO(), depNamespacedName, druid)
				return druid.Spec.Nodes["brokers"].Replicas == 2
			}, timeout, interval).Should(BeTrue())

			// Fetch deployment and check replicas
			Eventually(func() bool {
				k8sClient.Get(context.TODO(), depNamespacedName, createdDeploy)
				return *createdDeploy.Spec.Replicas == 2
			}, timeout, interval).Should(BeTrue())
		})

		// Test statefulsets replica count and update the replica count then match
		expectedStatefulSets := []string{"coordinators", "historicals", "routers"}
		for _, names := range expectedStatefulSets {
			names := names
			It(fmt.Sprintf("Statefulset test for %s", names), func() {
				createdSts := &appsv1.StatefulSet{}
				stsName := fmt.Sprintf("druid-%s-%s", druidCR.Name, names)
				stsNamespacedName := types.NamespacedName{Name: stsName, Namespace: druidCR.Namespace}

				// Match statefulset replicas
				By(fmt.Sprintf("Get statefulset and check replicas for %s ", stsName))
				Eventually(func() bool {
					err := k8sClient.Get(context.TODO(), stsNamespacedName, createdSts)
					return err == nil
				}, timeout, interval).Should(BeTrue())
				Expect(*createdSts.Spec.Replicas).To(Equal(druidCR.Spec.Nodes[names].Replicas))

				By(fmt.Sprintf("Update statefulset replicas %s ", stsName))
				replicaCount := 2
				if druidRep, ok := druid.Spec.Nodes[names]; ok {
					druidRep.Replicas = int32(replicaCount)
					druid.Spec.Nodes[names] = druidRep
				}
				// updating CR
				Expect(k8sClient.Update(context.TODO(), druid)).Should(Succeed())

				// Fetch druid CR and check replicas
				Eventually(func() bool {
					k8sClient.Get(context.TODO(), types.NamespacedName{Name: druidCR.Name, Namespace: druidCR.Namespace}, druid)
					return druid.Spec.Nodes[names].Replicas == 2
				}, timeout, interval).Should(BeTrue())

				// Fetch statefulset and check replicas
				Eventually(func() bool {
					k8sClient.Get(context.TODO(), stsNamespacedName, createdSts)
					return *createdSts.Spec.Replicas == 2
				}, timeout, interval).Should(BeTrue())
			})
		}
	})
})

func areStringArraysEqual(a1, a2 []string) bool {
	if len(a1) == len(a2) {
		for i, v := range a1 {
			if v != a2[i] {
				return false
			}
		}
	} else {
		return false
	}
	return true
}
