package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	jgerrors "github.com/jasonhancock/jasongen/errors"
	"github.com/stretchr/testify/require"
)

type ctxKey struct{}

type recordingResponder struct {
	err error
}

func (r *recordingResponder) Err(w http.ResponseWriter, _ *http.Request, err error) {
	r.err = err
	status := http.StatusInternalServerError
	var sc interface{ StatusCode() int }
	if errors.As(err, &sc) {
		status = sc.StatusCode()
	}
	w.WriteHeader(status)
}

func noCreds(*http.Request) (context.Context, error) { return nil, ErrNoCredentials }

// missingHeaderErr stands in for an implementer's own error type: it satisfies
// NoCredentialer without referring to anything in this package.
type missingHeaderErr struct {
	header string
	absent bool
}

func (e missingHeaderErr) Error() string       { return "missing " + e.header }
func (e missingHeaderErr) NoCredentials() bool { return e.absent }

func accept(name string) Authenticator {
	return func(r *http.Request) (context.Context, error) {
		return context.WithValue(r.Context(), ctxKey{}, name), nil
	}
}

func reject(err error) Authenticator {
	return func(*http.Request) (context.Context, error) { return nil, err }
}

type scheme struct {
	gotArgs []string
}

func (s *scheme) Authenticate(r *http.Request, args ...string) (context.Context, error) {
	s.gotArgs = args
	return r.Context(), nil
}

func TestAnyAuth(t *testing.T) {
	forbidden := jgerrors.NewHTTP(errors.New("missing scope"), http.StatusForbidden)

	tests := []struct {
		desc           string
		authenticators []Authenticator
		wantStatus     int
		wantCaller     string
	}{
		{"first matches", []Authenticator{accept("a"), accept("b")}, http.StatusOK, "a"},
		{"falls through absent credentials", []Authenticator{noCreds, accept("b")}, http.StatusOK, "b"},
		{"wrapped no credentials falls through", []Authenticator{reject(errors.Join(errors.New("x"), ErrNoCredentials)), accept("b")}, http.StatusOK, "b"},
		{"own error type falls through", []Authenticator{reject(missingHeaderErr{"X-Signed", true}), accept("b")}, http.StatusOK, "b"},
		{"wrapped own error type falls through", []Authenticator{reject(fmt.Errorf("checker: %w", missingHeaderErr{"X-Signed", true})), accept("b")}, http.StatusOK, "b"},
		{"opting out is final", []Authenticator{reject(jgerrors.NewHTTP(missingHeaderErr{"X-Signed", false}, http.StatusUnauthorized)), accept("b")}, http.StatusUnauthorized, ""},
		{"rejection is final", []Authenticator{reject(forbidden), accept("b")}, http.StatusForbidden, ""},
		{"none present", []Authenticator{noCreds, noCreds}, http.StatusUnauthorized, ""},
		{"nil skipped", []Authenticator{nil, accept("b")}, http.StatusOK, "b"},
		{"all nil passes through", []Authenticator{nil, nil}, http.StatusOK, ""},
		{"untyped error is a 500", []Authenticator{reject(errors.New("boom"))}, http.StatusInternalServerError, ""},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var gotCaller string
			called := false
			h := AnyAuth(&recordingResponder{}, tt.authenticators...)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				gotCaller, _ = r.Context().Value(ctxKey{}).(string)
			}))

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

			require.Equal(t, tt.wantStatus, rec.Code)
			require.Equal(t, tt.wantStatus == http.StatusOK, called)
			require.Equal(t, tt.wantCaller, gotCaller)
		})
	}
}

func TestBind(t *testing.T) {
	require.Nil(t, Bind(nil, "x"))

	var s scheme
	_, err := Bind(&s, "scope_a", "scope_b")(httptest.NewRequest(http.MethodGet, "/", nil))
	require.NoError(t, err)
	require.Equal(t, []string{"scope_a", "scope_b"}, s.gotArgs)
}
