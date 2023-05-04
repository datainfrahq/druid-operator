package druid

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"

	druidv1alpha1 "github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
)

// +kubebuilder:docs-gen:collapse=Imports

// testHandler
var _ = Describe("Test handler", func() {
	Context("When testing handler", func() {
		It("should make statefulset for broker", func() {
			By("By making statefulset for broker")
			filePath := "testdata/druid-test-cr.yaml"
			clusterSpec, err := readDruidClusterSpecFromFile(filePath)
			Expect(err).Should(BeNil())

			nodeSpecUniqueStr := makeNodeSpecificUniqueString(clusterSpec, "brokers")
			nodeSpec := clusterSpec.Spec.Nodes["brokers"]

			actual, _ := makeStatefulSet(&nodeSpec, clusterSpec, makeLabelsForNodeSpec(&nodeSpec, clusterSpec, clusterSpec.Name, nodeSpecUniqueStr), nodeSpecUniqueStr, "blah", nodeSpecUniqueStr)
			addHashToObject(actual)

			expected := new(appsv1.StatefulSet)
			err = readAndUnmarshallResource("testdata/broker-statefulset.yaml", &expected)
			Expect(err).Should(BeNil())

			Expect(actual).Should(Equal(expected))
		})

		It("should make statefulset for broker with sidecar", func() {
			By("By making statefulset for broker with sidecar")
			filePath := "testdata/druid-test-cr-sidecar.yaml"
			clusterSpec, err := readDruidClusterSpecFromFile(filePath)
			Expect(err).Should(BeNil())

			nodeSpecUniqueStr := makeNodeSpecificUniqueString(clusterSpec, "brokers")
			nodeSpec := clusterSpec.Spec.Nodes["brokers"]

			actual, _ := makeStatefulSet(&nodeSpec, clusterSpec, makeLabelsForNodeSpec(&nodeSpec, clusterSpec, clusterSpec.Name, nodeSpecUniqueStr), nodeSpecUniqueStr, "blah", nodeSpecUniqueStr)
			addHashToObject(actual)

			expected := new(appsv1.StatefulSet)
			readAndUnmarshallResource("testdata/broker-statefulset-sidecar.yaml", &expected)
			Expect(err).Should(BeNil())

			Expect(actual).Should(Equal(expected))
		})

		It("should make deployment for broker", func() {
			By("By making deployment for broker")
			filePath := "testdata/druid-test-cr.yaml"
			clusterSpec, err := readDruidClusterSpecFromFile(filePath)
			Expect(err).Should(BeNil())

			nodeSpecUniqueStr := makeNodeSpecificUniqueString(clusterSpec, "brokers")
			nodeSpec := clusterSpec.Spec.Nodes["brokers"]

			actual, _ := makeDeployment(&nodeSpec, clusterSpec, makeLabelsForNodeSpec(&nodeSpec, clusterSpec, clusterSpec.Name, nodeSpecUniqueStr), nodeSpecUniqueStr, "blah", nodeSpecUniqueStr)
			addHashToObject(actual)

			expected := new(appsv1.Deployment)
			readAndUnmarshallResource("testdata/broker-deployment.yaml", &expected)
			Expect(err).Should(BeNil())

			Expect(actual).Should(Equal(expected))
		})

		It("should make PDB for broker", func() {
			By("By making PDB for broker")
			filePath := "testdata/druid-test-cr.yaml"
			clusterSpec, err := readDruidClusterSpecFromFile(filePath)
			Expect(err).Should(BeNil())

			nodeSpecUniqueStr := makeNodeSpecificUniqueString(clusterSpec, "brokers")
			nodeSpec := clusterSpec.Spec.Nodes["brokers"]

			actual, _ := makePodDisruptionBudget(&nodeSpec, clusterSpec, makeLabelsForNodeSpec(&nodeSpec, clusterSpec, clusterSpec.Name, nodeSpecUniqueStr), nodeSpecUniqueStr)
			addHashToObject(actual)

			expected := new(policyv1.PodDisruptionBudget)
			readAndUnmarshallResource("testdata/broker-pod-disruption-budget.yaml", &expected)
			Expect(err).Should(BeNil())

			Expect(actual).Should(Equal(expected))
		})

		It("should make headless service", func() {
			By("By making headless service")
			filePath := "testdata/druid-test-cr.yaml"
			clusterSpec, err := readDruidClusterSpecFromFile(filePath)
			Expect(err).Should(BeNil())

			nodeSpecUniqueStr := makeNodeSpecificUniqueString(clusterSpec, "brokers")
			nodeSpec := clusterSpec.Spec.Nodes["brokers"]

			actual, _ := makeService(&nodeSpec.Services[0], &nodeSpec, clusterSpec, makeLabelsForNodeSpec(&nodeSpec, clusterSpec, clusterSpec.Name, nodeSpecUniqueStr), nodeSpecUniqueStr)
			addHashToObject(actual)

			expected := new(corev1.Service)
			readAndUnmarshallResource("testdata/broker-headless-service.yaml", &expected)
			Expect(err).Should(BeNil())

			Expect(actual).Should(Equal(expected))
		})

		It("should make load balancer service", func() {
			By("By making load balancer service")
			filePath := "testdata/druid-test-cr.yaml"
			clusterSpec, err := readDruidClusterSpecFromFile(filePath)
			Expect(err).Should(BeNil())

			nodeSpecUniqueStr := makeNodeSpecificUniqueString(clusterSpec, "brokers")
			nodeSpec := clusterSpec.Spec.Nodes["brokers"]

			actual, _ := makeService(&nodeSpec.Services[1], &nodeSpec, clusterSpec, makeLabelsForNodeSpec(&nodeSpec, clusterSpec, clusterSpec.Name, nodeSpecUniqueStr), nodeSpecUniqueStr)
			addHashToObject(actual)

			expected := new(corev1.Service)
			readAndUnmarshallResource("testdata/broker-load-balancer-service.yaml", &expected)
			Expect(err).Should(BeNil())

			Expect(actual).Should(Equal(expected))
		})

		It("should make config map", func() {
			By("By making config map")
			filePath := "testdata/druid-test-cr.yaml"
			clusterSpec, err := readDruidClusterSpecFromFile(filePath)
			Expect(err).Should(BeNil())

			actual, _ := makeCommonConfigMap(clusterSpec, makeLabelsForDruid(clusterSpec.Name))
			addHashToObject(actual)

			expected := new(corev1.ConfigMap)
			readAndUnmarshallResource("testdata/common-config-map.yaml", &expected)
			Expect(err).Should(BeNil())

			Expect(actual).Should(Equal(expected))
		})

		It("should make broker config map", func() {
			By("By making broker config map")
			filePath := "testdata/druid-test-cr.yaml"
			clusterSpec, err := readDruidClusterSpecFromFile(filePath)
			Expect(err).Should(BeNil())

			nodeSpecUniqueStr := makeNodeSpecificUniqueString(clusterSpec, "brokers")
			nodeSpec := clusterSpec.Spec.Nodes["brokers"]

			actual, _ := makeConfigMapForNodeSpec(&nodeSpec, clusterSpec, makeLabelsForNodeSpec(&nodeSpec, clusterSpec, clusterSpec.Name, nodeSpecUniqueStr), nodeSpecUniqueStr)
			addHashToObject(actual)

			expected := new(corev1.ConfigMap)
			readAndUnmarshallResource("testdata/broker-config-map.yaml", &expected)
			Expect(err).Should(BeNil())

			Expect(actual).Should(Equal(expected))
		})

	})
})

func readDruidClusterSpecFromFile(filePath string) (*druidv1alpha1.Druid, error) {
	clusterSpec := new(druidv1alpha1.Druid)
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return clusterSpec, err
	}

	err = yaml.Unmarshal(bytes, &clusterSpec)
	if err != nil {
		return clusterSpec, err
	}
	return clusterSpec, nil
}

func readAndUnmarshallResource(file string, res interface{}) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(bytes, res)
	if err != nil {
		return err
	}
	return nil
}
