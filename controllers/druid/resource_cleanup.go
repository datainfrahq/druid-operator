package druid

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalev2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Constants for resource type names to avoid repetition
const (
	ResourceTypeStatefulSets         = "StatefulSets"
	ResourceTypeDeployments          = "Deployments"
	ResourceTypeServices             = "Services"
	ResourceTypeConfigMaps           = "ConfigMaps"
	ResourceTypeHPAutoScalers        = "HPAutoScalers"
	ResourceTypeIngress              = "Ingress"
	ResourceTypePodDisruptionBudgets = "PodDisruptionBudgets"
)

// ResourceExpectations - Simple map of resource types to expected names
type ResourceExpectations map[string]map[string]bool

// ConsolidatedResourceCleanupResult holds all cleanup results
type ConsolidatedResourceCleanupResult struct {
	Status *v1alpha1.DruidClusterStatus
	Errors []error
}

// ResourceTypeConfig holds the configuration for each resource type
type ResourceTypeConfig struct {
	Name       string
	CreateList func() client.ObjectList
}

// deleteAllUnusedResources - THE SIMPLEST, CLEANEST VERSION
func deleteAllUnusedResources(
	ctx context.Context,
	sdk client.Client,
	drd *v1alpha1.Druid,
	selectorLabels map[string]string,
	expectedResources ResourceExpectations,
	emitEvents EventEmitter,
) (*ConsolidatedResourceCleanupResult, error) {

	// Define all resource types with proper type safety - NO MORE REPETITION!
	resourceTypes := map[string]ResourceTypeConfig{
		ResourceTypeStatefulSets:         {ResourceTypeStatefulSets, func() client.ObjectList { return &appsv1.StatefulSetList{} }},
		ResourceTypeDeployments:          {ResourceTypeDeployments, func() client.ObjectList { return &appsv1.DeploymentList{} }},
		ResourceTypeServices:             {ResourceTypeServices, func() client.ObjectList { return &v1.ServiceList{} }},
		ResourceTypeConfigMaps:           {ResourceTypeConfigMaps, func() client.ObjectList { return &v1.ConfigMapList{} }},
		ResourceTypeHPAutoScalers:        {ResourceTypeHPAutoScalers, func() client.ObjectList { return &autoscalev2.HorizontalPodAutoscalerList{} }},
		ResourceTypeIngress:              {ResourceTypeIngress, func() client.ObjectList { return &networkingv1.IngressList{} }},
		ResourceTypePodDisruptionBudgets: {ResourceTypePodDisruptionBudgets, func() client.ObjectList { return &policyv1.PodDisruptionBudgetList{} }},
	}

	status := &v1alpha1.DruidClusterStatus{}
	resultChan := make(chan struct {
		resourceType string
		survivors    []string
		err          error
	}, len(resourceTypes))
	var wg sync.WaitGroup

	// Process all resource types in parallel
	for resourceType, config := range resourceTypes {
		wg.Add(1)
		go func(resType string, cfg ResourceTypeConfig) {
			defer wg.Done()

			// Get expected names, default to empty if not provided
			expectedNames := expectedResources[resType]
			if expectedNames == nil {
				expectedNames = make(map[string]bool)
			}

			// Generic cleanup
			survivors, err := cleanupSingleResourceType(
				ctx, sdk, drd, cfg, expectedNames, selectorLabels, emitEvents,
			)

			resultChan <- struct {
				resourceType string
				survivors    []string
				err          error
			}{resType, survivors, err}
		}(resourceType, config)
	}

	// Wait and collect results
	wg.Wait()
	close(resultChan)

	var errors []error
	for result := range resultChan {
		if result.err != nil {
			errors = append(errors, result.err)
			continue
		}

		sort.Strings(result.survivors)

		// Update status fields
		switch result.resourceType {
		case ResourceTypeStatefulSets:
			status.StatefulSets = result.survivors
		case ResourceTypeDeployments:
			status.Deployments = result.survivors
		case ResourceTypeServices:
			status.Services = result.survivors
		case ResourceTypeConfigMaps:
			status.ConfigMaps = result.survivors
		case ResourceTypeHPAutoScalers:
			status.HPAutoScalers = result.survivors
		case ResourceTypeIngress:
			status.Ingress = result.survivors
		case ResourceTypePodDisruptionBudgets:
			status.PodDisruptionBudgets = result.survivors
		}
	}

	return &ConsolidatedResourceCleanupResult{
		Status: status,
		Errors: errors,
	}, nil
}

// Generic cleanup for any resource type
func cleanupSingleResourceType(
	ctx context.Context,
	sdk client.Client,
	drd *v1alpha1.Druid,
	config ResourceTypeConfig,
	expectedNames map[string]bool,
	selectorLabels map[string]string,
	emitEvents EventEmitter,
) ([]string, error) {

	// Create list object with proper type safety
	listObj := config.CreateList()

	// List resources
	listOpts := []client.ListOption{
		client.InNamespace(drd.Namespace),
		client.MatchingLabels(selectorLabels),
	}

	if err := sdk.List(ctx, listObj, listOpts...); err != nil {
		return nil, fmt.Errorf("failed to list %s: %w", config.Name, err)
	}

	// Extract items using reflection (still needed to be generic across types)
	items := extractItemsFromList(listObj)
	survivorNames := make([]string, 0, len(expectedNames))

	for _, item := range items {
		itemMeta := item.(client.Object)
		name := itemMeta.GetName()

		if !expectedNames[name] {
			// Delete unexpected resource
			if err := writers.Delete(ctx, sdk, drd, item.(object), emitEvents, &client.DeleteOptions{}); err != nil {
				survivorNames = append(survivorNames, name) // Failed to delete, so it's a survivor
			}
		} else {
			// Keep expected resource
			survivorNames = append(survivorNames, name)
		}
	}

	return survivorNames, nil
}

// Extract items from any Kubernetes list object using reflection
func extractItemsFromList(listObj client.ObjectList) []interface{} {
	// Use reflection to get the Items field from any list type
	listValue := reflect.ValueOf(listObj).Elem()
	itemsField := listValue.FieldByName("Items")

	if !itemsField.IsValid() {
		return nil
	}

	items := make([]interface{}, itemsField.Len())
	for i := 0; i < itemsField.Len(); i++ {
		// Get pointer to the item
		itemValue := itemsField.Index(i)
		items[i] = itemValue.Addr().Interface()
	}
	return items
}

// Helper to build ResourceExpectations from existing variables
func BuildResourceExpectations(
	statefulSetNames, deploymentNames, serviceNames, configMapNames,
	podDisruptionBudgetNames, hpaNames, ingressNames map[string]bool,
) ResourceExpectations {
	return ResourceExpectations{
		ResourceTypeStatefulSets:         statefulSetNames,
		ResourceTypeDeployments:          deploymentNames,
		ResourceTypeServices:             serviceNames,
		ResourceTypeConfigMaps:           configMapNames,
		ResourceTypePodDisruptionBudgets: podDisruptionBudgetNames,
		ResourceTypeHPAutoScalers:        hpaNames,
		ResourceTypeIngress:              ingressNames,
	}
}
