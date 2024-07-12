package lookup

import (
	"encoding/json"
	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type Report interface {
	MergeStatus(status *v1alpha1.DruidLookupStatus) error
	ShouldResultInRequeue() bool
}

type SuccessReport struct {
	ts      metav1.Time
	spec    interface{}
	cluster v1.LocalObjectReference
	tier    string
}

func NewSuccessReport(cluster v1.LocalObjectReference, tier string, spec interface{}) *SuccessReport {
	return &SuccessReport{
		ts:      metav1.Time{Time: time.Now()},
		spec:    spec,
		cluster: cluster,
		tier:    tier,
	}
}

func (r *SuccessReport) MergeStatus(status *v1alpha1.DruidLookupStatus) error {
	now := &metav1.Time{Time: time.Now()}

	spec, err := json.Marshal(r.spec)
	if err != nil {
		return err
	}

	status.LastClusterAppliedIn = r.cluster
	status.LastTierAppliedIn = r.tier
	status.LastAppliedTemplate = string(spec)
	status.LastSuccessfulUpdateAt = now
	status.LastUpdateAttemptAt = now
	status.LastUpdateAttemptSuccessful = true
	status.ErrorMessage = ""

	return nil
}

func (r *SuccessReport) ShouldResultInRequeue() bool {
	return false
}

type ErrorReport struct {
	ts  metav1.Time
	err error
}

func NewErrorReport(err error) *ErrorReport {
	return &ErrorReport{
		ts:  metav1.Time{Time: time.Now()},
		err: err,
	}
}

func (r *ErrorReport) MergeStatus(status *v1alpha1.DruidLookupStatus) error {
	now := &metav1.Time{Time: time.Now()}

	status.LastUpdateAttemptAt = now
	status.LastUpdateAttemptSuccessful = false
	status.ErrorMessage = r.err.Error()

	return nil
}

func (r *ErrorReport) ShouldResultInRequeue() bool {
	return true
}
