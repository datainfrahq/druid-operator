// +kubebuilder:docs-gen:collapse=Apache License

package druid

import (
	"context"
	"path/filepath"
	"testing"

	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	druidv1alpha1 "github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

// +kubebuilder:docs-gen:collapse=Imports

// Now, let's go through the code generated.

var (
	cfg       *rest.Config
	k8sClient client.Client // You'll be using this client in your tests.
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	ctrl.SetLogger(zap.New())

	ctx, cancel = context.WithCancel(context.TODO())

	/*
		First, the envtest cluster is configured to read CRDs from the CRD directory Kubebuilder scaffolds for you.
	*/
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDInstallOptions: envtest.CRDInstallOptions{
			Paths: []string{filepath.Join("..", "..", "config", "crd", "bases")},
		},
		ErrorIfCRDPathMissing: true,
	}

	//	Then, we start the envtest cluster.
	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = druidv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	/*
		After the schemas, you will see the following marker.
		This marker is what allows new schemas to be added here automatically when a new API is added to the project.
	*/

	//+kubebuilder:scaffold:scheme

	// A client is created for our test CRUD operations.
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&DruidReconciler{
		Client:        k8sManager.GetClient(),
		Log:           ctrl.Log.WithName("controllers").WithName("Druid"),
		Scheme:        k8sManager.GetScheme(),
		ReconcileWait: LookupReconcileTime(),
		Recorder:      k8sManager.GetEventRecorderFor("druid-operator"),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

})

/*
Kubebuilder also generates boilerplate functions for cleaning up envtest and actually running your test files in your controllers/ directory.
You won't need to touch these.
*/

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
