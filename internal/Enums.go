package internal

import "strings"

// StringEnum represents a string-based enumeration with a fixed set of members.
type StringEnum struct {
	members []string
}

func (d StringEnum) Members() []string {
	members := make([]string, len(d.members))
	copy(members, d.members)
	return members
}

func (d StringEnum) IsMember(id string) bool {
	for _, member := range d.members {
		if strings.EqualFold(member, id) {
			return true
		}
	}
	return false
}

// DebtTypes lists the valid categories of technical debt.
var DebtTypes = StringEnum{
	members: []string{"code", "documentation", "testing", "architecture", "infrastructure", "security"},
}

// DebtStatus lists possible states for a piece of debt.
var DebtStatus = StringEnum{
	members: []string{"pending", "remediated", "in_progress"},
}

var Exposure = StringEnum{
	members: []string{"public", "private", "mixed"},
}

var ImpactDomain = StringEnum{
	members: []string{"revenue", "compliance", "data", "security"},
}

var ArchitectureRole = StringEnum{
	members: []string{"entrypoint", "application", "infrastructure", "data"},
}

var InteractionType = StringEnum{
	members: []string{"data", "security", "performance", "async", "config"},
}
