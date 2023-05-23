package logger

import (
	"time"

	"go.uber.org/zap/zapcore"
)

// https://confluence.kksharmadevdev.com/pages/viewpage.action?spaceKey=PDStandards&title=Standard+Log+Schema

// Error represent error information from the event
type Error struct {
	ID         string `json:"Id,omitempty"`
	Message    string `json:"Message,omitempty"`
	Type       string `json:"Type,omitempty"`
	StackTrace string `json:"StackTrace,omitempty"`
}

// MarshalLogObject Marshal Error to zap Object
func (e Error) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if e.ID != "" {
		enc.AddString("ID", e.ID)
	}
	if e.Message != "" {
		enc.AddString("Message", e.Message)
	}
	if e.Type != "" {
		enc.AddString("Type", e.Type)
	}
	if e.StackTrace != "" {
		enc.AddString("StackTrace", e.StackTrace)
	}
	return nil
}

// Request represent request information from the event
type Request struct {
	Body                  interface{}       `json:"Body,omitempty"`
	Domain                string            `json:"Domain,omitempty"`
	Headers               map[string]string `json:"Headers,omitempty"`
	Method                string            `json:"Method,omitempty"`
	Path                  string            `json:"Path,omitempty"`
	QueryStringParameters map[string]string `json:"QueryStringParameters,omitempty"`
	RemoteAddress         string            `json:"RemoteAddress,omitempty"`
	URL                   string            `json:"Url,omitempty"`
}

// MarshalLogObject Marshal Request to zap Object
func (r Request) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if r.Body != nil {
		bodyOM, implementsOM := interface{}(r.Body).(zapcore.ObjectMarshaler)
		if implementsOM {
			enc.AddObject("Body", bodyOM)
		} else {
			enc.AddReflected("Body", r.Body)
		}
	}
	if r.Domain != "" {
		enc.AddString("Domain", r.Domain)
	}
	if r.Headers != nil {
		for k, v := range r.Headers {
			enc.AddString(k, v)
		}
	}
	if r.Method != "" {
		enc.AddString("Method", r.Method)
	}
	if r.Path != "" {
		enc.AddString("Path", r.Path)
	}
	if r.QueryStringParameters != nil {
		for k, v := range r.QueryStringParameters {
			enc.AddString(k, v)
		}
	}
	if r.RemoteAddress != "" {
		enc.AddString("RemoteAddress", r.RemoteAddress)
	}
	if r.URL != "" {
		enc.AddString("Url", r.URL)
	}
	return nil
}

// Response represent response information from the event
type Response struct {
	Body       interface{}       `json:"Body,omitempty"`
	Headers    map[string]string `json:"Headers,omitempty"`
	StatusCode int               `json:"StatusCode,omitempty"`
}

// MarshalLogObject Marshal Response to zap Object
func (r Response) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if r.Body != nil {
		bodyOM, implementsOM := interface{}(r.Body).(zapcore.ObjectMarshaler)
		if implementsOM {
			enc.AddObject("Body", bodyOM)
		} else {
			enc.AddReflected("Body", r.Body)
		}
	}
	if r.Headers != nil {
		for k, v := range r.Headers {
			enc.AddString(k, v)
		}
	}
	enc.AddInt("StatusCode", r.StatusCode)
	return nil
}

// Event represent log information
type Event struct {
	Date     time.Time `json:"Date,omitempty"`
	Source   string    `json:"Source,omitempty"`
	LogLevel string    `json:"Type,omitempty"`
	Message  string    `json:"Message,omitempty"`

	ID          string `json:"Id,omitempty"`
	Description string `json:"Description,omitempty"`
	Duration    int    `json:"Duration,omitempty"`
	// Response    *Response `json:"Response,omitempty"`
	// Request     *Request               `json:"Request,omitempty"`
	Error *Error                  `json:"Error,omitempty"`
	Data  map[string]interface{}  `json:"Data,omitempty"`
	Audit zapcore.ObjectMarshaler `json:"Audit,omitempty"`
}

