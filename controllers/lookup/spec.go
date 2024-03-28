package lookup

import "k8s.io/apimachinery/pkg/types"

type Spec struct {
	name types.NamespacedName
	spec interface{}
}
