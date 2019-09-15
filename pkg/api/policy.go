package api

import (
	"github.com/gradecak/watchdog/pkg/policy"
)

type Policy struct {
	policy policy.Policy
}

func NewPolicyAPI(p policy.Policy) *Policy {
	return &Policy{p}
}

func (p *Policy) GetEnforcers(prefix string) (error, []policy.Enforcer) {
	return p.policy.Actions(prefix)
}
