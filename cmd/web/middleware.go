package main

import (
	"fmt"
	"github.com/justinas/nosurf"
	"net/http"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		resp.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		resp.Header().Set("X-Content-Type-Options", "nosniff")
		resp.Header().Set("X-Frame-Options", "deny")
		resp.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(resp, req)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", req.RemoteAddr, req.Proto, req.Method, req.URL.RequestURI())
		next.ServeHTTP(resp, req)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				resp.Header().Set("Connection", "close")
				app.serverError(resp, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(resp, req)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if !app.isAuthenticated(req) {
			http.Redirect(resp, req, "/user/login", http.StatusSeeOther)
			return
		}
		resp.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(resp, req)
	})
}

func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true, Path: "/", Secure: true,
	})
	return csrfHandler
}