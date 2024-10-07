package report

import (
	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type SuccessReport struct {
	ts       meta.Time
	template string
	cluster  core.LocalObjectReference
	tier     string
}

func NewSuccessReport(cluster core.LocalObjectReference, tier string, template string) *SuccessReport {
	return &SuccessReport{
		ts:       meta.Time{Time: time.Now()},
		template: template,
		cluster:  cluster,
		tier:     tier,
	}
}

func (r *SuccessReport) MergeStatus(status *v1alpha1.DruidLookupStatus) error {
	now := &meta.Time{Time: time.Now()}

	status.LastClusterAppliedIn = r.cluster
	status.LastTierAppliedIn = r.tier
	status.LastAppliedTemplate = r.template
	status.LastSuccessfulUpdateAt = now
	status.LastUpdateAttemptAt = now
	status.LastUpdateAttemptSuccessful = true
	status.ErrorMessage = ""

	return nil
}

func (r *SuccessReport) ShouldResultInRequeue() bool {
	return false
}
