package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMiddlewareCORS(t *testing.T) {
	nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	corsHandler := middlewareCORS(nextHandler)

	t.Run("regular request headers", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/test", nil)

		corsHandler.ServeHTTP(recorder, request)

		if origin := recorder.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("expected origin *, got %q", origin)
		}
		if methods := recorder.Header().Get("Access-Control-Allow-Methods"); methods != "GET, POST, PUT, DELETE, OPTIONS" {
			t.Errorf("unexpected methods: %q", methods)
		}
		if headers := recorder.Header().Get("Access-Control-Allow-Headers"); headers != "Content-Type, X-User-ID" {
			t.Errorf("unexpected headers: %q", headers)
		}
		if recorder.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", recorder.Code)
		}
	})

	t.Run("OPTIONS preflight request", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodOptions, "/api/test", nil)

		corsHandler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", recorder.Code)
		}
	})
}

func TestMiddlewareRecovery(t *testing.T) {
	panicHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		panic("test panic")
	})

	recoveryHandler := middlewareRecovery(panicHandler)

	t.Run("panic recovery status code", func(t *testing.T) {
		var logBuffer bytes.Buffer
		log.SetOutput(&logBuffer)
		defer log.SetOutput(os.Stderr)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/test", nil)

		recoveryHandler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}

		if !strings.Contains(logBuffer.String(), "panic: test panic") {
			t.Errorf("expected log to contain panic message, got %q", logBuffer.String())
		}
	})
}

func TestMiddlewareLogging(t *testing.T) {
	nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusAccepted)
	})

	loggingHandler := middlewareLogging(nextHandler)

	t.Run("log output verification", func(t *testing.T) {
		var logBuffer bytes.Buffer
		log.SetOutput(&logBuffer)
		defer log.SetOutput(os.Stderr)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/test", nil)

		loggingHandler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusAccepted {
			t.Errorf("expected status 202, got %d", recorder.Code)
		}

		logString := logBuffer.String()
		if !strings.Contains(logString, "POST /api/test 202") {
			t.Errorf("expected log to contain method, path and status, got %q", logString)
		}
	})
}

func TestBuildHandler(t *testing.T) {
	nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	handlerChain := BuildHandler(nextHandler)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/test", nil)

	handlerChain.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	if origin := recorder.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("expected origin *, got %q", origin)
	}
}
