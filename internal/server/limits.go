package server

import "github.com/stockyard-dev/stockyard-lasso/internal/license"

type Limits struct {
	MaxLinks      int  // 0 = unlimited
	ClickTracking bool // detailed click analytics
	CustomSlugs   bool // vanity URLs
	PasswordLinks bool // password-protected redirects
	RetentionDays int
}

var freeLimits = Limits{
	MaxLinks:      25,
	ClickTracking: true, // free hook
	CustomSlugs:   true, // free hook
	PasswordLinks: false,
	RetentionDays: 7,
}

var proLimits = Limits{
	MaxLinks:      0,
	ClickTracking: true,
	CustomSlugs:   true,
	PasswordLinks: true,
	RetentionDays: 365,
}

func LimitsFor(info *license.Info) Limits {
	if info != nil && info.IsPro() {
		return proLimits
	}
	return freeLimits
}

func LimitReached(limit, current int) bool {
	if limit == 0 {
		return false
	}
	return current >= limit
}
