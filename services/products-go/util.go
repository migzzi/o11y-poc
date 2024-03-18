package main

import (
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func newProbabilisticFailureMW(chance float64, handler http.Handler) http.HandlerFunc {
	if chance > 1 {
		panic("chance must be less than 1")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rand.Float64() < chance {
			http.Error(w, "{ \"error\": \"Internal Server Error\" }", http.StatusInternalServerError)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func newRandomLatencyMW(min, max int, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		latency := rand.Intn(max-min) + min
		time.Sleep(time.Duration(latency) * time.Millisecond)
		handler.ServeHTTP(w, r)
	})
}

type IDGenerator struct {
	curr  int
	mutex sync.Mutex
}

func (g *IDGenerator) NextID() int {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.curr++
	return g.curr
}

func newIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

var idGen = newIDGenerator()
