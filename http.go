package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func recommendHandler(cache Cache, httpClient httpClient) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		today := time.Now()
		todayCacheKey := today.Format("2006-01-02")

		if val, exists := cache.Get(todayCacheKey); exists && val != ""{
			w.Write([]byte(fmt.Sprintf("%v", val)))
			cacheHits.Inc()
			return
		}

		cacheMisses.Inc()

		sevenDaysAgo := time.Now().AddDate(0, 0, -7)

		respchToday := make(chan httpCallInfo, 1)
		respchSevenDaysAgo := make(chan httpCallInfo, 1)
		errch := make(chan httpCallInfo, 2)

		go doreq(httpClient, respchToday, errch, today)
		go doreq(httpClient, respchSevenDaysAgo, errch, sevenDaysAgo)

		var respToday httpCallInfo
		var respSevenDaysAgo httpCallInfo

		var respCount int
	ReadRespLoop:
		for {
			select {
			case firstErr := <-errch:
				firstErr.sendLogsAndMetrics()
				http.Error(w, firstErr.httpErr.Error(), firstErr.httpStatus)
				return
			case respToday = <-respchToday:
				respToday.sendLogsAndMetrics()
				respCount++
				if respCount == 2 {
					break ReadRespLoop
				}
			case respSevenDaysAgo = <-respchSevenDaysAgo:
				respSevenDaysAgo.sendLogsAndMetrics()
				respCount++
				if respCount == 2 {
					break ReadRespLoop
				}
			}
		}

		rec := newRecommendation(respToday.apiResp, respSevenDaysAgo.apiResp)
		w.Write([]byte(rec.String()))

		cache.SetWithTTL(todayCacheKey, rec.String(), time.Duration(*cacheTtl)*time.Second)
	}
}

type httpCallInfo struct {
	forDate    time.Time
	apiResp    apiResp
	httpStatus int
	httpErr    error
	logMsg     string
	promErr    string
}

func (resp httpCallInfo) sendLogsAndMetrics() {
	if resp.promErr != "" {
		errorCount.WithLabelValues(resp.promErr).Inc()
	}

	if resp.logMsg != "" {
		logger.Error(resp.logMsg)
	}
}

// doreq is a helper func that builds and does the actual HTTP call
func doreq(httpClient httpClient, outch chan httpCallInfo, errch chan httpCallInfo, date time.Time) {
	httpcall := httpCallInfo{
		forDate: date,
	}

	defer func() {
		if httpcall.httpErr != nil {
			errch <- httpcall
		} else {
			outch <- httpcall
		}
	}()

	apiurl, err := buildApiUrl(date)
	if err != nil {
		httpcall.httpStatus = http.StatusInternalServerError
		httpcall.logMsg = fmt.Sprintf("failed to build API url: %s", err)
		httpcall.httpErr = fmt.Errorf("failed to build API url")
		httpcall.promErr = "http_req_build"
		return
	}

	req, err := http.NewRequest(http.MethodGet, apiurl, nil)
	if err != nil {
		httpcall.httpStatus = http.StatusInternalServerError
		httpcall.logMsg = fmt.Sprintf("failed to build HTTP request: %s", err)
		httpcall.httpErr = fmt.Errorf("failed to build HTTP request")
		httpcall.promErr = "http_req_build"
		return
	}

	ctx := context.Background()
	ctx, cancelfn := context.WithTimeout(ctx, time.Duration(*timeout)*time.Millisecond)
	defer cancelfn()

	req = req.WithContext(ctx)
	resp, err := httpClient.Do(req)
	if err != nil {
		httpcall.httpStatus = http.StatusInternalServerError
		httpcall.logMsg = fmt.Sprintf("failed to do HTTP request: %s", err)
		httpcall.httpErr = fmt.Errorf("failed to do HTTP request")
		httpcall.promErr = "http_req_build"
		return
	}
	defer resp.Body.Close()

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		httpcall.httpStatus = http.StatusInternalServerError
		httpcall.logMsg = fmt.Sprintf("failed to read HTTP response body: %s", err)
		httpcall.httpErr = fmt.Errorf("failed to read HTTP response body")
		httpcall.promErr = "http_req_build"
		return
	}

	var apiresp apiResp
	if err = json.Unmarshal(respbody, &apiresp); err != nil {
		httpcall.httpStatus = http.StatusInternalServerError
		httpcall.logMsg = fmt.Sprintf("failed to unmarshal HTTP response: %s", err)
		httpcall.httpErr = fmt.Errorf("failed to unmarshal HTTP response")
		httpcall.promErr = "http_req_build"
		return
	}

	httpcall.apiResp = apiresp

	if ok, err := validateApiResponse(apiresp); !ok || err != nil {
		httpcall.httpStatus = http.StatusInternalServerError
		httpcall.logMsg = fmt.Sprintf("failed to validate HTTP response: %s", err)
		httpcall.httpErr = fmt.Errorf("failed to validate HTTP response")
		httpcall.promErr = "http_req_build"
		return
	}

	httpcall.httpStatus = http.StatusOK
}

func buildApiUrl(date time.Time) (string, error) {
	apiUrlForADate := fmt.Sprintf(apiurltpl, date.Format("2006-01-02"))
	dateUrl, err := url.Parse(apiUrlForADate)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %s", err)
	}

	dateUrl.Query().Set("symbols", "GBP")
	return dateUrl.String(), nil
}

type apiResp struct {
	Rates struct {
		GBP float64 `json:"GBP"`
	} `json:"rates"`
	Base string `json:"base"`
	Date string `json:"date"`
}

func validateApiResponse(ar apiResp) (bool, error) {
	if ar.Rates.GBP == 0 {
		return false, fmt.Errorf("GBP value must not be 0")
	}
	return true, nil
}
