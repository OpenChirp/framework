package rest

import (
	"encoding/json"
	"net/http"
)

// HealthStatus represents the result of a health check
type HealthStatus string

const (
	// HealthStatusUnknown indicates that the health status was not reported
	HealthStatusUnknown HealthStatus = ""
	// HealthStatusOK indicates that the server reported the status to be ok
	HealthStatusOK HealthStatus = "ok"
	// HealthStatusDegraded indicates that the server reported the status
	// to be degraded
	HealthStatusDegraded HealthStatus = "degraded"
)

// HealthCheckResponse encapsulates the response from the server health check
type HealthCheckResponse struct {
	Status HealthStatus `json:"status"`
}

// HealthCheck requests the health of the rest server
func (host Host) HealthCheck() (HealthStatus, error) {
	var checkStatus HealthCheckResponse

	uri := host.uri + healthCheckSubPath
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return HealthStatusUnknown, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := host.client.Do(req)
	if err != nil {
		// should report auth problems here in future
		return HealthStatusUnknown, err
	}
	defer resp.Body.Close()
	if err := DecodeOCError(resp); err != nil {
		return HealthStatusUnknown, err
	}
	err = json.NewDecoder(resp.Body).Decode(&checkStatus)
	return checkStatus.Status, err
}
