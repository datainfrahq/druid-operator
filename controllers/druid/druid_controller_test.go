/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// +kubebuilder:docs-gen:collapse=Apache License

/*
As usual, we start with the necessary imports. We also define some utility variables.
*/
package druid

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

	Context("When testing Druid Operator", func() {
		// Read and marshal CR file
		bytes, err := ioutil.ReadFile(filePath)
		Expect(err).ToNot(HaveOccurred(), "Failed to read druid cluster spec")

		druidCR := new(druidv1alpha1.Druid)
		err = yaml.Unmarshal(bytes, &druidCR)
		Expect(err).ToNot(HaveOccurred(), "Failed to unmarshall druid cluster spec")

		It("should create druidCR - testDruidOperator", func() {
			By("By creating a new druidCR")
			Expect(k8sClient.Create(context.TODO(), druidCR)).To(Succeed())

			// Get CR and match ConfigMaps
			druid := &druidv1alpha1.Druid{}

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
