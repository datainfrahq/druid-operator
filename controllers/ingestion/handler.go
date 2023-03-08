package ingestion

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	"github.com/datainfrahq/druid-operator/pkg/common"
	internalHTTP "github.com/datainfrahq/druid-operator/pkg/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	supervisorPrefix = "/druid/indexer/v1/supervisor"
)

func deployDruidIngestion(sdk client.Client, dI *v1alpha1.DruidIngestion) error {

	switch dI.Spec.IngestionSpec.Type {
	case v1alpha1.Kafka:

		if err := CreateUpdate(sdk, dI); err != nil {
			return err
		}
	case v1alpha1.Kinesis:
	case v1alpha1.HadoopIndexHadoop:
	case v1alpha1.QueryControllerSQL:
	case v1alpha1.NativeBatchIndexParallel:
	}

	return nil
}

func CreateUpdate(sdk client.Client, dI *v1alpha1.DruidIngestion) error {
	var currentObj v1alpha1.DruidIngestion
	if err := sdk.Get(context.TODO(), *common.NamespacedName(dI.GetName(), dI.GetNamespace()), &currentObj); err != nil {
		return err
	}

	if !reflect.DeepEqual(&currentObj.Spec.IngestionSpec, dI.Spec.IngestionSpec) {
		fmt.Println("reconcile")
		client := internalHTTP.NewHTTPClient(http.MethodPost, "http://"+dI.Spec.RouterURL+supervisorPrefix, http.Client{}, []byte(dI.Spec.IngestionSpec.SupervisorSpec))
		_, err := client.Do()
		if err != nil {
			return err
		}
	}

	return nil
}
