# Agent Chat Sessions

This package provides functionality to store and retrieve GitHub Copilot agent chat sessions within Syncthing.

## Overview

The agent session tracking system allows you to:

- Store chat sessions between users and AI agents
- Search and filter previous sessions
- View session details including messages and metadata
- Delete old sessions

## API Endpoints

### List All Sessions

```
GET /rest/agent/sessions
```

Query Parameters:
- `query` (optional): Search term to filter sessions

Example Response:
```json
[
  {
    "id": "session-123",
    "timestamp": "2026-01-29T00:00:00Z",
    "title": "Fix authentication bug",
    "messages": [...],
    "metadata": {
      "branch": "bugfix/auth",
      "repository": "bright-spark-1111-01/syncthing",
      "author": "agent",
      "taskType": "bugfix",
      "status": "completed",
      "tags": ["security", "auth"]
    }
  }
]
```

### Get Specific Session

```
GET /rest/agent/sessions/{id}
```

Example Response:
```json
{
  "id": "session-123",
  "timestamp": "2026-01-29T00:00:00Z",
  "title": "Fix authentication bug",
  "messages": [
    {
      "role": "user",
      "content": "Fix the authentication issue in the login flow",
      "timestamp": "2026-01-29T00:00:00Z"
    },
    {
      "role": "assistant",
      "content": "I'll help you fix that...",
      "timestamp": "2026-01-29T00:00:05Z"
    }
  ],
  "metadata": {
    "branch": "bugfix/auth",
    "status": "completed"
  }
}
```

### Create/Update Session

```
POST /rest/agent/sessions
Content-Type: application/json

{
  "id": "session-123",
  "title": "Fix authentication bug",
  "messages": [...],
  "metadata": {...}
}
```

### Delete Session

```
DELETE /rest/agent/sessions/{id}
```

## Usage in Go

```go
import (
    "github.com/syncthing/syncthing/lib/agent"
    "github.com/syncthing/syncthing/internal/db"
)

// Create a session manager
mgr := agent.NewManager(miscDB)

// Save a session
session := &agent.Session{
    ID:        "session-123",
    Title:     "Example Session",
    Timestamp: time.Now(),
    Messages: []agent.Message{
        {
            Role:      "user",
            Content:   "Hello",
            Timestamp: time.Now(),
        },
    },
    Metadata: agent.Metadata{
        Branch:   "main",
        TaskType: "feature",
        Status:   "in_progress",
    },
}

err := mgr.SaveSession(session)

// Retrieve a session
session, err := mgr.GetSession("session-123")

// List all sessions
sessions, err := mgr.ListSessions()

// Search sessions
results, err := mgr.SearchSessions("authentication")

// Delete a session
err := mgr.DeleteSession("session-123")
```

## Web UI

Access the Agent Chat Sessions viewer from the Syncthing web UI:

1. Open the Syncthing web interface
2. Click on "Actions" in the top navigation bar
3. Select "Agent Sessions"

The interface allows you to:
- View all stored sessions
- Search for specific sessions
- View session details and messages
- Delete sessions

## Data Storage

Sessions are stored in the Syncthing database using the key-value store with the prefix `agentsession:`. The session list is maintained in the key `agentsessions:list`.

## Security

Access to agent sessions requires authentication to the Syncthing API. All API endpoints use the same authentication mechanism as other Syncthing REST endpoints.
