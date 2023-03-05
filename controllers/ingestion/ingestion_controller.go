package ingestion

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	druidv1alpha1 "github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	"github.com/datainfrahq/druid-operator/controllers/druid"
)

// DruidReconciler reconciles a Druid object
type IngestionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	// reconcile time duration, defaults to 10s
	ReconcileWait time.Duration
	Recorder      record.EventRecorder
}

func NewDruidReconciler(mgr ctrl.Manager) *IngestionReconciler {
	return &IngestionReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("controllers").WithName("Ingestion"),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("druid-operator"),
	}
}

// +kubebuilder:rbac:groups=druid.apache.org,resources=ingestions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=druid.apache.org,resources=ingestion/status,verbs=get;update;patch
func (r *IngestionReconciler) Reconcile(ctx context.Context, request reconcile.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("ingestion", request.NamespacedName)
	fmt.Println("hello")

	// Fetch the Druid instance
	instance := &druidv1alpha1.DruidIngestion{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	var emitEvent druid.EventEmitter = druid.EmitEventFuncs{r.Recorder}
	var readers druid.Reader = druid.ReaderFuncs{}

	if err := deployDruidIngestion(r.Client, instance, emitEvent); err != nil {
		return ctrl.Result{}, err
	} else {
		return ctrl.Result{RequeueAfter: r.ReconcileWait}, nil
	}
	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

func (r *IngestionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&druidv1alpha1.DruidIngestion{}).
		Complete(r)
}
