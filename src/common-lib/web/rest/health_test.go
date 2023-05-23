package rest

import (
	"fmt"
	"net/http"
	"testing"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/runtime/logger"
)

const listenURL = ":12124"

func TestRegistryHealth(t *testing.T) {
	logger.Update(logger.Config{Destination: logger.DISCARD})
	const status = "testing"

	oldHealthData := healthData
	defer func() { healthData = oldHealthData }()

	RegistryHealth(&Health{Status: status})
	if healthData.(*Health).Status != status {
		t.Errorf("expected status %s, but got %s", status, healthData.(*Health).Status)
	}
}

func TestHealth(t *testing.T) {
	logger.Update(logger.Config{Destination: logger.DISCARD})
	h := &Health{ListenURL: listenURL}
	if err := h.health(); err != nil {
		t.Error(err)
	}

	if h.Type != healthType {
		t.Errorf("expected Type %s, but got %s", healthType, h.Type)
	}

	if h.LastStartTimeUTC != TimeStampUTC {
		t.Errorf("expected LastStartTimeUTC %s, but got %s", TimeStampUTC, h.LastStartTimeUTC)
	}

	if len(h.OutboundConnectionStatus) != 0 {
		t.Error("OutboundConnectionStatus must be empty")
	}

	if h.Status != healthRunningStatus {
		t.Errorf("expected Status %s, but got %s", healthRunningStatus, h.Status)
	}

	h.ListenURL = ""
	if err := h.health(); err == nil {
		t.Error("Error can not be <nil>")
	}
}

func TestGetHealthStatus(t *testing.T) {
	logger.Update(logger.Config{Destination: logger.DISCARD})
	var conns = []OutboundConnectionStatus{
		{ConnectionStatus: ConnectionStatusUnavailable},
	}
	if status := GetHealthStatus(conns); status != healthDegradedStatus {
		t.Errorf("expected Status %s, but got %s", healthDegradedStatus, status)
	}
}

type mockStatuser struct{}

func (mockStatuser) Status(status OutboundConnectionStatus) *OutboundConnectionStatus {
	return &status
}

func TestGetOutboundConnectionStatus(t *testing.T) {
	logger.Update(logger.Config{Destination: logger.DISCARD})
	const serviceName = "testServiceName"
	var methods = []Statuser{mockStatuser{}}
	connections := GetOutboundConnectionStatus(methods, serviceName)
	name := fmt.Sprintf("%s-%s", serviceName, dbNameSuffix)
	if connections[0].Name != name {
		t.Errorf("expected Name %q, but got %q", name, connections[0].Name)
	}
}

func TestHandlerHealth_OK(t *testing.T) {
	logger.Update(logger.Config{Destination: logger.DISCARD})
	oldHealthData := healthData
	healthData = &Health{ListenURL: listenURL}
	defer func() { healthData = oldHealthData }()
	healthLastTimeSuccess = 0

	mock := &mockResponseWriter{dataHeader: http.Header{}}
	HandlerHealth(mock, nil)
	if mock.dataWriteHeader != http.StatusOK {
		t.Errorf("expected code %d, but got %d", http.StatusOK, mock.dataWriteHeader)
	}
}

func TestHandlerHealth_InternalServerError(t *testing.T) {
	logger.Update(logger.Config{Destination: logger.DISCARD})
	oldHealthData := healthData
	healthData = &Health{}
	defer func() { healthData = oldHealthData }()
	healthLastTimeSuccess = 0

	mock := &mockResponseWriter{dataHeader: http.Header{}}
	HandlerHealth(mock, nil)
	if mock.dataWriteHeader != http.StatusInternalServerError {
		t.Errorf("expected code %d, but got %d", http.StatusInternalServerError, mock.dataWriteHeader)
	}
}

func TestHandlerHealth_TwoInternalServerErrorsConcurrently(t *testing.T) {
	logger.Update(logger.Config{Destination: logger.DISCARD})
	oldHealthData := healthData
	// The healthData won't be ready until *after* a send/close occurs on this channel
	sendHealthDataCh := make(chan bool, 2)
	healthData = &Health{
		ListenURL: listenURL,
		HealthCode: func(h *Health) int {
			<-sendHealthDataCh
			return 500
		},
	}
	defer func() { healthData = oldHealthData }()

	// When you try to check the health, after a response is sent, this channel will be sent on
	receivedHealthDataCh := make(chan bool)
	checkHealth := func() {
		mock := &mockResponseWriter{dataHeader: http.Header{}}
		HandlerHealth(mock, nil)
		if mock.dataWriteHeader != http.StatusInternalServerError {
			t.Errorf("expected code %d, but got %d", http.StatusInternalServerError, mock.dataWriteHeader)
		}
		receivedHealthDataCh <- true
	}

	// The idea is that you fire off 2 concurrent requests to get a health check, and it takes time, so one of them finishes first.
	healthLastTimeSuccess = 0
	go checkHealth()
	go checkHealth()
	// Finish one of the health checks, completely
	sendHealthDataCh <- true
	<-receivedHealthDataCh
	// Now let the other one finish
	sendHealthDataCh <- true
	<-receivedHealthDataCh
}
