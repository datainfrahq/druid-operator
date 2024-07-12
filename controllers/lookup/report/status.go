package report

import "github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"

type StatusReport struct {
	Loaded       bool     `json:"loaded"`
	PendingNodes []string `json:"pendingNodes"`
}

func (r *StatusReport) MergeStatus(status *v1alpha1.DruidLookupStatus) error {
	numberOfPendingNodes := len(r.PendingNodes)

	status.PendingNodes = r.PendingNodes
	status.NumberOfPendingNodes = &numberOfPendingNodes
	status.Loaded = &r.Loaded

	return nil
}

func (r *StatusReport) ShouldResultInRequeue() bool {
	return !r.Loaded
}
