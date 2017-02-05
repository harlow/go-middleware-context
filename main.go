package main

import (
	"expvar"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/harlow/go-middleware-example/requestid"
	"github.com/harlow/go-middleware-example/userip"
	"github.com/paulbellamy/ratecounter"
)

var (
	counter       *ratecounter.RateCounter
	hitsperminute = expvar.NewInt("hits_per_minute")
)

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if reqID, ok := requestid.FromRequest(r); ok == nil {
			ctx = requestid.NewContext(ctx, reqID)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func userIPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if userIP, ok := userip.FromRequest(r); ok == nil {
			ctx = userip.NewContext(ctx, userIP)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requestCtrMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter.Incr(1)
		hitsperminute.Set(counter.Rate())
		next.ServeHTTP(w, r)
	})
}

func reqHandler(w http.ResponseWriter, r *http.Request) {
	reqID, _ := requestid.FromContext(r.Context())
	userIP, _ := userip.FromContext(r.Context())
	fmt.Fprintf(w, "Hello request: %s, from %s\n", reqID, userIP)
}

func main() {
	counter = ratecounter.NewRateCounter(1 * time.Minute)

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	var h http.Handler
	h = http.HandlerFunc(reqHandler)
	h = userIPMiddleware(h)
	h = requestIDMiddleware(h)
	h = requestCtrMiddleware(h)

	http.Handle("/", h)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
