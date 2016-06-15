package fargo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type roundtripper struct {
	TripCount int
}

func (r *roundtripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r.TripCount++
	return http.DefaultTransport.RoundTrip(req)
}

func TestHttpClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World")
	}))
	defer server.Close()

	Convey("Given fargo.HttpClient is set to a custom client", t, func() {
		rt := new(roundtripper)
		HttpClient = &http.Client{
			Transport: rt,
		}

		Convey("netReq uses that client to handle requests", func() {
			req, err := http.NewRequest("GET", server.URL, nil)
			So(err, ShouldBeNil)

			respBody, respCode, err := netReq(req)
			So(err, ShouldBeNil)
			So(respCode, ShouldEqual, 200)
			So(string(respBody), ShouldEqual, "Hello World")

			So(rt.TripCount, ShouldEqual, 1)
		})
	})
}
