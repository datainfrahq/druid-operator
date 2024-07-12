package report

import "github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"

type NothingChangedReport struct{}

func NewNothingChangedReport() *NothingChangedReport {
	return &NothingChangedReport{}
}

func (r *NothingChangedReport) MergeStatus(status *v1alpha1.DruidLookupStatus) error {
	return nil
}

func (r *NothingChangedReport) ShouldResultInRequeue() bool {
	return false
}
