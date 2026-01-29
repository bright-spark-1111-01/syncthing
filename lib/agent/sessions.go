// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package agent

import (
"encoding/json"
"fmt"
"sort"
"strings"
"time"

"github.com/syncthing/syncthing/internal/db"
)

const (
sessionKeyPrefix = "agentsession:"
sessionListKey   = "agentsessions:list"
)

// Session represents a GitHub Copilot agent chat session
// Session represents a GitHub Copilot agent chat session containing
// chat messages, metadata, and timestamps for tracking AI interactions.
type Session struct {
// ID is the unique identifier for the session
ID        string    `json:"id"`
// Timestamp is when the session was created or last modified
Timestamp time.Time `json:"timestamp"`
// Title is a human-readable description of the session
Title     string    `json:"title"`
// Messages contains the conversation between user and assistant
Messages  []Message `json:"messages,omitempty"`
// Metadata contains additional information about the session
Metadata  Metadata  `json:"metadata,omitempty"`
}

// Message represents a single message in a chat session
// Message represents a single message in a chat session with
// role, content, and timestamp information.
type Message struct {
// Role is either "user" or "assistant"
Role      string    `json:"role"`
// Content is the text content of the message
Content   string    `json:"content"`
// Timestamp is when the session was created or last modified
Timestamp time.Time `json:"timestamp"`
}

// Metadata contains additional information about the session
// Metadata contains additional contextual information about a session
// including repository details, task categorization, and custom data.
type Metadata struct {
// Branch is the git branch associated with the session
Branch      string            `json:"branch,omitempty"`
// Repository is the full repository name (e.g., "owner/repo")
Repository  string            `json:"repository,omitempty"`
// Author is the user or agent that created the session
Author      string            `json:"author,omitempty"`
// TaskType categorizes the type of work (e.g., "bugfix", "feature")
TaskType    string            `json:"taskType,omitempty"`
// Status indicates the current state (e.g., "completed", "in_progress")
Status      string            `json:"status,omitempty"`
// Tags are keywords for categorization and searching
Tags        []string          `json:"tags,omitempty"`
// CustomData allows storing arbitrary key-value pairs
CustomData  map[string]string `json:"customData,omitempty"`
}

// Manager handles storage and retrieval of agent chat sessions
type Manager struct {
db *db.Typed
}

// NewManager creates a new session manager
func NewManager(database *db.Typed) *Manager {
return &Manager{
db: database,
}
}

// SaveSession stores a chat session in the database
func (m *Manager) SaveSession(session *Session) error {
if session.ID == "" {
return fmt.Errorf("session ID cannot be empty")
}

// Set timestamp if not provided
if session.Timestamp.IsZero() {
session.Timestamp = time.Now()
}

// Serialize session to JSON
data, err := json.Marshal(session)
if err != nil {
return fmt.Errorf("failed to marshal session: %w", err)
}

// Store session data
key := sessionKeyPrefix + session.ID
if err := m.db.PutBytes(key, data); err != nil {
return fmt.Errorf("failed to store session: %w", err)
}

// Update session list
if err := m.addToSessionList(session.ID); err != nil {
return fmt.Errorf("failed to update session list: %w", err)
}

return nil
}

// GetSession retrieves a specific session by ID
func (m *Manager) GetSession(id string) (*Session, error) {
key := sessionKeyPrefix + id
data, ok, err := m.db.Bytes(key)
if err != nil {
return nil, fmt.Errorf("failed to retrieve session: %w", err)
}
if !ok {
return nil, fmt.Errorf("session not found: %s", id)
}

var session Session
if err := json.Unmarshal(data, &session); err != nil {
return nil, fmt.Errorf("failed to unmarshal session: %w", err)
}

return &session, nil
}

// ListSessions returns all stored sessions, optionally filtered
func (m *Manager) ListSessions() ([]*Session, error) {
sessionIDs, err := m.getSessionList()
if err != nil {
return nil, fmt.Errorf("failed to get session list: %w", err)
}

sessions := make([]*Session, 0, len(sessionIDs))
for _, id := range sessionIDs {
session, err := m.GetSession(id)
if err != nil {
// Skip sessions that can't be loaded
continue
}
sessions = append(sessions, session)
}

// Sort by timestamp, most recent first
sort.Slice(sessions, func(i, j int) bool {
return sessions[i].Timestamp.After(sessions[j].Timestamp)
})

return sessions, nil
}

// DeleteSession removes a session from the database
func (m *Manager) DeleteSession(id string) error {
key := sessionKeyPrefix + id
if err := m.db.Delete(key); err != nil {
return fmt.Errorf("failed to delete session: %w", err)
}

if err := m.removeFromSessionList(id); err != nil {
return fmt.Errorf("failed to update session list: %w", err)
}

return nil
}

// SearchSessions finds sessions matching a query
func (m *Manager) SearchSessions(query string) ([]*Session, error) {
allSessions, err := m.ListSessions()
if err != nil {
return nil, err
}

if query == "" {
return allSessions, nil
}

query = strings.ToLower(query)
filtered := make([]*Session, 0)

for _, session := range allSessions {
if m.sessionMatchesQuery(session, query) {
filtered = append(filtered, session)
}
}

return filtered, nil
}

// sessionMatchesQuery checks if a session matches the search query
func (m *Manager) sessionMatchesQuery(session *Session, query string) bool {
// Check title
if strings.Contains(strings.ToLower(session.Title), query) {
return true
}

// Check metadata fields
if strings.Contains(strings.ToLower(session.Metadata.Branch), query) {
return true
}
if strings.Contains(strings.ToLower(session.Metadata.TaskType), query) {
return true
}
if strings.Contains(strings.ToLower(session.Metadata.Status), query) {
return true
}

// Check tags
for _, tag := range session.Metadata.Tags {
if strings.Contains(strings.ToLower(tag), query) {
return true
}
}

// Check messages
for _, msg := range session.Messages {
if strings.Contains(strings.ToLower(msg.Content), query) {
return true
}
}

return false
}

// addToSessionList adds a session ID to the master list
func (m *Manager) addToSessionList(id string) error {
sessionIDs, err := m.getSessionList()
if err != nil {
return err
}

// Check if already in list
for _, existingID := range sessionIDs {
if existingID == id {
return nil
}
}

sessionIDs = append(sessionIDs, id)
return m.saveSessionList(sessionIDs)
}

// removeFromSessionList removes a session ID from the master list
func (m *Manager) removeFromSessionList(id string) error {
sessionIDs, err := m.getSessionList()
if err != nil {
return err
}

filtered := make([]string, 0, len(sessionIDs))
for _, existingID := range sessionIDs {
if existingID != id {
filtered = append(filtered, existingID)
}
}

return m.saveSessionList(filtered)
}

// getSessionList retrieves the list of all session IDs
func (m *Manager) getSessionList() ([]string, error) {
data, ok, err := m.db.Bytes(sessionListKey)
if err != nil {
return nil, err
}
if !ok {
return []string{}, nil
}

var sessionIDs []string
if err := json.Unmarshal(data, &sessionIDs); err != nil {
return nil, err
}

return sessionIDs, nil
}

// saveSessionList saves the list of all session IDs
func (m *Manager) saveSessionList(sessionIDs []string) error {
data, err := json.Marshal(sessionIDs)
if err != nil {
return err
}

return m.db.PutBytes(sessionListKey, data)
}
