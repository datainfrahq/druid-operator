package druid

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalev2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResourceConfig defines configuration for each resource type cleanup
type ResourceConfig struct {
	Name             string
	ExpectedNames    map[string]bool
	EmptyListObjFn   func() objectList
	ItemsExtractorFn func(obj runtime.Object) []object
}

// ResourceCleanupResult holds the result of cleaning up a specific resource type
type ResourceCleanupResult struct {
	ResourceType  string
	SurvivorNames []string
	Error         error
}

// ConsolidatedResourceCleanupResult holds all cleanup results
type ConsolidatedResourceCleanupResult struct {
	Status *v1alpha1.DruidClusterStatus
	Errors []error
}

// deleteAllUnusedResources consolidates all resource cleanup operations into parallel execution
func deleteAllUnusedResources(
	ctx context.Context,
	sdk client.Client,
	drd *v1alpha1.Druid,
	selectorLabels map[string]string,
	statefulSetNames map[string]bool,
	deploymentNames map[string]bool,
	serviceNames map[string]bool,
	configMapNames map[string]bool,
	podDisruptionBudgetNames map[string]bool,
	hpaNames map[string]bool,
	ingressNames map[string]bool,
	emitEvents EventEmitter,
) (*ConsolidatedResourceCleanupResult, error) {

	// Define all resource types to clean up
	resourceConfigs := []ResourceConfig{
		{
			Name:           "StatefulSets",
			ExpectedNames:  statefulSetNames,
			EmptyListObjFn: func() objectList { return &appsv1.StatefulSetList{} },
			ItemsExtractorFn: func(listObj runtime.Object) []object {
				items := listObj.(*appsv1.StatefulSetList).Items
				result := make([]object, len(items))
				for i := 0; i < len(items); i++ {
					result[i] = &items[i]
				}
				return result
			},
		},
		{
			Name:           "Deployments",
			ExpectedNames:  deploymentNames,
			EmptyListObjFn: func() objectList { return &appsv1.DeploymentList{} },
			ItemsExtractorFn: func(listObj runtime.Object) []object {
				items := listObj.(*appsv1.DeploymentList).Items
				result := make([]object, len(items))
				for i := 0; i < len(items); i++ {
					result[i] = &items[i]
				}
				return result
			},
		},
		{
			Name:           "Services",
			ExpectedNames:  serviceNames,
			EmptyListObjFn: func() objectList { return &v1.ServiceList{} },
			ItemsExtractorFn: func(listObj runtime.Object) []object {
				items := listObj.(*v1.ServiceList).Items
				result := make([]object, len(items))
				for i := 0; i < len(items); i++ {
					result[i] = &items[i]
				}
				return result
			},
		},
		{
			Name:           "ConfigMaps",
			ExpectedNames:  configMapNames,
			EmptyListObjFn: func() objectList { return &v1.ConfigMapList{} },
			ItemsExtractorFn: func(listObj runtime.Object) []object {
				items := listObj.(*v1.ConfigMapList).Items
				result := make([]object, len(items))
				for i := 0; i < len(items); i++ {
					result[i] = &items[i]
				}
				return result
			},
		},
		{
			Name:           "HPAutoScalers",
			ExpectedNames:  hpaNames,
			EmptyListObjFn: func() objectList { return &autoscalev2.HorizontalPodAutoscalerList{} },
			ItemsExtractorFn: func(listObj runtime.Object) []object {
				items := listObj.(*autoscalev2.HorizontalPodAutoscalerList).Items
				result := make([]object, len(items))
				for i := 0; i < len(items); i++ {
					result[i] = &items[i]
				}
				return result
			},
		},
		{
			Name:           "Ingress",
			ExpectedNames:  ingressNames,
			EmptyListObjFn: func() objectList { return &networkingv1.IngressList{} },
			ItemsExtractorFn: func(listObj runtime.Object) []object {
				items := listObj.(*networkingv1.IngressList).Items
				result := make([]object, len(items))
				for i := 0; i < len(items); i++ {
					result[i] = &items[i]
				}
				return result
			},
		},
		{
			Name:           "PodDisruptionBudgets",
			ExpectedNames:  podDisruptionBudgetNames,
			EmptyListObjFn: func() objectList { return &policyv1.PodDisruptionBudgetList{} },
			ItemsExtractorFn: func(listObj runtime.Object) []object {
				items := listObj.(*policyv1.PodDisruptionBudgetList).Items
				result := make([]object, len(items))
				for i := 0; i < len(items); i++ {
					result[i] = &items[i]
				}
				return result
			},
		},
	}

	// Channel to collect results from parallel goroutines
	resultChan := make(chan ResourceCleanupResult, len(resourceConfigs))
	var wg sync.WaitGroup

	// Launch parallel cleanup operations
	for _, config := range resourceConfigs {
		wg.Add(1)
		go func(cfg ResourceConfig) {
			defer wg.Done()

			// Call the existing deleteUnusedResources function for this resource type
			survivors := deleteUnusedResources(
				ctx, sdk, drd, cfg.ExpectedNames, selectorLabels,
				cfg.EmptyListObjFn, cfg.ItemsExtractorFn, emitEvents,
			)

			// Send result to channel
			resultChan <- ResourceCleanupResult{
				ResourceType:  cfg.Name,
				SurvivorNames: survivors,
				Error:         nil, // deleteUnusedResources doesn't return errors currently
			}
		}(config)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultChan)

	// Collect all results
	status := &v1alpha1.DruidClusterStatus{}
	var errors []error

	for result := range resultChan {
		if result.Error != nil {
			errors = append(errors, fmt.Errorf("failed to cleanup %s: %w", result.ResourceType, result.Error))
			continue
		}

		// Assign results to appropriate status fields
		switch result.ResourceType {
		case "StatefulSets":
			status.StatefulSets = result.SurvivorNames
		case "Deployments":
			status.Deployments = result.SurvivorNames
		case "Services":
			status.Services = result.SurvivorNames
		case "ConfigMaps":
			status.ConfigMaps = result.SurvivorNames
		case "HPAutoScalers":
			status.HPAutoScalers = result.SurvivorNames
		case "Ingress":
			status.Ingress = result.SurvivorNames
		case "PodDisruptionBudgets":
			status.PodDisruptionBudgets = result.SurvivorNames
		}
	}

	// Sort all result slices for consistency (matching original behavior)
	sort.Strings(status.StatefulSets)
	sort.Strings(status.Deployments)
	sort.Strings(status.Services)
	sort.Strings(status.ConfigMaps)
	sort.Strings(status.HPAutoScalers)
	sort.Strings(status.Ingress)
	sort.Strings(status.PodDisruptionBudgets)

	return &ConsolidatedResourceCleanupResult{
		Status: status,
		Errors: errors,
	}, nil
}
