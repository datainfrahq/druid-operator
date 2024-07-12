package report

import (
	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
)

type Report interface {
	MergeStatus(status *v1alpha1.DruidLookupStatus) error
	ShouldResultInRequeue() bool
}
