# Readwise-Compatible Highlight Server Design

## [S1] Problem

Moon Reader (Android reading app) supports syncing highlights/notes to Readwise via its API. The user wants to self-host a Readwise-compatible server to receive and store these highlights locally, avoiding dependence on the Readwise cloud service.

## [S2] Solution Overview

Build a lightweight Go + Gin HTTP server that implements the minimal Readwise API v3 endpoints required by Moon Reader for highlight syncing. The server stores highlights in SQLite and provides export functionality via API.

## [S3] Architecture

```
Moon Reader → HTTP POST → Readwise Server → SQLite
                              ↓
                         Export API (optional)
```

### Components

1. **HTTP Server** - Gin-based REST API
2. **Storage Layer** - SQLite via `modernc.org/sqlite` (pure Go, no CGO)
3. **Token Auth** - Config-file-based bearer token authentication
4. **Export Module** - JSON/CSV export endpoints

## [S4] API Endpoints

### Authentication

All endpoints require `Authorization: Token <token>` header.

### Highlight Endpoints (Moon Reader Integration)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v3/highlights/` | Create or update highlights (batch) |
| GET | `/api/v3/highlights/` | List all highlights |
| GET | `/api/v3/highlights/:id` | Get single highlight |
| DELETE | `/api/v3/highlights/:id` | Delete a highlight |
| GET | `/api/v3/books/` | List books with highlights |
| GET | `/api/v3/books/:id` | Get book details |

### Export Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/export/json` | Export all highlights as JSON |
| GET | `/api/export/csv` | Export all highlights as CSV |

## [S5] Data Model

### highlights table

| Column | Type | Description |
|--------|------|-------------|
| id | INTEGER PRIMARY KEY | Auto-increment ID |
| readwise_id | TEXT UNIQUE | Readwise-compatible ID |
| text | TEXT NOT NULL | Highlighted text |
| note | TEXT | User's note |
| title | TEXT NOT NULL | Book title |
| author | TEXT | Book author |
| source | TEXT | Source URL or identifier |
| color | TEXT | Highlight color |
| book_type | TEXT | Book type (book, article, etc.) |
| created_at | DATETIME | Creation timestamp |
| updated_at | DATETIME | Last update timestamp |

### books table

| Column | Type | Description |
|--------|------|-------------|
| id | INTEGER PRIMARY KEY | Auto-increment ID |
| title | TEXT NOT NULL | Book title |
| author | TEXT | Book author |
| num_highlights | INTEGER DEFAULT 0 | Highlight count |

## [S6] Moon Reader API Compatibility

Moon Reader typically sends:

```json
POST /api/v3/highlights/
{
  "highlights": [
    {
      "text": "highlighted text",
      "title": "Book Title",
      "author": "Author Name",
      "source": "moon+reader",
      "color": "yellow"
    }
  ]
}
```

The server should:
1. Accept this format
2. Generate Readwise-compatible IDs
3. Store in SQLite
4. Return success response in Readwise format

## [S7] Configuration

`config.yaml`:
```yaml
server:
  host: "0.0.0.0"
  port: 8080

auth:
  token: "your-secret-token-here"

database:
  path: "./data/highlights.db"
```

## [S8] Export Format

### JSON Export
```json
{
  "highlights": [...],
  "total": 123,
  "exported_at": "2026-06-13T10:00:00Z"
}
```

### CSV Export
```csv
id,text,note,title,author,source,color,created_at
1,"highlight text","note","Book","Author","source","yellow","2026-01-01"
```

## [S9] Error Handling

| Status | Description |
|--------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad request / invalid JSON |
| 401 | Invalid or missing token |
| 404 | Resource not found |
| 500 | Internal server error |

## [S10] Project Structure

```
moon/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── router.go
│   │   ├── middleware.go
│   │   ├── highlights.go
│   │   ├── books.go
│   │   └── export.go
│   ├── storage/
│   │   ├── sqlite.go
│   │   └── models.go
│   └── config/
│       └── config.go
├── config.yaml
├── go.mod
└── go.sum
```

## [S11] Deployment

- Single binary, no dependencies
- systemd service file provided
- Data directory for SQLite database
- Log output to stdout/stderr

## [S12] Success Criteria

1. Moon Reader can successfully sync highlights to the server
2. Highlights are stored correctly in SQLite
3. Token authentication works
4. Export endpoints return correct data
5. Server starts and runs without errors
