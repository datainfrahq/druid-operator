package druid

import (
	"fmt"

	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

func addAdditionalContainers(m *v1alpha1.Druid, nodeSpec *v1alpha1.DruidNodeSpec, podSpec *v1.PodSpec) {
	allAdditional := getAllAdditionalContainers(m, nodeSpec)

	for _, additional := range allAdditional {
		container := convertAdditionalContainer(&additional)

		if additional.RunAsInit {
			podSpec.InitContainers = append(podSpec.InitContainers, container)
		} else {
			podSpec.Containers = append(podSpec.Containers, container)
		}
	}
}

func getAllAdditionalContainers(m *v1alpha1.Druid, nodeSpec *v1alpha1.DruidNodeSpec) []v1alpha1.AdditionalContainer {
	var allAdditional []v1alpha1.AdditionalContainer
	if m.Spec.AdditionalContainer != nil {
		allAdditional = append(allAdditional, m.Spec.AdditionalContainer...)
	}
	if nodeSpec.AdditionalContainer != nil {
		allAdditional = append(allAdditional, nodeSpec.AdditionalContainer...)
	}
	return allAdditional
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

func validateAdditionalContainersSpec(drd *v1alpha1.Druid) error {
	for _, nodeSpec := range drd.Spec.Nodes {
		if err := validateNodeAdditionalContainersSpec(drd, &nodeSpec); err != nil {
			return err
		}
	}
	return nil
}

func validateNodeAdditionalContainersSpec(drd *v1alpha1.Druid, nodeSpec *v1alpha1.DruidNodeSpec) error {
	allAdditional := getAllAdditionalContainers(drd, nodeSpec)
	var containerNames []string
	for _, container := range allAdditional {
		containerNames = append(containerNames, container.ContainerName)
	}
	if duplicate, containerName := hasDuplicateString(containerNames); duplicate {
		return fmt.Errorf("node group %s has duplicate container name: %s",
			nodeSpec.NodeType, containerName)
	}
	return nil
}
