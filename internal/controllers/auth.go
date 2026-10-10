package controllers

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
)

// Login serves GET /login. TODO(stage 4).
func (c *Controllers) Login(w http.ResponseWriter, r *http.Request) { httpx.NotImplemented(w) }

// LoginSubmit serves POST /login (starts OAuth or sends a magic link). TODO(stage 4).
func (c *Controllers) LoginSubmit(w http.ResponseWriter, r *http.Request) { httpx.NotImplemented(w) }

// AuthCallback serves GET /auth/callback. TODO(stage 4).
func (c *Controllers) AuthCallback(w http.ResponseWriter, r *http.Request) { httpx.NotImplemented(w) }

// Logout serves POST /logout. TODO(stage 4).
func (c *Controllers) Logout(w http.ResponseWriter, r *http.Request) { httpx.NotImplemented(w) }
