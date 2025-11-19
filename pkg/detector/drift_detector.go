package detector

import (
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
)

// AttributeDrift represents a single attribute change
type AttributeDrift struct {
	Attribute string
	OldValue  interface{}
	NewValue  interface{}
}

// detectDrifts detects attribute changes
func (d *Detector) detectDrifts(resource *terraform.Resource, changes map[string]interface{}) []AttributeDrift {
	var drifts []AttributeDrift

	for key, newValue := range changes {
		oldValue, exists := resource.Attributes[key]
		if !exists || oldValue != newValue {
			drifts = append(drifts, AttributeDrift{
				Attribute: key,
				OldValue:  oldValue,
				NewValue:  newValue,
			})
		}
	}

	return drifts
}

// evaluateRules evaluates drift rules
func (d *Detector) evaluateRules(resourceType, attribute string) []string {
	var matched []string

	for _, rule := range d.cfg.DriftRules {
		// Check if resource type matches
		typeMatch := false
		for _, rt := range rule.ResourceTypes {
			if rt == resourceType {
				typeMatch = true
				break
			}
		}
		if !typeMatch {
			continue
		}

		// Check if attribute matches
		attrMatch := false
		for _, wa := range rule.WatchedAttributes {
			if wa == attribute {
				attrMatch = true
				break
			}
		}
		if !attrMatch {
			continue
		}

		matched = append(matched, rule.Name)
	}

	return matched
}

// getSeverity determines the highest severity from matched rules
func (d *Detector) getSeverity(matchedRules []string) string {
	severity := "low"

	for _, ruleName := range matchedRules {
		for _, rule := range d.cfg.DriftRules {
			if rule.Name == ruleName {
				if rule.Severity == "critical" {
					return "critical"
				}
				if rule.Severity == "high" && severity != "critical" {
					severity = "high"
				}
				if rule.Severity == "medium" && severity == "low" {
					severity = "medium"
				}
			}
		}
	}

	return severity
}
