// Package auth defines the interfaces generated servers require of security
// schemes, and the runtime support for operations that list more than one
// security requirement. OpenAPI treats such a list as alternatives: a request that
// satisfies any one of them is allowed.
package auth

import (
	"context"
	"errors"
	"net/http"

	jgerrors "github.com/jasonhancock/jasongen/errors"
)

// ErrNoCredentials is returned by an Authenticator when the request carries none
// of the credentials it understands, so the next Authenticator should be tried.
// Any other error is final.
var ErrNoCredentials = errors.New("no credentials")

// Authenticator authenticates and authorizes a request. On success it returns the
// context the handler should run with, typically carrying the caller's identity.
//
// Return ErrNoCredentials (or an error wrapping it) only when the request has none
// of this scheme's credentials. Credentials that are present but invalid, or a
// caller lacking the required scopes, must return some other error so that a
// different scheme can't override the rejection. Errors implementing
// StatusCode() int are rendered with that status.
type Authenticator func(r *http.Request) (context.Context, error)

// Authorizer is implemented by a security scheme that an operation lists as its
// only requirement. The returned middleware enforces the scheme with the given
// scopes and renders its own errors. The generated interface for such a scheme
// embeds this.
type Authorizer interface {
	Authorized(args ...string) func(next http.Handler) http.Handler
}

// Scheme is implemented by a security scheme that can take part in an AnyAuth.
// The generated interface for such a scheme embeds this method.
type Scheme interface {
	Authenticate(r *http.Request, args ...string) (context.Context, error)
}

// Responder renders errors. The Responder generated alongside each HTTPServer
// satisfies it.
type Responder interface {
	Err(w http.ResponseWriter, req *http.Request, err error)
}

// Bind returns an Authenticator for scheme with the requirement's scopes fixed. A
// nil scheme yields a nil Authenticator, which AnyAuth skips. That mirrors how a
// nil scheme disables its single-requirement middleware.
func Bind(scheme Scheme, args ...string) Authenticator {
	if scheme == nil {
		return nil
	}
	return func(r *http.Request) (context.Context, error) {
		return scheme.Authenticate(r, args...)
	}
}

// AnyAuth tries each authenticator in order. The first one to find its
// credentials decides the outcome: success runs the handler with the returned
// context, and any error other than ErrNoCredentials is rendered without trying
// the rest. If none find credentials, the request is rejected with a 401.
//
// Nil authenticators are skipped. If all are nil, requests pass through
// unauthenticated, consistent with passing nil schemes to a generated server.
func AnyAuth(respond Responder, authenticators ...Authenticator) func(http.Handler) http.Handler {
	active := make([]Authenticator, 0, len(authenticators))
	for _, a := range authenticators {
		if a != nil {
			active = append(active, a)
		}
	}

	return func(next http.Handler) http.Handler {
		if len(active) == 0 {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, authenticate := range active {
				ctx, err := authenticate(r)
				if errors.Is(err, ErrNoCredentials) {
					continue
				}
				if err != nil {
					respond.Err(w, r, err)
					return
				}
				if ctx == nil {
					ctx = r.Context()
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			respond.Err(w, r, jgerrors.NewHTTP(ErrNoCredentials, http.StatusUnauthorized))
		})
	}
}
