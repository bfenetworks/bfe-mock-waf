// Copyright (c) 2019 The BFE Authors.
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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type ResponseBody struct {
	EventID    string `json:"event_id"`
	ResultFlag int    `json:"result_flag"`
}

type HCResponseBody struct {
	HcFlag int    `json:"result_flag"`
	Msg    string `json:"msg"`
}

var rand1 *rand.Rand
var rand2 *rand.Rand

func handler(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodPost {
	// 	http.Error(w, "Only support POST", http.StatusMethodNotAllowed)
	// 	return
	// }

	eventId := fmt.Sprintf("event_id_%d", time.Now().Unix())
	fmt.Println("header list:")
	for key, values := range r.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
		if key == "XInner-LogId" {
			eventId = "event_id_" + strings.Join(values, "_")
		}
	}

	// body, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	http.Error(w, "failed to read request body", http.StatusInternalServerError)
	// 	return
	// }
	// defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	flag := 0
	n := rand1.Intn(1000)
	if n < 5 {
		flag = 1
		fmt.Printf("http req attack eventId:%s\n", eventId)
	}

	resp := ResponseBody{
		EventID:    eventId,
		ResultFlag: flag,
	}

	json.NewEncoder(w).Encode(resp)
}

func hchandler(w http.ResponseWriter, r *http.Request) {
	n := rand2.Intn(1000)
	msg := "Succ"
	hcFlag := 0

	if n < 10 {
		msg = "server unavailable"
		hcFlag = 1
	}
	resp := HCResponseBody{
		HcFlag: hcFlag,
		Msg:    msg,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	source1 := rand.NewSource(time.Now().UnixNano())
	rand1 = rand.New(source1)

	source2 := rand.NewSource(time.Now().UnixNano() + int64(os.Getpid()))
	rand2 = rand.New(source2)

	port := flag.String("port", "8899", "WAF HTTP server listening port")
	flag.Parse()

	fmt.Printf("WAF HTTP server listening port:%s\n", *port)

	http.HandleFunc("/detect", handler)
	http.HandleFunc("/hccheck", hchandler)

	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start server: %v\n", err)
		os.Exit(1)
	}
}
