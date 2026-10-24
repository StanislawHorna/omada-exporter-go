package Utils

// millisecondThreshold is the smallest value treated as a millisecond epoch.
// A timestamp in seconds stays below it until the year 5138, while millisecond
// timestamps have exceeded it since 2001, so the two scales cannot be confused
// for any clock that is anywhere near correct.
const millisecondThreshold float64 = 1e11

// UnixSeconds normalises an Omada "last seen" timestamp to Unix seconds.
//
// The Open API documents the device lastSeen value as a millisecond epoch
// ("Last online timestamp (ms)") and that is what the controller returns, e.g.
// 1792833739702. Metrics are exposed in seconds because that is the Prometheus
// convention and what timestamp arithmetic (time() - device_last_seen) and
// Grafana/alerting expect. A value that is already in seconds is passed through
// unchanged, so an endpoint that reports seconds does not surface as 1970.
func UnixSeconds(lastSeen float64) float64 {
	if lastSeen >= millisecondThreshold {
		return lastSeen / 1000
	}
	return lastSeen
}
