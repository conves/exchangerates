package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type HttpMockClient struct {
	wantedResponseForToday    http.Response
	wantedResponseForLastWeek http.Response
	nextResponse              int
}

func (c *HttpMockClient) Do(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.String(), time.Now().Format("2006-01-02")) {
		return &c.wantedResponseForToday, nil
	}
	return &c.wantedResponseForLastWeek, nil
}

func Test_recommendHandler(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	cache := NewImMemCache()
	defer cache.Close()

	cache.SetWithTTL(today, "", time.Duration(0)) // Invalidate Cache

	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/recommend", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Test 500 on the API
	{
		httpMockClient := &HttpMockClient{
			wantedResponseForToday:http.Response{
				StatusCode:       http.StatusInternalServerError,
				Body: ioutil.NopCloser(bytes.NewReader([]byte{})),
			},
			wantedResponseForLastWeek:http.Response{
				StatusCode:       http.StatusInternalServerError,
				Body: ioutil.NopCloser(bytes.NewReader([]byte{})),
			},
		}

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(recommendHandler(cache, httpMockClient))

		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		wantedStatus := http.StatusInternalServerError
		if status := rr.Code; status != wantedStatus {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, wantedStatus)
		}
	}

	// Test valid response on the API - "buying" case
	{
		apiRespForToday := apiResp{
			Rates: struct {
				GBP float64 `json:"GBP"`
			}{
				GBP: 0.8978,
			},
			Base: "",
			Date: "",
		}
		wantRespForToday, _ := json.Marshal(apiRespForToday)
		apiRespForLastWeek := apiResp{
			Rates: struct {
				GBP float64 `json:"GBP"`
			}{
				GBP: 0.9978,
			},
			Base: "",
			Date: "",
		}
		wantRespForLastWeek, _ := json.Marshal(apiRespForLastWeek)
		httpMockClient := &HttpMockClient{
			wantedResponseForToday:http.Response{
				StatusCode:       http.StatusOK,
				Body: ioutil.NopCloser(bytes.NewReader(wantRespForToday)),
			},
			wantedResponseForLastWeek:http.Response{
				StatusCode:       http.StatusOK,
				Body: ioutil.NopCloser(bytes.NewReader(wantRespForLastWeek)),
			},
		}

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(recommendHandler(cache, httpMockClient))

		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		wantedStatus := http.StatusOK
		if status := rr.Code; status != wantedStatus {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, wantedStatus)
		}

		expected := newRecommendation(apiRespForToday, apiRespForLastWeek).String()
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	}

	// Test valid response on the API - "selling" case
	{
		cache.SetWithTTL(today, "", time.Duration(0)) // Invalidate Cache

		apiRespForToday := apiResp{
			Rates: struct {
				GBP float64 `json:"GBP"`
			}{
				GBP: 0.9978,
			},
			Base: "",
			Date: "",
		}
		wantRespForToday, _ := json.Marshal(apiRespForToday)
		apiRespForLastWeek := apiResp{
			Rates: struct {
				GBP float64 `json:"GBP"`
			}{
				GBP: 0.8978,
			},
			Base: "",
			Date: "",
		}
		wantRespForLastWeek, _ := json.Marshal(apiRespForLastWeek)
		httpMockClient := &HttpMockClient{
			wantedResponseForToday:http.Response{
				StatusCode:       http.StatusOK,
				Body: ioutil.NopCloser(bytes.NewReader(wantRespForToday)),
			},
			wantedResponseForLastWeek:http.Response{
				StatusCode:       http.StatusOK,
				Body: ioutil.NopCloser(bytes.NewReader(wantRespForLastWeek)),
			},
		}

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(recommendHandler(cache, httpMockClient))

		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		wantedStatus := http.StatusOK
		if status := rr.Code; status != wantedStatus {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, wantedStatus)
		}

		expected := newRecommendation(apiRespForToday, apiRespForLastWeek).String()
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: \n got %v\n want %v",
				rr.Body.String(), expected)
		}
	}

	// Test valid response on the API - "selling" case
	{
		cache.SetWithTTL(today, "", time.Duration(0)) // Invalidate Cache

		apiRespForToday := apiResp{
			Rates: struct {
				GBP float64 `json:"GBP"`
			}{
				GBP: 0.8978,
			},
			Base: "",
			Date: "",
		}
		wantRespForToday, _ := json.Marshal(apiRespForToday)
		apiRespForLastWeek := apiResp{
			Rates: struct {
				GBP float64 `json:"GBP"`
			}{
				GBP: 0.8978,
			},
			Base: "",
			Date: "",
		}
		wantRespForLastWeek, _ := json.Marshal(apiRespForLastWeek)
		httpMockClient := &HttpMockClient{
			wantedResponseForToday:http.Response{
				StatusCode:       http.StatusOK,
				Body: ioutil.NopCloser(bytes.NewReader(wantRespForToday)),
			},
			wantedResponseForLastWeek:http.Response{
				StatusCode:       http.StatusOK,
				Body: ioutil.NopCloser(bytes.NewReader(wantRespForLastWeek)),
			},
		}

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(recommendHandler(cache, httpMockClient))

		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		wantedStatus := http.StatusOK
		if status := rr.Code; status != wantedStatus {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, wantedStatus)
		}

		expected := newRecommendation(apiRespForToday, apiRespForLastWeek).String()
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: \n got %v\n want %v",
				rr.Body.String(), expected)
		}
	}

	// Test valid response from Cache
	{
		// Use responses cached in the previous test scenario
		apiRespForToday := apiResp{
			Rates: struct {
				GBP float64 `json:"GBP"`
			}{
				GBP: 0.8978,
			},
			Base: "",
			Date: "",
		}
		apiRespForLastWeek := apiResp{
			Rates: struct {
				GBP float64 `json:"GBP"`
			}{
				GBP: 0.8978,
			},
			Base: "",
			Date: "",
		}

		httpMockClient := &HttpMockClient{
			wantedResponseForToday:http.Response{
				StatusCode:       http.StatusInternalServerError,
				Body: ioutil.NopCloser(bytes.NewReader([]byte{})),
			},
			wantedResponseForLastWeek:http.Response{
				StatusCode:       http.StatusInternalServerError,
				Body: ioutil.NopCloser(bytes.NewReader([]byte{})),
			},
		}

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(recommendHandler(cache, httpMockClient))

		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		wantedStatus := http.StatusOK
		if status := rr.Code; rr.Code != wantedStatus {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, wantedStatus)
		}

		expected := newRecommendation(apiRespForToday, apiRespForLastWeek).String()
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: \n got %v\n want %v",
				rr.Body.String(), expected)
		}
	}

	// Test one failed HTTP call
	{
		// Invalidate Cache
		cache.SetWithTTL(today, "", time.Duration(0))

		apiRespForToday := apiResp{
			Rates: struct {
				GBP float64 `json:"GBP"`
			}{
				GBP: 0.8978,
			},
			Base: "",
			Date: "",
		}
		wantRespForToday, _ := json.Marshal(apiRespForToday)

		httpMockClient := &HttpMockClient{
			wantedResponseForToday:http.Response{
				StatusCode:       http.StatusOK,
				Body: ioutil.NopCloser(bytes.NewReader(wantRespForToday)),
			},
			wantedResponseForLastWeek:http.Response{
				StatusCode:       http.StatusInternalServerError,
				Body: ioutil.NopCloser(bytes.NewReader([]byte{})), // No response for last week call
			},
		}

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(recommendHandler(cache, httpMockClient))

		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		wantedStatus := http.StatusInternalServerError
		if status := rr.Code; rr.Code != wantedStatus {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, wantedStatus)
		}
	}
}
