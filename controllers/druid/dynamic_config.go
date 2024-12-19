package druid

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	druidapi "github.com/datainfrahq/druid-operator/pkg/druidapi"
	internalhttp "github.com/datainfrahq/druid-operator/pkg/http"
	"github.com/datainfrahq/druid-operator/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// updateDruidDynamicConfigs updates the Druid cluster's dynamic configurations.
func updateDruidDynamicConfigs(ctx context.Context, client client.Client, druid *v1alpha1.Druid, emitEvent EventEmitter) error {
	if druid.Spec.DynamicConfig.Size() == 0 {
		return nil
	}

	svcName, err := druidapi.GetRouterSvcUrl(druid.Namespace, druid.Name, client)
	if err != nil {
		emitEvent.EmitEventGeneric(druid, "GetRouterSvcUrlFailed", "Failed to get router service URL", err)
		return err
	}

	basicAuth, err := druidapi.GetAuthCreds(
		ctx,
		client,
		druid.Spec.Auth,
	)
	if err != nil {
		emitEvent.EmitEventGeneric(druid, "GetAuthCredsFailed", "Failed to get authentication credentials", err)
		return err
	}

	// Create the HTTP client with basic authentication
	httpClient := internalhttp.NewHTTPClient(
		&http.Client{},
		&internalhttp.Auth{BasicAuth: basicAuth},
	)

	// Define the URL path for dynamic configurations
	dynamicConfigPath := druidapi.MakePath(svcName, "indexer", "worker")

	// Fetch current dynamic configurations
	currentResp, err := httpClient.Do(
		http.MethodGet,
		dynamicConfigPath,
		nil,
	)
	if err != nil {
		emitEvent.EmitEventGeneric(druid, "FetchCurrentConfigsFailed", "Failed to fetch current dynamic configurations", err)
		return err
	}
	if currentResp.StatusCode != http.StatusOK {
		err = fmt.Errorf(
			"failed to retrieve current Druid dynamic configurations. Status code: %d, Response body: %s",
			currentResp.StatusCode, string(currentResp.ResponseBody),
		)
		emitEvent.EmitEventGeneric(druid, "FetchCurrentConfigsFailed", "Failed to fetch current dynamic configurations", err)
		return err
	}

	// Handle empty response body
	var currentConfigsJson string
	if len(currentResp.ResponseBody) == 0 {
		currentConfigsJson = "{}" // Initialize as empty JSON object if response body is empty
	} else {
		currentConfigsJson = currentResp.ResponseBody
	}

	// Compare current and desired configurations
	equal, err := util.IncludesJson(currentConfigsJson, string(druid.Spec.DynamicConfig.Raw))
	if err != nil {
		emitEvent.EmitEventGeneric(druid, "ConfigComparisonFailed", "Failed to compare configurations", err)
		return err
	}
	if equal {
		// Configurations are already up-to-date
		emitEvent.EmitEventGeneric(druid, "ConfigsUpToDate", "Dynamic configurations are already up-to-date", nil)
		return nil
	}

	// Update the Druid cluster's dynamic configurations if needed
	respDynamicConfigs, err := httpClient.Do(
		http.MethodPost,
		dynamicConfigPath,
		druid.Spec.DynamicConfig.Raw,
	)
	if err != nil {
		emitEvent.EmitEventGeneric(druid, "UpdateConfigsFailed", "Failed to update dynamic configurations", err)
		return err
	}
	if respDynamicConfigs.StatusCode != http.StatusOK {
		return errors.New("failed to update Druid dynamic configurations")
	}

	emitEvent.EmitEventGeneric(druid, "ConfigsUpdated", "Successfully updated dynamic configurations", nil)
	return nil
}
