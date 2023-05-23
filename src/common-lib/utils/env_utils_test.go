package utils

import (
	"os"
	"testing"
)

func TestGetServiceName(t *testing.T) {
	testCases := []struct {
		desc  string
		value string
	}{
		{
			desc:  "default",
			value: "value",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			os.Setenv(ServiceNameEnv, tC.value)
			if got := GetServiceName(); got != tC.value {
				t.Errorf("expected: %s got: %s", tC.value, got)
			}
		})
	}
}

func TestGetServiceVersion(t *testing.T) {
	testCases := []struct {
		desc  string
		value string
	}{
		{
			desc:  "default",
			value: "1.0.0",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			os.Setenv(ServiceVersionEnv, tC.value)
			if got := GetServiceVersion(); got != tC.value {
				t.Errorf("expected: %s got: %s", tC.value, got)
			}
		})
	}
}

func TestGetEnvVar(t *testing.T) {
	testCases := []struct {
		desc  string
		prep  func()
		key   string
		value string
	}{
		{
			desc: "default",
			prep: func() {
				os.Setenv("TEST_ENV_VAR", "value")
			},
			key:   "TEST_ENV_VAR",
			value: "value",
		},
		{
			desc:  "not found",
			prep:  func() {},
			key:   "TEST_ENV_VAR_NOTFOUND",
			value: "",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			tC.prep()
			if got := GetEnvVar(tC.key, ""); got != tC.value {
				t.Errorf("expected: %s got: %s", tC.value, got)
			}
		})
	}
}
