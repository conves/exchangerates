package main

import "fmt"

type recommendation struct {
	today             string
	value             float64
	sevenDaysAgo      string
	sevenDaysAgoValue float64
	msg               string
}

// String makes recommendation implement Stringer interface
func (r recommendation) String() string {
	return fmt.Sprintf("For today's rate of %f and last week's historic rate of %f, we recommend %s.",
		r.value, r.sevenDaysAgoValue, r.msg)
}

func newRecommendation(today apiResp, sevenDaysAgo apiResp) recommendation {
	rec := recommendation{
		today:             today.Date,
		value:             today.Rates.GBP,
		sevenDaysAgo:      sevenDaysAgo.Date,
		sevenDaysAgoValue: sevenDaysAgo.Rates.GBP,
	}
	if today.Rates.GBP > sevenDaysAgo.Rates.GBP {
		rec.msg = "selling"
	} else if today.Rates.GBP < sevenDaysAgo.Rates.GBP {
		rec.msg = "buying"
	} else {
		rec.msg = "doing nothing and wait for another time..."
	}
	return rec
}
