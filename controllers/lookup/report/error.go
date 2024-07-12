package report

import (
	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type ErrorReport struct {
	ts  v1.Time
	err error
}

func NewErrorReport(err error) *ErrorReport {
	return &ErrorReport{
		ts:  v1.Time{Time: time.Now()},
		err: err,
	}
}

func (r *ErrorReport) MergeStatus(status *v1alpha1.DruidLookupStatus) error {
	now := &v1.Time{Time: time.Now()}

	status.LastUpdateAttemptAt = now
	status.LastUpdateAttemptSuccessful = false
	status.ErrorMessage = r.err.Error()

	return nil
}

func (r *ErrorReport) ShouldResultInRequeue() bool {
	return true
}
