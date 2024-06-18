package main

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

func MakeRequests(routes []string, ch chan map[string]int, wg *sync.WaitGroup, numRequests int) {
	defer wg.Done()
	client := &http.Client{}
	length := len(routes)
	for i := 0; i < numRequests; i++ {
		req, err := http.NewRequest("GET", routes[i%length], nil)
		if err != nil {
			fmt.Println("==== ERROR: ", err)
			panic("critical error")
		}

		authorization := "gr0KeGhUltOHL9BCGHg0EVVRXVE2:eyJhbGciOiJSUzI1NiIsImtpZCI6ImFlYzU4NjcwNGNhOTZiZDcwMzZiMmYwZDI4MGY5NDlmM2E5NzZkMzgiLCJ0eXAiOiJKV1QifQ.eyJuYW1lIjoiR3JpZmZpbiBCb3VyZG9uIiwicGljdHVyZSI6Imh0dHBzOi8vbGgzLmdvb2dsZXVzZXJjb250ZW50LmNvbS9hL0FDZzhvY0xkTG9BTGgwUG4xS3NCdElEd1lnV2dsTHlnclRFUVVPdmdjRElySGY0Vz1zOTYtYyIsImlzcyI6Imh0dHBzOi8vc2VjdXJldG9rZW4uZ29vZ2xlLmNvbS94b3BhY2tzLWRldmVsb3BtZW50LTc4YWI3IiwiYXVkIjoieG9wYWNrcy1kZXZlbG9wbWVudC03OGFiNyIsImF1dGhfdGltZSI6MTcwODA1NzYwNiwidXNlcl9pZCI6ImdyMEtlR2hVbHRPSEw5QkNHSGcwRVZWUlhWRTIiLCJzdWIiOiJncjBLZUdoVWx0T0hMOUJDR0hnMEVWVlJYVkUyIiwiaWF0IjoxNzA4MDU3NjA2LCJleHAiOjE3MDgwNjEyMDYsImVtYWlsIjoiYWRtaW5AeG9wYWNrcy5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZmlyZWJhc2UiOnsiaWRlbnRpdGllcyI6eyJnb29nbGUuY29tIjpbIjEwNzU2NDk0NTU1NjU1MTA2MTMyNyJdLCJlbWFpbCI6WyJhZG1pbkB4b3BhY2tzLmNvbSJdfSwic2lnbl9pbl9wcm92aWRlciI6Imdvb2dsZS5jb20ifX0.f3aWSkr49QCVGmbflg41b9F2ML29G6CeMOj_hJdK-P3XrzeLh5gHS39NWGIjdZX9gXpbqi0h-vTYgKLJftKWl2xqQhQMl7RsPdsEG80LtneGhgUPF87HoYAWbzz8xOG4Hd4jW-0e0wmX90eGxPGQVrl60-Bzsssr4BTw32cwdtgZA6Zt9gTXvJ6KpDWYKXsjn_gFLjY9kQgFeuNBo9ifiEbEvSAmapCMKXKEOJLMu8_N3gptyMsZDCd_zQVgZCoK22Zgl8OQgqCvjuFVPORKi2_dMF41U1OzHPKHc9yARx7qt70Tfcx2pkk7w1Yx9KTa1OCFeu-gpVsQ8P2tACX5xQ"
		req.Header.Set("Authorization", authorization)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("==== ERROR: ", err)
			panic("CRITICAL ERROR SERVER CRASH")
		}
		ch <- map[string]int{routes[i%length]: resp.StatusCode}
		if resp.StatusCode == 503 {
			fmt.Println(" 503 Request Url: ", routes[i%length])
		}
		if resp.StatusCode == 500 {
			fmt.Println(" 500 Request URL: ", routes[i%length])
		}
		time.Sleep(time.Millisecond * 500)
	}
	return
}

func RequestConsumer(ch chan map[string]int, wg *sync.WaitGroup, requestLogs map[string]map[int]int) {
	defer wg.Done()

	for {
		reqData, ok := <-ch
		if !ok {
			fmt.Println("Channel close. Exiting consumer..")
			return
		}

		for k, v := range reqData {
			_, exists := requestLogs[k][v]
			if !exists {
				requestLogs[k][v] = 1
			} else {
				requestLogs[k][v] += 1
			}
		}
	}
}

func FormatLog(log map[string]map[int]int) {
	for k, reqMap := range log {
		fmt.Printf("URL: %v -> Results: %v\n", k, reqMap)
		fmt.Println()
	}
}

func TestAppLoad1(t *testing.T) {
	routes := []string{
		"https://apidev.xopacks.com/vendors/packs?pageNum=1",
		"https://apidev.xopacks.com/categories/all",
		// "https://apidev.xopacks.com/firebase/getUserByEmail?email=phillip.bourdon@xopacks.com",
		"https://apidev.xopacks.com/user/username/Phirmware",
		"https://apidev.xopacks.com/user/gr0KeGhUltOHL9BCGHg0EVVRXVE2?authorizedUid=gr0KeGhUltOHL9BCGHg0EVVRXVE2",
		//"https://apidev.xopacks.com/analytics/packQtySold/gr0KeGhUltOHL9BCGHg0EVVRXVE2?uid=gr0KeGhUltOHL9BCGHg0EVVRXVE2",
		"https://apidev.xopacks.com/user/items/gr0KeGhUltOHL9BCGHg0EVVRXVE2?pageNum=1&sortBy=&sortDir=&filterOn=&search=&authorizedUid=gr0KeGhUltOHL9BCGHg0EVVRXVE2",
		"https://apidev.xopacks.com/user/packs/gr0KeGhUltOHL9BCGHg0EVVRXVE2?pageNum=1&sortBy=&sortDir=&filterOn=&search=&authorizedUid=gr0KeGhUltOHL9BCGHg0EVVRXVE2",
	}

	numWorkers := 1000
	numRequests := 18
	var consumerWg sync.WaitGroup
	consumerWg.Add(1)
	reqChannel := make(chan map[string]int, numWorkers*numRequests)
	requestLogs := map[string]map[int]int{}
	routeMap := map[int]int{}

	for _, route := range routes {
		requestLogs[route] = routeMap
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go MakeRequests(routes, reqChannel, &wg, numRequests)
		// time.Sleep(time.Millisecond * 100)
	}

	go RequestConsumer(reqChannel, &consumerWg, requestLogs)

	wg.Wait()
	close(reqChannel)
	consumerWg.Wait()

	fmt.Println("halt")
	FormatLog(requestLogs)
}
