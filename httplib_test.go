package httplibrary

import (
	"fmt"
	"sync"
	"testing"
)

func TestRequestBuilder( t *testing.T) {
	rb := NewRequestBuilder("GET", "https://jsonplaceholder.typicode.com/posts/1")
	rb.WithHeader("Authorization", "Beerer testkeyyy")
	res, err := rb.Send()
	if err != nil {
		fmt.Printf("err: %s", err)
	}

	var apiResponse ApiResponse

	err = rb.ReadJsonBody(res, &apiResponse)

	if err != nil {
		fmt.Printf("err: %s", err)
	}

	fmt.Printf("apiResponse: %+v", apiResponse)

}

func TestAsyncRequest(t *testing.T) {
	rb := NewRequestBuilder("GET", "https://jsonplaceholder.typicode.com/posts/1")
	rb.WithHeader("Authorization", "Beerer testkeyyy")

	resChan := rb.SendAsync()
	res := <-resChan
	fmt.Println("res read from the channel", "Title:", res.Title, "UserId:", res.UserId)

}

func TestPostRequest(t *testing.T) {

	rb := NewRequestBuilder("POST", "https://jsonplaceholder.typicode.com/posts")

	rb.WithTimeout(5)
	rb.WithRateLimter(5, 3)

	var wg sync.WaitGroup

	///make 10 requests
	for i := 0; i < 10; i++ {
		apiRequest := ApiRequest{
			Title:  "Go HTTP Client",
			UserId: 23,
			Body:   "Sending data with http",
		}

		rb.WithJsonData(apiRequest)
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			res, err := rb.Send()

			if err != nil {
				fmt.Printf("err: %s", err)
				wg.Done()
				return
			}
			//fmt.Println("res", res)

			var apiResponse ApiResponse

			err = rb.ReadJsonBody(res, &apiResponse)

			if err != nil {
				fmt.Printf("err: %s", err)
			}

			fmt.Println("Title:", res.Status, "UserId:", apiResponse.UserId)
			defer wg.Done()
		}(&wg)
		wg.Wait()
	}
}
