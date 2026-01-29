// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
"encoding/json"
"net/http"

"github.com/julienschmidt/httprouter"
"github.com/syncthing/syncthing/lib/agent"
)

// getAgentSessions returns a list of all agent chat sessions
func (s *service) getAgentSessions(w http.ResponseWriter, r *http.Request) {
query := r.URL.Query().Get("query")

mgr := s.agentMgr

var sessions []*agent.Session
var err error

if query != "" {
sessions, err = mgr.SearchSessions(query)
} else {
sessions, err = mgr.ListSessions()
}

if err != nil {
http.Error(w, err.Error(), http.StatusInternalServerError)
return
}

sendJSON(w, sessions)
}

// getAgentSession returns a specific agent chat session by ID
func (s *service) getAgentSession(w http.ResponseWriter, r *http.Request) {
params := httprouter.ParamsFromContext(r.Context())
sessionID := params.ByName("id")

if sessionID == "" {
http.Error(w, "session ID is required", http.StatusBadRequest)
return
}

mgr := s.agentMgr
session, err := mgr.GetSession(sessionID)
if err != nil {
http.Error(w, err.Error(), http.StatusNotFound)
return
}

sendJSON(w, session)
}

// postAgentSession creates or updates an agent chat session
func (s *service) postAgentSession(w http.ResponseWriter, r *http.Request) {
var session agent.Session

if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
return
}

if session.ID == "" {
http.Error(w, "session ID is required", http.StatusBadRequest)
return
}

mgr := s.agentMgr
if err := mgr.SaveSession(&session); err != nil {
http.Error(w, err.Error(), http.StatusInternalServerError)
return
}

sendJSON(w, map[string]string{
"id":      session.ID,
"message": "session saved successfully",
})
}

// deleteAgentSession deletes a specific agent chat session
func (s *service) deleteAgentSession(w http.ResponseWriter, r *http.Request) {
params := httprouter.ParamsFromContext(r.Context())
sessionID := params.ByName("id")

if sessionID == "" {
http.Error(w, "session ID is required", http.StatusBadRequest)
return
}

mgr := s.agentMgr
if err := mgr.DeleteSession(sessionID); err != nil {
http.Error(w, err.Error(), http.StatusInternalServerError)
return
}

sendJSON(w, map[string]string{
"message": "session deleted successfully",
})
}
