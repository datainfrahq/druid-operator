package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestDruidLookup_ShouldDeleteWhenClusterChanges(t *testing.T) {
	// Arrange
	lookup := DruidLookup{
		Spec: DruidLookupSpec{
			DruidCluster: v1.LocalObjectReference{
				Name: "clusterB",
			},
			Tier: "tierA",
		},
		Status: DruidLookupStatus{
			LastClusterAppliedIn: v1.LocalObjectReference{
				Name: "clusterA",
			},
			LastTierAppliedIn: "tierA",
		},
	}

	// Act
	shouldDelete := lookup.ShouldDeleteLastAppliedLookup()

	// Assert
	if !shouldDelete {
		t.Fatalf("Did not indicate that last applied lookup needs to be deleted")
	}
}

func TestDruidLookup_ShouldDeleteWhenTierChanges(t *testing.T) {
	// Arrange
	lookup := DruidLookup{
		Spec: DruidLookupSpec{
			DruidCluster: v1.LocalObjectReference{
				Name: "clusterA",
			},
			Tier: "tierB",
		},
		Status: DruidLookupStatus{
			LastClusterAppliedIn: v1.LocalObjectReference{
				Name: "clusterA",
			},
			LastTierAppliedIn: "tierA",
		},
	}

	// Act
	shouldDelete := lookup.ShouldDeleteLastAppliedLookup()

	// Assert
	if !shouldDelete {
		t.Fatalf("Did not indicate that last applied lookup needs to be deleted")
	}
}

func TestDruidLookup_ShouldNotDeleteWhenNotPreviouslyApplied(t *testing.T) {
	// Arrange
	lookup := DruidLookup{
		Spec: DruidLookupSpec{
			DruidCluster: v1.LocalObjectReference{
				Name: "clusterB",
			},
			Tier: "tierA",
		},
		Status: DruidLookupStatus{},
	}

	// Act
	shouldDelete := lookup.ShouldDeleteLastAppliedLookup()

	// Assert
	if shouldDelete {
		t.Fatalf("Did indicate that last applied lookup needs to be deleted")
	}
}

func TestDruidLookup_ShouldNotDeleteWhenTierAndClusterAreUnchanged(t *testing.T) {
	// Arrange
	lookup := DruidLookup{
		Spec: DruidLookupSpec{
			DruidCluster: v1.LocalObjectReference{
				Name: "clusterA",
			},
			Tier: "tierA",
		},
		Status: DruidLookupStatus{
			LastClusterAppliedIn: v1.LocalObjectReference{
				Name: "clusterA",
			},
			LastTierAppliedIn: "tierA",
		},
	}

	// Act
	shouldDelete := lookup.ShouldDeleteLastAppliedLookup()

	// Assert
	if shouldDelete {
		t.Fatalf("Did indicate that last applied lookup needs to be deleted")
	}
}

func TestDruidLookup_ShouldGetTemplateToApplyWhenNotPreviouslyApplied(t *testing.T) {
	// Arrange
	lookup := DruidLookup{
		Spec: DruidLookupSpec{
			DruidCluster: v1.LocalObjectReference{
				Name: "clusterA",
			},
			Tier:     "tierA",
			Template: `{"type": "map","map": {"SE": "Sweden"}}`,
		},
		Status: DruidLookupStatus{},
	}

	// Act
	template, err := lookup.GetTemplateToApply()

	// Assert
	if err != nil {
		t.Fatalf("An unexpected error occurred: %v", err)
	}
	if template == nil {
		t.Fatalf("Did not get a template to apply")
	}
}

func TestDruidLookup_ShouldGetTemplateToApplyWhenClusterChanges(t *testing.T) {
	// Arrange
	lookup := DruidLookup{
		Spec: DruidLookupSpec{
			DruidCluster: v1.LocalObjectReference{
				Name: "clusterB",
			},
			Tier:     "tierA",
			Template: `{"type": "map","map": {"SE": "Sweden"}}`,
		},
		Status: DruidLookupStatus{
			LastAppliedTemplate: `{"type": "map","map": {"SE": "Sweden"}}`,
			LastClusterAppliedIn: v1.LocalObjectReference{
				Name: "clusterA",
			},
			LastTierAppliedIn: "tierA",
		},
	}

	// Act
	template, err := lookup.GetTemplateToApply()

	// Assert
	if err != nil {
		t.Fatalf("An unexpected error occurred: %v", err)
	}
	if template == nil {
		t.Fatalf("Did not get a template to apply")
	}
}

func TestDruidLookup_ShouldGetTemplateToApplyWhenTierChanges(t *testing.T) {
	// Arrange
	lookup := DruidLookup{
		Spec: DruidLookupSpec{
			DruidCluster: v1.LocalObjectReference{
				Name: "clusterA",
			},
			Tier:     "tierB",
			Template: `{"type": "map","map": {"SE": "Sweden"}}`,
		},
		Status: DruidLookupStatus{
			LastAppliedTemplate: `{"type": "map","map": {"SE": "Sweden"}}`,
			LastClusterAppliedIn: v1.LocalObjectReference{
				Name: "clusterA",
			},
			LastTierAppliedIn: "tierA",
		},
	}

	// Act
	template, err := lookup.GetTemplateToApply()

	// Assert
	if err != nil {
		t.Fatalf("An unexpected error occurred: %v", err)
	}
	if template == nil {
		t.Fatalf("Did not get a template to apply")
	}
}

func TestDruidLookup_ShouldGetTemplateToApplyWhenTemplateChanges(t *testing.T) {
	// Arrange
	lookup := DruidLookup{
		Spec: DruidLookupSpec{
			DruidCluster: v1.LocalObjectReference{
				Name: "clusterA",
			},
			Tier:     "tierA",
			Template: `{"type": "map","map": {"SE": "Sweden"}}`,
		},
		Status: DruidLookupStatus{
			LastAppliedTemplate: `{"type": "map","map": {"SE": "Denmark"}}`,
			LastClusterAppliedIn: v1.LocalObjectReference{
				Name: "clusterA",
			},
			LastTierAppliedIn: "tierA",
		},
	}

	// Act
	template, err := lookup.GetTemplateToApply()

	// Assert
	if err != nil {
		t.Fatalf("An unexpected error occurred: %v", err)
	}
	if template == nil {
		t.Fatalf("Did not get a template to apply")
	}
}

func TestDruidLookup_ShouldNotGetTemplateToApplyWhenClusterTierNorTemplateChanges(t *testing.T) {
	// Arrange
	lookup := DruidLookup{
		Spec: DruidLookupSpec{
			DruidCluster: v1.LocalObjectReference{
				Name: "clusterA",
			},
			Tier:     "tierA",
			Template: `{"type": "map","map": {"SE": "Sweden"}}`,
		},
		Status: DruidLookupStatus{
			LastAppliedTemplate: `{"type": "map","map": {"SE": "Sweden"}}`,
			LastClusterAppliedIn: v1.LocalObjectReference{
				Name: "clusterA",
			},
			LastTierAppliedIn: "tierA",
		},
	}

	// Act
	template, err := lookup.GetTemplateToApply()

	// Assert
	if err != nil {
		t.Fatalf("An unexpected error occurred: %v", err)
	}
	if template != nil {
		t.Fatalf("Did get a template to apply")
	}
}
