package common

import "k8s.io/apimachinery/pkg/types"

func NamespacedName(name, namespace string) *types.NamespacedName {
	return &types.NamespacedName{Name: name, Namespace: namespace}
}
