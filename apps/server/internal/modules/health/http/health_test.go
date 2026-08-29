package healthhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	platformhealth "github.com/complexus-tech/projects-api/internal/platform/health"
	"github.com/stretchr/testify/require"
)

type readinessStub struct {
	report platformhealth.Report
}

func (s readinessStub) Report(context.Context) platformhealth.Report {
	return s.report
}

func TestReadinessStatusReflectsSupervisorReport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		report     platformhealth.Report
		wantStatus int
	}{
		{
			name: "ready",
			report: platformhealth.Report{
				Status: "ready",
				Phase:  platformhealth.PhaseReady,
				Checks: map[string]string{"postgres": "ready", "redis": "ready"},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "draining",
			report: platformhealth.Report{
				Status: "not_ready",
				Phase:  platformhealth.PhaseDraining,
				Checks: map[string]string{"postgres": "not_checked", "redis": "not_checked"},
			},
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := New(nil, readinessStub{report: tt.report})
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/readiness", nil)

			err := handler.Readiness(request.Context(), response, request)
			require.NoError(t, err)
			require.Equal(t, tt.wantStatus, response.Code)

			var envelope struct {
				Data struct {
					Status string               `json:"status"`
					Phase  platformhealth.Phase `json:"phase"`
					Checks map[string]string    `json:"checks"`
				} `json:"data"`
			}
			require.NoError(t, json.NewDecoder(response.Body).Decode(&envelope))
			wantStatus := tt.report.Status
			if wantStatus == "ready" {
				wantStatus = "ok"
			}
			require.Equal(t, wantStatus, envelope.Data.Status)
			require.Equal(t, tt.report.Phase, envelope.Data.Phase)
			require.Equal(t, tt.report.Checks, envelope.Data.Checks)
		})
	}
}

func TestReadinessFailsClosedWithoutReporter(t *testing.T) {
	t.Parallel()

	handler := New(nil, nil)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readiness", nil)

	require.NoError(t, handler.Readiness(request.Context(), response, request))
	require.Equal(t, http.StatusServiceUnavailable, response.Code)
}

func TestLivenessPreservesInfrastructureProbeContract(t *testing.T) {
	t.Parallel()

	handler := New(nil, nil)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/liveness", nil)

	require.NoError(t, handler.Liveness(request.Context(), response, request))
	require.Equal(t, http.StatusOK, response.Code)

	var envelope struct {
		Data struct {
			Status     string `json:"status"`
			Hostname   string `json:"hostname"`
			GOMAXPROCS int    `json:"GOMAXPROCS"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&envelope))
	require.Equal(t, "ok", envelope.Data.Status)
	require.NotEmpty(t, envelope.Data.Hostname)
	require.Positive(t, envelope.Data.GOMAXPROCS)
}
