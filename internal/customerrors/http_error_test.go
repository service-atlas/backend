package customerrors

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleError_HTTPError(t *testing.T) {
	rr := httptest.NewRecorder()
	msg := "not found"
	HandleError(rr, &HTTPError{Status: http.StatusNotFound, Msg: msg})

	res := rr.Result()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, res.StatusCode)
	}

	// http.Error writes the message with a trailing newline
	body := rr.Body.String()
	expectedBody := msg + "\n"
	if body != expectedBody {
		t.Fatalf("unexpected body: got %q want %q", body, expectedBody)
	}

	// Basic header checks that http.Error sets
	if ct := res.Header.Get("Content-Type"); ct == "" {
		t.Fatalf("expected Content-Type to be set, got empty")
	}
}

func TestHandleError_WrappedHTTPError(t *testing.T) {
	rr := httptest.NewRecorder()
	msg := "unauthorized access"
	err := &HTTPError{Status: http.StatusUnauthorized, Msg: msg}
	wrappedErr := fmt.Errorf("error wrapping: %w", err)

	HandleError(rr, wrappedErr)

	res := rr.Result()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.StatusCode)
	}

	body := rr.Body.String()
	expectedBody := msg + "\n"
	if body != expectedBody {
		t.Fatalf("unexpected body: got %q want %q", body, expectedBody)
	}
}

func TestHandleError_GenericError(t *testing.T) {
	rr := httptest.NewRecorder()
	genErr := errors.New("boom")

	HandleError(rr, genErr)

	res := rr.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, res.StatusCode)
	}

	body := rr.Body.String()
	expectedBody := genErr.Error() + "\n"
	if body != expectedBody {
		t.Fatalf("unexpected body: got %q want %q", body, expectedBody)
	}
}