// MarshalLogObject Marshal Event to zap Object
func (e Event) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddTime("Date", e.Date)
	if e.Source != "" {
		enc.AddString("Source", e.Source)
	}
	if e.LogLevel != "" {
		enc.AddString("Type", e.LogLevel)
	}
	if e.Message != "" {
		enc.AddString("Message", e.Message)
	}
	if e.ID != "" {
		enc.AddString("Id", e.ID)
	}
	if e.Description != "" {
		enc.AddString("Description", e.Description)
	}
	if e.Duration != 0 {
		enc.AddInt("Duration", e.Duration)
	}
	// if e.Response != nil {
	// 	enc.AddObject("Response", e.Response)
	// }
	// if e.Request != nil {
	// 	enc.AddObject("Request", e.Request)
	// }
	if e.Error != nil {
		enc.AddObject("Error", e.Error)
	}
	if e.Data != nil {
		for key, dataValue := range e.Data {
			dataOM, implementsOM := interface{}(dataValue).(zapcore.ObjectMarshaler)
			if implementsOM {
				enc.AddObject(key, dataOM)
			} else {
				enc.AddReflected(key, dataValue)
			}
		}
	}
	if e.Audit != nil {
		enc.AddObject("Audit", e.Audit)
	}
	return nil
}

// Resource represent resource information from the event
type Resource struct {
	ClientID      string `json:"ClientID,omitempty"`
	PartnerID     string `json:"PartnerID,omitempty"`
	CorrelationID string `json:"CorrelationID,omitempty"`
	RequestID     string `json:"RequestID,omitempty"`
	UserID        string `json:"UserID,omitempty"`
	CompanyID     string `json:"CompanyID,omitempty"`
	SiteID        string `json:"SiteID,omitempty"`
	AgentID       string `json:"AgentID,omitempty"`
	EndpointID    string `json:"EndpointID,omitempty"`
}

// MarshalLogObject Marshal Resource to zap Object
func (r Resource) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if r.ClientID != "" {
		enc.AddString("ClientId", r.ClientID)
	}
	if r.PartnerID != "" {
		enc.AddString("PartnerId", r.PartnerID)
	}
	if r.CorrelationID != "" {
		enc.AddString("CorrelationId", r.CorrelationID)
	}
	if r.UserID != "" {
		enc.AddString("UserId", r.UserID)
	}
	if r.CompanyID != "" {
		enc.AddString("CompanyId", r.CompanyID)
	}
	if r.SiteID != "" {
		enc.AddString("SiteId", r.SiteID)
	}
	if r.AgentID != "" {
		enc.AddString("AgentId", r.AgentID)
	}
	if r.EndpointID != "" {
		enc.AddString("EndpointId", r.EndpointID)
	}
	if r.RequestID != "" {
		enc.AddString("RequestId", r.RequestID)
	}
	return nil
}

// Host represent host information from the event
type Host struct {
	Name string `json:"HostName,omitempty"`
}

// MarshalLogObject Marshal Host to zap Object
func (h Host) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if h.Name != "" {
		enc.AddString("HostName", h.Name)
	}
	return nil
}

// Service represent service information from the event
type Service struct {
	Name         string `json:"Name,omitempty"`
	Version      string `json:"Version,omitempty"`
	Owner        string `json:"Owner,omitempty"`
	BuildNumber  string `json:"BuildNumber,omitempty"`
	CommitSHA    string `json:"CommitId,omitempty"`
	BuildVersion string `json:"BuildVersion,omitempty"`
	Environment  string `json:"Environment,omitempty"`
}

// MarshalLogObject Marshal Service to zap Object
func (s Service) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if s.Name != "" {
		enc.AddString("Name", s.Name)
	}
	if s.Version != "" {
		enc.AddString("Version", s.Version)
	}
	if s.Owner != "" {
		enc.AddString("Owner", s.Owner)
	}
	if s.BuildNumber != "" {
		enc.AddString("BuildNumber", s.BuildNumber)
	}
	if s.CommitSHA != "" {
		enc.AddString("CommitId", s.CommitSHA)
	}
	if s.BuildVersion != "" {
		enc.AddString("BuildVersion", s.BuildVersion)
	}
	if s.Environment != "" {
		enc.AddString("Environment", s.Environment)
	}
	return nil
}

// LogContent consolidated log information
type LogContent struct {
	Event    Event    `json:"Event,omitempty"`
	Resource Resource `json:"Resource,omitempty"`
}

// MarshalLogObject Marshal LogContent to zap Object
func (l LogContent) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddObject("Event", l.Event)
	enc.AddObject("Resource", l.Resource)
	return nil
}

// LogReponse represent final log response
type LogReponse struct {
	Event    Event    `json:"Event,omitempty"`
	Host     Host     `json:"Host,omitempty"`
	Service  Service  `json:"Service,omitempty"`
	Resource Resource `json:"Resource,omitempty"`
}

// MarshalLogObject Marshal LogResponse to zap Object
func (l LogReponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddObject("Event", l.Event)
	enc.AddObject("Host", l.Host)
	enc.AddObject("Service", l.Service)
	enc.AddObject("Resource", l.Resource)
	return nil
}
