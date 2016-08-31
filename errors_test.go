package fargo

// MIT Licensed (see README.md) - Copyright (c) 2013 Hudl <@Hudl>

import (
	"errors"
	"net/http"
	"sort"
	"testing"
)

func isMember(haystack []int, needle int) bool {
	i := sort.SearchInts(haystack, needle)
	return i != len(haystack) && haystack[i] == needle
}

func predicateOnlyTrueFor(t *testing.T, pred func(error) bool, codesByOp map[httpOperation][]int) {
	makeError := func(op httpOperation, code int) error {
		return &unsuccessfulHTTPResponse{op, code, ""}
	}
	for op := opRegistration; op <= opRetrieval; op++ {
		if codes, ok := codesByOp[op]; ok {
			if !sort.IntsAreSorted(codes) {
				t.Fatalf("codes for op %d are not sorted: %d", op, codes)
			}
			for c := 100; c != 600; c++ {
				if got, want := pred(makeError(op, c)), isMember(codes, c); got != want {
					t.Errorf("op %d, code %d: got %t, want %t", op, c, got, want)
				}
			}
		} else {
			// None of the codes should provoke a true return value.
			for c := 100; c != 600; c++ {
				if pred(makeError(op, c)) {
					t.Errorf("op %d, code %d: got true, want false", op, c)
				}
			}
		}
	}

	if pred(errors.New("other")) {
		t.Errorf("non HTTP-related error: got true, want false")
	}
}

func TestInvalidInstanceDetection(t *testing.T) {
	predicateOnlyTrueFor(t, InstanceWasInvalid, map[httpOperation][]int{
		opRegistration: {http.StatusBadRequest},
	})
}

func TestMissingInstanceDetection(t *testing.T) {
	predicateOnlyTrueFor(t, InstanceWasMissing, map[httpOperation][]int{
		opDeregistration: {http.StatusNotFound},
		opMetadataUpdate: {http.StatusInternalServerError},
		opStatusUpdate:   {http.StatusNotFound},
		opLeaseRenewal:   {http.StatusNotFound},
		opRetrieval:      {http.StatusNotFound},
	})
}
