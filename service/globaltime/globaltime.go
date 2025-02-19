package globaltime

import "time"

var FixedTime time.Time

func Now() time.Time {
	if FixedTime.After(time.Time{}) {
		return FixedTime
	}
	return time.Now()
}

func Since(tm time.Time) time.Duration {
	return Now().Sub(tm)
}
