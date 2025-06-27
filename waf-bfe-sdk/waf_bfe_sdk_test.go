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
	"bytes"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"
)

var (
	dtServerAddr   string        = "127.0.0.1:8899"
	hcServerAddr   string        = "127.0.0.1:8899"
	connectTimeout time.Duration = time.Duration(1000 * time.Millisecond)
	poolSize       int           = 10
)

func TestBfeWafSdkCase1(t *testing.T) {
	detectSockFunc := func() (net.Conn, error) {
		conn, err := net.DialTimeout("tcp", dtServerAddr, connectTimeout)
		if err != nil {
			fmt.Printf("failed to dial to %s\n", dtServerAddr)
		}
		return conn, err
	}

	for idx := 0; idx < 10; idx++ {
		hcconn, err := net.DialTimeout("tcp", dtServerAddr, connectTimeout)
		if err != nil {
			fmt.Printf("failed to dial to hc waf server:%s\n", dtServerAddr)
		}
		err = HealthCheck(hcconn)
		fmt.Printf("HealthCheck result: %+v\n", err)
	}

	server := NewWafServerWithPoolSize(detectSockFunc, poolSize)

	url := "https://api.example.com/endpoint"

	body := "abcdefg"
	bodyBytes := bytes.NewBuffer([]byte(body))

	req, err := http.NewRequest("POST", url, bodyBytes)
	if err != nil {
		panic(fmt.Sprintf("failed to create: %v", err))
	}
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "MyGoClient/1.0")
	req.Header.Set("Authorization", "Bearer xyz123")
	req.Header.Set("XInner-LogId", "123e4567-e89b-12d3-a456-426655440000")
	req.Header.Set("Custom-Header", "my-value")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))

	cookie := &http.Cookie{
		Name:  "session_id",
		Value: "abc123xyz",
		Path:  "/",
	}
	req.AddCookie(cookie)

	for idx := 0; idx < 10; idx++ {
		logId := fmt.Sprintf("logId-%d-%d", idx, time.Now().UnixNano())
		wafRes, err := server.DetectRequest(req, logId)
		fmt.Printf("DetectRequest result: %+v, %+v\n", wafRes, err)
	}
}
