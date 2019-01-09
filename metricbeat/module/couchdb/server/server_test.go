// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	mbtest "github.com/elastic/beats/metricbeat/mb/testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchEventContent(t *testing.T) {
	absPath, err := filepath.Abs("../_meta/test/")
	assert.NoError(t, err)

	response, err := ioutil.ReadFile(absPath + "/serverstats.json")
	assert.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json;")
		w.Write([]byte(response))
	}))
	defer server.Close()

	config := map[string]interface{}{
		"module":     "couchdb",
		"metricsets": []string{"server"},
		"hosts":      []string{server.URL},
	}
	f := mbtest.NewEventFetcher(t, config)
	event, err := f.Fetch()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s/%s event: %+v", f.Module().Name(), f.Name(), event)
}

func TestFetchTimeout(t *testing.T) {
	absPath, err := filepath.Abs("../_meta/test/")
	assert.NoError(t, err)

	response, err := ioutil.ReadFile(absPath + "/serverstats.json")
	assert.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json;")
		w.Write([]byte(response))
		<-r.Context().Done()
	}))
	defer server.Close()

	config := map[string]interface{}{
		"module":     "couchdb",
		"metricsets": []string{"server"},
		"hosts":      []string{server.URL},
		"timeout":    "50ms",
	}

	f := mbtest.NewEventFetcher(t, config)

	start := time.Now()
	_, err = f.Fetch()
	elapsed := time.Since(start)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "request canceled (Client.Timeout exceeded")
	}

	assert.True(t, elapsed < 5*time.Second, "elapsed time: %s", elapsed.String())
}
