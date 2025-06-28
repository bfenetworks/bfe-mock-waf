// Copyright (c) 2025 The BFE Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package waf_bfe_sdk

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/bfenetworks/bwi/bwi"
)

const DefaultTimeOut time.Duration = 10 * time.Second
const DefaultDetectURL = "http://127.0.0.1:8899/detect"
const DefaultHCURL = "http://127.0.0.1:8899/hccheck"

type ResponseBody struct {
	EventID    string `json:"event_id"`
	ResultFlag int    `json:"result_flag"`
}

type HCResponseBody struct {
	HcFlag int    `json:"result_flag"`
	Msg    string `json:"msg"`
}

type MockWafResult struct {
	eventId    string
	resultFlag int
}

func (r *MockWafResult) GetEventId() string {
	return r.eventId
}

func (r *MockWafResult) GetResultFlag() int {
	if r.resultFlag == 0 {
		return bwi.WAF_RESULT_PASS
	} else {
		return bwi.WAF_RESULT_BLOCK
	}
}

type MockWafServerAgent struct {
	client    *http.Client
	closeFlag bool
}

func (c *MockWafServerAgent) Close() {
	c.client.CloseIdleConnections()
	c.closeFlag = true
}

func (c *MockWafServerAgent) DetectRequest(req *http.Request, logId string) (bwi.WafResult, error) {
	if c.closeFlag {
		return nil, errors.New("MockWafServerAgent has been closed")
	}

	//construct waf server request by req
	body, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", DefaultDetectURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Host = req.Host
	request.Header.Add("Host", req.Host)
	lenStr := fmt.Sprintf("%d", len(body))
	request.Header.Add("Content-Length", lenStr)
	request.Header.Add("XInner-LogId", logId)
	//...add other header...

	//access waf server
	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	//construct result
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("non-200 status code received: " + resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var res ResponseBody
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	wafResult := &MockWafResult{
		eventId:    res.EventID,
		resultFlag: res.ResultFlag,
	}

	return wafResult, err
}

func (c *MockWafServerAgent) UpdateSockFactory(socketFactory func() (net.Conn, error)) {
	if c.closeFlag {
		return
	}
	if transport, ok := c.client.Transport.(*http.Transport); ok {
		transport.DialContext = func(ctx context.Context, network string, addr string) (net.Conn, error) {
			return socketFactory()
		}
	}
}

func NewWafServerWithPoolSize(socketFactory func() (net.Conn, error), poolSize int) bwi.WafServer {
	/*
		dial := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}

	*/
	client := MockWafServerAgent{
		client: &http.Client{
			Timeout: time.Duration(DefaultTimeOut),
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
					return socketFactory()
				},
				ForceAttemptHTTP2:     true,
				MaxConnsPerHost:       poolSize,
				MaxIdleConnsPerHost:   poolSize,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		closeFlag: false,
	}
	return &client
}

func HealthCheck(conn net.Conn) error {
	//construct health check request
	request := "GET /hccheck HTTP/1.1\r\nHost: " + conn.RemoteAddr().String() + "\r\nConnection: close\r\n\r\n"

	//access waf server
	_, err := conn.Write([]byte(request))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	//read & construct result
	response, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}
	defer response.Body.Close()

	//check
	if response.StatusCode != http.StatusOK {
		return errors.New("non-200 status code received: " + response.Status)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var hcRes HCResponseBody
	if err := json.Unmarshal(bodyBytes, &hcRes); err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if hcRes.HcFlag == 0 {
		return nil
	} else {
		return fmt.Errorf("waf server is not available. flag:%d", hcRes.HcFlag)
	}
}
