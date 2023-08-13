package druid

import (
	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

func addAdditionalContainers(m *v1alpha1.Druid, nodeSpec *v1alpha1.DruidNodeSpec, podSpec *v1.PodSpec) {
	var allAdditional []v1alpha1.AdditionalContainer
	if m.Spec.AdditionalContainer != nil {
		allAdditional = append(allAdditional, m.Spec.AdditionalContainer...)
	}
	if nodeSpec.AdditionalContainer != nil {
		allAdditional = append(allAdditional, nodeSpec.AdditionalContainer...)
	}

	for _, additional := range allAdditional {
		container := convertAdditionalContainer(&additional)

		if additional.RunAsInit {
			podSpec.InitContainers = append(podSpec.InitContainers, container)
		} else {
			podSpec.Containers = append(podSpec.Containers, container)
		}
	}
}

func convertAdditionalContainer(additional *v1alpha1.AdditionalContainer) v1.Container {
	return v1.Container{
		Image:           additional.Image,
		Name:            additional.ContainerName,
		Resources:       additional.Resources,
		VolumeMounts:    additional.VolumeMounts,
		Command:         additional.Command,
		Args:            additional.Args,
		ImagePullPolicy: additional.ImagePullPolicy,
		SecurityContext: additional.ContainerSecurityContext,
		Env:             additional.Env,
		EnvFrom:         additional.EnvFrom,
	}
}
