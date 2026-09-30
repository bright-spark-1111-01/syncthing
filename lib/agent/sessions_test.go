// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package agent

import (
"testing"
"strconv"
"time"

"github.com/syncthing/syncthing/internal/db"
"github.com/syncthing/syncthing/internal/db/sqlite"
)

func TestSessionSaveAndGet(t *testing.T) {
tmpDir := t.TempDir()

ldb, err := sqlite.Open(tmpDir)
if err != nil {
t.Fatal(err)
}
defer ldb.Close()

typed := db.NewMiscDB(ldb)
mgr := NewManager(typed)

session := &Session{
ID:        "test-session-1",
Timestamp: time.Now(),
Title:     "Test Session",
Messages: []Message{
{
Role:      "user",
Content:   "Hello",
Timestamp: time.Now(),
},
{
Role:      "assistant",
Content:   "Hi there!",
Timestamp: time.Now(),
},
},
Metadata: Metadata{
Branch:     "main",
Repository: "test/repo",
Author:     "testuser",
TaskType:   "feature",
Status:     "completed",
Tags:       []string{"test", "example"},
},
}

// Save session
err = mgr.SaveSession(session)
if err != nil {
t.Fatalf("Failed to save session: %v", err)
}

// Retrieve session
retrieved, err := mgr.GetSession(session.ID)
if err != nil {
t.Fatalf("Failed to get session: %v", err)
}

// Verify data
if retrieved.ID != session.ID {
t.Errorf("ID mismatch: got %s, want %s", retrieved.ID, session.ID)
}
if retrieved.Title != session.Title {
t.Errorf("Title mismatch: got %s, want %s", retrieved.Title, session.Title)
}
if len(retrieved.Messages) != len(session.Messages) {
t.Errorf("Message count mismatch: got %d, want %d", len(retrieved.Messages), len(session.Messages))
}
if retrieved.Metadata.Branch != session.Metadata.Branch {
t.Errorf("Branch mismatch: got %s, want %s", retrieved.Metadata.Branch, session.Metadata.Branch)
}
}

func TestListSessions(t *testing.T) {
tmpDir := t.TempDir()

ldb, err := sqlite.Open(tmpDir)
if err != nil {
t.Fatal(err)
}
defer ldb.Close()

typed := db.NewMiscDB(ldb)
mgr := NewManager(typed)

// Save multiple sessions
for i := 1; i <= 3; i++ {
session := &Session{
ID:        "session-" + strconv.Itoa(i),
Title:     "Session " + strconv.Itoa(i),
Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
}
if err := mgr.SaveSession(session); err != nil {
t.Fatalf("Failed to save session %d: %v", i, err)
}
}

// List all sessions
sessions, err := mgr.ListSessions()
if err != nil {
t.Fatalf("Failed to list sessions: %v", err)
}

if len(sessions) != 3 {
t.Errorf("Expected 3 sessions, got %d", len(sessions))
}

// Verify sessions are sorted by timestamp (most recent first)
for i := 1; i < len(sessions); i++ {
if sessions[i].Timestamp.After(sessions[i-1].Timestamp) {
t.Error("Sessions not sorted correctly by timestamp")
}
}
}

func TestDeleteSession(t *testing.T) {
tmpDir := t.TempDir()

ldb, err := sqlite.Open(tmpDir)
if err != nil {
t.Fatal(err)
}
defer ldb.Close()

typed := db.NewMiscDB(ldb)
mgr := NewManager(typed)

session := &Session{
ID:    "session-to-delete",
Title: "Delete Me",
}

// Save and then delete
err = mgr.SaveSession(session)
if err != nil {
t.Fatalf("Failed to save session: %v", err)
}

err = mgr.DeleteSession(session.ID)
if err != nil {
t.Fatalf("Failed to delete session: %v", err)
}

// Verify it's gone
_, err = mgr.GetSession(session.ID)
if err == nil {
t.Error("Expected error when getting deleted session, got nil")
}

// Verify it's not in the list
sessions, err := mgr.ListSessions()
if err != nil {
t.Fatalf("Failed to list sessions: %v", err)
}

for _, s := range sessions {
if s.ID == session.ID {
t.Error("Deleted session still in list")
}
}
}

func TestSearchSessions(t *testing.T) {
tmpDir := t.TempDir()

ldb, err := sqlite.Open(tmpDir)
if err != nil {
t.Fatal(err)
}
defer ldb.Close()

typed := db.NewMiscDB(ldb)
mgr := NewManager(typed)

sessions := []*Session{
{
ID:    "session-1",
Title: "Fix bug in authentication",
Metadata: Metadata{
Branch:   "bugfix/auth",
TaskType: "bugfix",
Tags:     []string{"security", "auth"},
},
},
{
ID:    "session-2",
Title: "Add new feature",
Metadata: Metadata{
Branch:   "feature/new-api",
TaskType: "feature",
Tags:     []string{"api", "enhancement"},
},
},
{
ID:    "session-3",
Title: "Update documentation",
Metadata: Metadata{
Branch:   "docs/update",
TaskType: "documentation",
Tags:     []string{"docs"},
},
},
}

for _, s := range sessions {
if err := mgr.SaveSession(s); err != nil {
t.Fatalf("Failed to save session: %v", err)
}
}

// Test search by title
results, err := mgr.SearchSessions("bug")
if err != nil {
t.Fatalf("Search failed: %v", err)
}
if len(results) != 1 || results[0].ID != "session-1" {
t.Error("Search by title failed")
}

// Test search by branch
results, err = mgr.SearchSessions("feature")
if err != nil {
t.Fatalf("Search failed: %v", err)
}
if len(results) != 1 || results[0].ID != "session-2" {
t.Error("Search by branch failed")
}

// Test search by tag
results, err = mgr.SearchSessions("docs")
if err != nil {
t.Fatalf("Search failed: %v", err)
}
if len(results) != 1 || results[0].ID != "session-3" {
t.Error("Search by tag failed")
}

// Test empty search returns all
results, err = mgr.SearchSessions("")
if err != nil {
t.Fatalf("Search failed: %v", err)
}
if len(results) != 3 {
t.Errorf("Empty search should return all sessions, got %d", len(results))
}
}

func TestEmptySessionID(t *testing.T) {
tmpDir := t.TempDir()

ldb, err := sqlite.Open(tmpDir)
if err != nil {
t.Fatal(err)
}
defer ldb.Close()

typed := db.NewMiscDB(ldb)
mgr := NewManager(typed)

session := &Session{
ID:    "",
Title: "No ID",
}

err = mgr.SaveSession(session)
if err == nil {
t.Error("Expected error when saving session with empty ID, got nil")
}
}
