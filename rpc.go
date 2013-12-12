package fargo

/*
 * The MIT License (MIT)
 *
 * Copyright (c) 2013 Ryan S. Brown <sb@ryansb.com>
 * Copyright (c) 2013 Hudl <@Hudl>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to
 * deal in the Software without restriction, including without limitation the
 * rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
 * sell copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 * FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
 * IN THE SOFTWARE.
 */

import (
	"bytes"
	"github.com/mreiferson/go-httpclient"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

func postXML(reqUrl string, reqBody []byte) ([]byte, int, error) {
	req, err := http.NewRequest("POST", reqUrl, bytes.NewReader(reqBody))
	if err != nil {
		log.Error("Could not create POST %s with body %s Error: %s", reqUrl, string(reqBody), err.Error())
		return nil, -1, err
	}
	body, rcode, err := reqXML(req)
	if err != nil {
		log.Error("Could not complete POST %s with body %s Error: %s", reqUrl, string(reqBody), err.Error())
		return nil, rcode, err
	}
	//eurekaCache.Flush()
	return body, rcode, nil
}

func putKV(reqUrl string, pairs map[string]string) ([]byte, int, error) {
	params := url.Values{}
	for k, v := range pairs {
		params.Add(k, v)
	}
	parameterizedUrl := reqUrl + "?" + params.Encode()
	log.Notice("Sending metadata request with URL %s", parameterizedUrl)
	req, err := http.NewRequest("PUT", parameterizedUrl, nil)
	if err != nil {
		log.Error("Could not create PUT %s with Error: %s", reqUrl, err.Error())
		return nil, -1, err
	}
	body, rcode, err := reqXML(req)
	if err != nil {
		log.Error("Could not complete PUT %s with Error: %s", reqUrl, err.Error())
		return nil, rcode, err
	}
	return body, rcode, nil
}

func getXML(reqUrl string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		log.Error("Could not create GET %s with Error: %s", reqUrl, err.Error())
		return nil, -1, err
	}
	body, rcode, err := reqXML(req)
	if err != nil {
		log.Error("Could not complete GET %s with Error: %s", reqUrl, err.Error())
		return nil, rcode, err
	}
	return body, rcode, nil
}

func reqXML(req *http.Request) ([]byte, int, error) {
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Accept", "application/xml")
	return netReq(req)
}

func netReq(req *http.Request) ([]byte, int, error) {
	transport := &httpclient.Transport{
		ConnectTimeout:        5 * time.Second,
		RequestTimeout:        30 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}
	defer transport.Close()

	// Send the request via a client
	client := &http.Client{Transport: transport}
	var resp *http.Response
	var err error
	for i := 0; i < 3; i++ {
		resp, err = client.Do(req)
		if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
			// it's a transient network error so we sleep for a bit and try
			// again in case it's a short-lived issue
			log.Warning("Retrying after temporary network failure, error: %s",
				nerr.Error())
			time.Sleep(10)
		}
	}
	if err != nil {
		return nil, -1, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failure reading request body Error: %s", err.Error())
		return nil, -1, err
	}
	// At this point we're done and shit worked, simply return the bytes
	return body, resp.StatusCode, nil
}
