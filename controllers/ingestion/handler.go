package ingestion

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	"github.com/datainfrahq/druid-operator/controllers/druid"
	internalHTTP "github.com/datainfrahq/druid-operator/pkg/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func deployDruidIngestion(sdk client.Client, dI *v1alpha1.DruidIngestion, emitEvents druid.EventEmitter) error {

	body, err := json.Marshal(dI.Spec.SupervisorSpec)
	if err != nil {
		return err
	}

	client := internalHTTP.NewHTTPClient(http.MethodPost, dI.Spec.RouterURL, http.Client{}, body)

	resp, err := client.Do()
	if err != nil {
		return err
	}

	fmt.Println(string(resp))
	return nil
}
