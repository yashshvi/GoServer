package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type RequestPayload struct {
	InputArray [][]int `json:"to_sort"`
}

type ResponsePayload struct {
	ProcessTime  string  `json:"time_ns"`
	SortedArrays [][]int `json:"sorted_arrays"`
}

func main() {

	http.HandleFunc("/process-single", sequentialSort)
	http.HandleFunc("/process-concurrent", concurrentSort)
	port := ":8000"

	fmt.Println("Server is running on port" + port)

	log.Fatal(http.ListenAndServe(port, nil))
}

func sequentialSort(w http.ResponseWriter, r *http.Request) {
	sortFun(w, r, false)
}

func concurrentSort(w http.ResponseWriter, r *http.Request) {
	sortFun(w, r, true)
}

func sortFun(w http.ResponseWriter, r *http.Request, concurrent bool) {
	// Parse JSON payload
	fmt.Println("bools is ", concurrent)
	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	sTime := time.Now()
	fmt.Println("bools is ", sTime)

	var sortedArrays [][]int
	if concurrent {
		sortedArrays = sortConcurrently(payload.InputArray)
	} else {
		sortedArrays = sortSequentially(payload.InputArray)
	}

	eTime := time.Now()
	duration := eTime.Sub(sTime)

	// fmt.Println("duration is " + eTime.Sub(sTime))
	// Prepare and send the response
	response := ResponsePayload{
		ProcessTime:  duration.String(),
		SortedArrays: sortedArrays,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
func sortSequentially(arrays [][]int) [][]int {

	fmt.Println("sortSequentially")
	var sortedArrays [][]int
	for _, arr := range arrays {
		sortedArr := make([]int, len(arr))
		copy(sortedArr, arr)
		sort.Ints(sortedArr)
		sortedArrays = append(sortedArrays, sortedArr)
	}
	return sortedArrays
	//
}

func sortConcurrently(arrays [][]int) [][]int {
	fmt.Println("sortConcurrently")
	var wg sync.WaitGroup
	var mu sync.Mutex
	var sortedArrays [][]int

	for _, arr := range arrays {
		wg.Add(1)
		go func(arr []int) {
			defer wg.Done()
			sortedArr := make([]int, len(arr))
			copy(sortedArr, arr)
			sort.Ints(sortedArr)

			mu.Lock()
			sortedArrays = append(sortedArrays, sortedArr)
			mu.Unlock()
		}(arr)
	}

	wg.Wait()
	return sortedArrays
}
