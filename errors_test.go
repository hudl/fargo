package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"errors"
	"testing"
)

func TestHTTPResponseStatusCode(t *testing.T) {
	tests := []struct {
		input   error
		present bool
		code    int
	}{
		{nil, false, 0},
		{errors.New("other"), false, 0},
		{&unsuccessfulHTTPResponse{404, "missing"}, true, 404},
	}
	for _, test := range tests {
		code, present := HTTPResponseStatusCode(test.input)
		if present {
			if !test.present {
				t.Errorf("input %v: want absent, got code %d", test.input, code)
				continue
			}
			if code != test.code {
				t.Errorf("input %v: want %d, got %d", test.input, test.code, code)
			}
		} else if test.present {
			t.Errorf("input %v: want code %d, got absent", test.input, test.code)
		}
	}
}
