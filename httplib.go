package httplibrary

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"

	//"sync"
	"time"

	"golang.org/x/time/rate"
)

type ApiRequest struct {
	UserId 		int   `json:"userId"`
	Title		string `json:"title"`
	Body		string `json:"body"`
}


type ApiResponse struct {
	UserId 		int    `json:"userId"`
	Title		string `json:"title"`
}

type RequestBuilder struct {
	method 		string
	url			string
	headers 	map[string]string
	body  		io.Reader
	timeout     time.Duration
	retries     int
	client		*http.Client
	middlewares  []Middleware
	limiter      *rate.Limiter
}

type RequestError struct {
	StatusCode    int
	Err           error
}

type UnmarshalError struct {
	Err           error
}

type MarshalError struct {
	Err           error
}


func (re *RequestError) Error() string {
	return fmt.Sprintf("httplib request error: %v with code: %v", re.Err, re.StatusCode)
}

func (re *UnmarshalError) Error() string {
	return fmt.Sprintf("httplib unmarshal error: %v", re.Err)
}

func (re *MarshalError) Error() string {
	return fmt.Sprintf("httplib unmarshal error: %v", re.Err)
}

var defaultTransport = &http.Transport{
	MaxIdleConns: 5,
	IdleConnTimeout: 5 * time.Second,
}

func NewRequestBuilder(method, url string) *RequestBuilder {
	return &RequestBuilder{
		url: url,
		method: method,
		headers: make(map[string]string),
		timeout: 3 * time.Second,
		retries:  3,
		client: &http.Client{Transport: defaultTransport},
	}
}

type Middleware func(*http.Request, *http.Response) error

type Backoff func(retry int) time.Duration

func (rb *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	rb.headers[key] = value
	return rb
}

func (rb *RequestBuilder) WithBody(body io.Reader) {
	rb.body = body
}

func (rb *RequestBuilder) WithRateLimter(rps int, burst int) *RequestBuilder {
	rb.limiter = rate.NewLimiter(rate.Limit(rps), burst)
	return rb
}

func (rb *RequestBuilder) WithJsonData(data interface{}) error {
	b, err := json.Marshal(data)

	if err != nil {
		return &MarshalError{Err: err}
	}

	reader := bytes.NewBuffer(b)

	rb.body = reader

	rb.headers["Content"] = "application/json"

	return nil

}

func (rb *RequestBuilder) ReadJsonBody(r *http.Response, rTarget *ApiResponse) error {

	var err error
	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)

	fmt.Println("b", string(b))

	if err != nil {
		return &UnmarshalError{Err: err}
	}

	err = json.Unmarshal(b, rTarget);

	fmt.Printf("target: %+v", rTarget)

	if err != nil {
		return &UnmarshalError{Err: err}
	}

	return nil
}

func (rb *RequestBuilder) Send() (*http.Response, error) {

	if rb.limiter != nil {
		fmt.Println("limiter exists")
		rb.limiter.Wait(context.Background())
	}

	req, err := http.NewRequest(rb.method, rb.url, rb.body)

	if err != nil  {
		return nil, &RequestError{Err: err, StatusCode: 500}
	}

	for key, value := range rb.headers {
		req.Header.Set(key, value)
	}

	res, err := rb.client.Do(req)

	for _, middleware := range(rb.middlewares) {
		err = middleware(req, res)
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	return res, nil

}

func (rb *RequestBuilder) WithTimeout(timeout time.Duration) *RequestBuilder {
	rb.timeout = timeout
	return rb 
}

func (rb *RequestBuilder) SendAsync() <- chan ApiResponse {
	resultChan := make(chan ApiResponse)


	//create a goroutine that brings about concurrent processing
	go func ()  {
		res, _ := rb.Send()
		var target ApiResponse
		rb.ReadJsonBody(res, &target)
		resultChan <- ApiResponse{Title: target.Title, UserId: target.UserId}
		close(resultChan)
	}()
	//wg.Wait()
	return resultChan
}

// ExponentialBackoff calculates the backoff duration for retries using exponential backoff strategy

func ExponentialBackoff(retry int) time.Duration {
	baseDelay := time.Millisecond * 100
	maxDelay := 10 * time.Second
	factor := 2.0
	jitter := 0.2

	//we start at 100
	delay := float64(baseDelay) * math.Pow(factor, float64(retry))

	if delay > float64(maxDelay) {
		delay = float64(maxDelay)
	}

	//20
	jitterRange := jitter * delay

	//100 - 10
	delay -= jitterRange/2
	delay += rand.Float64() * delay

	return time.Duration(delay)

}

func RetryTracker() {
	for i := 0; i < 4; i++ {
		fmt.Println("start")
		duration := ExponentialBackoff(i)
		fmt.Printf("wait for %s at %v\n", duration, i)
		time.Sleep(duration)
	}
}