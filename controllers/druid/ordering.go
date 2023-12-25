package druid

import "github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"

var (
	druidOrder = []string{historical, overlord, middleManager, indexer, broker, coordinator, router}
)

// getNodeSpecsByOrder returns all NodeSpecs f a given Druid object.
// Recommended order is described at http://druid.io/docs/latest/operations/rolling-updates.html
func getNodeSpecsByOrder(m *v1alpha1.Druid) []*v1alpha1.ScaledServiceSpec {

	scaledServiceSpecsByNodeType := map[string][]*v1alpha1.ScaledServiceSpec{}
	for _, t := range druidOrder {
		scaledServiceSpecsByNodeType[t] = []*v1alpha1.ScaledServiceSpec{}
	}

	for key, nodeSpec := range m.Spec.Nodes {
		scaledServiceSpec := scaledServiceSpecsByNodeType[nodeSpec.NodeType]
		scaledServiceSpecsByNodeType[nodeSpec.NodeType] = append(scaledServiceSpec, &v1alpha1.ScaledServiceSpec{Key: key, Spec: nodeSpec})
	}

	allScaledServiceSpecs := make([]*v1alpha1.ScaledServiceSpec, 0, len(m.Spec.Nodes))

	for _, t := range druidOrder {
		allScaledServiceSpecs = append(allScaledServiceSpecs, scaledServiceSpecsByNodeType[t]...)
	}

	return allScaledServiceSpecs
}
