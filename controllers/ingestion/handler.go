package ingestion

import (
	"net/http"

	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	"github.com/datainfrahq/druid-operator/controllers/druid"
	internalHTTP "github.com/datainfrahq/druid-operator/pkg/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	supervisorPrefix = "/druid/indexer/v1/supervisor"
)

func deployDruidIngestion(sdk client.Client, dI *v1alpha1.DruidIngestion, emitEvents druid.EventEmitter) error {

	switch dI.Spec.IngestionSpec.Type {
	case v1alpha1.Kafka:
		client := internalHTTP.NewHTTPClient(http.MethodPost, "http://"+dI.Spec.RouterURL+supervisorPrefix, http.Client{}, []byte(dI.Spec.IngestionSpec.SupervisorSpec))
		_, err := client.Do()
		if err != nil {
			return err
		}
	case v1alpha1.Kinesis:
	case v1alpha1.HadoopIndexHadoop:
	case v1alpha1.QueryControllerSQL:
	case v1alpha1.NativeBatchIndexParallel:
	}

	return nil
}
