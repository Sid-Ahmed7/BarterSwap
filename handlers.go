package main

import (
	"context"
	"net/http"
	"time"
)

const dbTimeout = 5 * time.Second

func newCtx(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), dbTimeout)
}