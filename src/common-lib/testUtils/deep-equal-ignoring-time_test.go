package testutils

import (
	"testing"
	"time"
)

func TestDeepEqual(t *testing.T) {
	type testCase struct {
		name  string
		a     interface{}
		b     interface{}
		equal bool
	}

	type structWithDate struct {
		s string
		d time.Time
	}

	type structWithNestedDate struct {
		d structWithDate
		s string
	}

	tcs := []testCase{
		testCase{
			name:  "strings equal",
			a:     "test",
			b:     "test",
			equal: true,
		},
		testCase{
			name:  "strings unequal",
			a:     "test",
			b:     "test2",
			equal: false,
		},
		testCase{
			name: "structs with dates",
			a: structWithDate{
				s: "test",
				d: time.Now().UTC(),
			},
			b: structWithDate{
				s: "test",
				d: time.Now().UTC().Add(time.Hour),
			},
			equal: true,
		},
		testCase{
			name: "structs with structs with dates",
			a: structWithNestedDate{
				d: structWithDate{
					s: "test",
					d: time.Now().UTC(),
				},
				s: "test",
			},
			b: structWithNestedDate{
				d: structWithDate{
					s: "test",
					d: time.Now().UTC().Add(time.Hour),
				},
				s: "test",
			},
			equal: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			want := tc.equal
			got := DeepEqualIgnoringTime(tc.a, tc.b)
			if want != got {
				t.Errorf("want %t, but got %t", want, got)
			}
		})
	}
}
