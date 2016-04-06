package middleware

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/mutil"
)

// LoggerVerbose is the same as Logger but in addition also logs out request headers
func LoggerVerbose(c *web.C, h http.Handler) http.Handler {
	return http.HandlerFunc(loggerHandler(c, h, true))
}

// Logger is a middleware that logs the start and end of each request, along
// with some useful data about what was requested, what the response status was,
// and how long it took to return. When standard output is a TTY, Logger will
// print in color, otherwise it will print in black and white.
//
// Logger prints a request ID if one is provided.
//
// Logger has been designed explicitly to be Good Enough for use in small
// applications and for people just getting started with Goji. It is expected
// that applications will eventually outgrow this middleware and replace it with
// a custom request logger, such as one that produces machine-parseable output,
// outputs logs to a different service (e.g., syslog), or formats lines like
// those printed elsewhere in the application.
func Logger(c *web.C, h http.Handler) http.Handler {
	return http.HandlerFunc(loggerHandler(c, h, false))
}

func loggerHandler(c *web.C, h http.Handler, verbose bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := GetReqID(*c)

		printStart(reqID, r)

		if verbose {
			printHeaders(reqID, r)
		}

		lw := mutil.WrapWriter(w)

		t1 := time.Now()
		h.ServeHTTP(lw, r)

		if lw.Status() == 0 {
			lw.WriteHeader(http.StatusOK)
		}
		t2 := time.Now()

		printEnd(reqID, lw, t2.Sub(t1))
	}
}

func printHeaders(reqID string, r *http.Request) {
	var buf bytes.Buffer

	for k, v := range r.Header {
		if reqID != "" {
			cW(&buf, bBlack, "[%s] ", reqID)
		}

		buf.WriteString(fmt.Sprintf("%s: ", k))

		for ks, vs := range v {
			buf.WriteString(vs)

			if ks+1 < len(v) {
				buf.WriteString(", ")
			}
		}

		log.Print(buf.String())

		buf.Reset()
	}
}

func printStart(reqID string, r *http.Request) {
	var buf bytes.Buffer

	if reqID != "" {
		cW(&buf, bBlack, "[%s] ", reqID)
	}
	buf.WriteString("Started ")
	cW(&buf, bMagenta, "%s ", r.Method)
	cW(&buf, nBlue, "%q ", r.URL.String())
	buf.WriteString("from ")

	if h := r.Header.Get("X-Forwarded-For"); h != "" {
		buf.WriteString(h)
	} else {
		buf.WriteString(r.RemoteAddr)
	}

	log.Print(buf.String())
}

func printEnd(reqID string, w mutil.WriterProxy, dt time.Duration) {
	var buf bytes.Buffer

	if reqID != "" {
		cW(&buf, bBlack, "[%s] ", reqID)
	}
	buf.WriteString("Returning ")
	status := w.Status()
	if status < 200 {
		cW(&buf, bBlue, "%03d", status)
	} else if status < 300 {
		cW(&buf, bGreen, "%03d", status)
	} else if status < 400 {
		cW(&buf, bCyan, "%03d", status)
	} else if status < 500 {
		cW(&buf, bYellow, "%03d", status)
	} else {
		cW(&buf, bRed, "%03d", status)
	}
	buf.WriteString(" in ")
	if dt < 500*time.Millisecond {
		cW(&buf, nGreen, "%s", dt)
	} else if dt < 5*time.Second {
		cW(&buf, nYellow, "%s", dt)
	} else {
		cW(&buf, nRed, "%s", dt)
	}

	log.Print(buf.String())
}
