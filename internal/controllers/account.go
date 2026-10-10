package controllers

import "net/http"

// Account serves GET /account: profile and statistics. Anonymous visitors are
// sent to the login page. TODO(stage 4): render the page for the logged-in user
// (see DevAccount for the template data).
func (c *Controllers) Account(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
