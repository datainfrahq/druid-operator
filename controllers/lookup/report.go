package lookup

import (
	"encoding/json"
	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type Report interface {
	MergeStatus(status *v1alpha1.DruidLookupStatus) error
}

type SuccessReport struct {
	ts   metav1.Time
	spec interface{}
}

func NewSuccessReport(spec interface{}) *SuccessReport {
	return &SuccessReport{
		ts:   metav1.Time{Time: time.Now()},
		spec: spec,
	}
}

func (r *SuccessReport) MergeStatus(status *v1alpha1.DruidLookupStatus) error {
	now := &metav1.Time{Time: time.Now()}

	spec, err := json.Marshal(r.spec)
	if err != nil {
		return err
	}

	status.LastAppliedSpec = string(spec)
	status.LastSuccessfulUpdateAt = now
	status.LastUpdateAttemptAt = now
	status.LastUpdateAttemptSuccessful = true
	status.ErrorMessage = ""

	return nil
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
