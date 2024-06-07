package lookup

import "k8s.io/apimachinery/pkg/types"

type LookupKey struct {
	Tier string
	Id   string
}

type LookupsPerCluster map[types.NamespacedName]map[LookupKey]Spec

type Spec struct {
	name types.NamespacedName
	spec interface{}
}
