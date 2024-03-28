package lookup

import (
	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
)

type Status struct {
	Loaded       bool     `json:"loaded"`
	PendingNodes []string `json:"pendingNodes"`
}

func (r *Status) MergeStatus(status *v1alpha1.DruidLookupStatus) error {
	numberOfPendingNodes := len(r.PendingNodes)

	status.PendingNodes = r.PendingNodes
	status.NumberOfPendingNodes = &numberOfPendingNodes
	status.Loaded = &r.Loaded

	return nil
}
