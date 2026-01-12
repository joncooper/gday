# gday Improvement Plan

Prioritized improvements based on codebase review.

---

## Phase 1: Critical Fixes

### 1.1 OAuth State Parameter (CSRF Protection)
**File:** `internal/auth/auth.go`

Add state parameter to OAuth flow:
```go
// Generate random state
state := generateRandomState()

// Include in auth URL
authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

// Verify in callback
if r.URL.Query().Get("state") != expectedState {
    http.Error(w, "Invalid state parameter", http.StatusBadRequest)
    return
}
```

### 1.2 Silent Message Failures
**File:** `internal/gmail/gmail.go:73-78`

Change from silent skip to collecting errors:
```go
type ListResult struct {
    Messages []*Message
    Errors   []error  // Failed message IDs
}
```

Warn user if partial results: "Loaded 47 of 50 messages (3 failed)"

### 1.3 Cross-Platform Browser Opening
**File:** `internal/auth/auth.go:393-402`

Use `exec.LookPath()` and support Windows:
```go
func openBrowser(url string) error {
    switch runtime.GOOS {
    case "darwin":
        return exec.Command("open", url).Start()
    case "windows":
        return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
    default:
        // Try common Linux browsers
        for _, browser := range []string{"xdg-open", "sensible-browser", "firefox", "chromium"} {
            if path, err := exec.LookPath(browser); err == nil {
                return exec.Command(path, url).Start()
            }
        }
    }
    return fmt.Errorf("no browser found")
}
```

---

## Phase 2: Safety & UX

### 2.1 Batch Operations Framework

Add to `cmd/helpers.go`:
```go
// BatchConfig controls confirmation behavior
type BatchConfig struct {
    DryRun      bool   // --dry-run: preview without executing
    Yes         bool   // --yes: skip confirmation
    Threshold   int    // Prompt if more than N items (default: 10)
}

// ConfirmBatch prompts for confirmation if needed
func ConfirmBatch(action string, count int, cfg BatchConfig) bool {
    if cfg.DryRun {
        fmt.Printf("[DRY RUN] Would %s %d item(s)\n", action, count)
        return false
    }
    if cfg.Yes || count <= cfg.Threshold {
        return true
    }
    fmt.Printf("%s %d items? [y/N] ", strings.Title(action), count)
    var response string
    fmt.Scanln(&response)
    return strings.ToLower(response) == "y"
}
```

Apply to: `mail archive`, `mail trash`, `cal delete`

Add global flags:
```go
rootCmd.PersistentFlags().Bool("dry-run", false, "Preview changes without executing")
rootCmd.PersistentFlags().Bool("yes", false, "Skip confirmation prompts")
```

### 2.2 Email Validation
**File:** `cmd/mail.go` (send command)

```go
import "net/mail"

func validateEmail(email string) error {
    _, err := mail.ParseAddress(email)
    if err != nil {
        return fmt.Errorf("invalid email address: %s", email)
    }
    return nil
}
```

### 2.3 Result Limiting Indicator

When results are limited, show:
```
Found 1,247 messages matching: from:amazon
Showing first 20 results (use -n to show more, or narrow with date filters)
```

---

## Phase 3: New Features

### 3.1 Label Management

```bash
gday mail label add <message-id> <label>      # Add label
gday mail label remove <message-id> <label>   # Remove label
gday mail label create <name>                 # Create new label
gday mail label delete <name>                 # Delete label
```

**Service layer:**
```go
func (s *Service) AddLabel(ctx context.Context, messageID, labelName string) error
func (s *Service) RemoveLabel(ctx context.Context, messageID, labelName string) error
func (s *Service) CreateLabel(ctx context.Context, name string) (*Label, error)
func (s *Service) DeleteLabel(ctx context.Context, labelID string) error
```

### 3.2 Mark as Read/Unread Batch

```bash
gday mail mark-read <id> [<id>...]
gday mail mark-unread <id> [<id>...]
```

### 3.3 Pagination Support

Add to JSON output:
```json
{
  "count": 50,
  "next_page_token": "abc123...",
  "messages": [...]
}
```

Add flag:
```bash
gday mail search "query" --page-token <token>
```

### 3.4 Verbosity Levels

```bash
gday mail list              # Quiet (default)
gday mail list -v           # Verbose: "Fetching messages... Found 10"
gday mail list --debug      # Debug: Full API logging
```

---

## Phase 4: Code Quality

### 4.1 Remove Custom min()
**File:** `cmd/calendar.go:753-758`

Delete custom `min()` function, use Go 1.21+ built-in.

### 4.2 Fix -q Flag Inconsistency

| Command | Current | Proposed |
|---------|---------|----------|
| `mail list -q` | query | query (keep) |
| `mail search` | (positional) | (keep) |
| `mail count -q` | query | query (keep) |
| `cal create -q` | quick | `--quick` only (remove `-q`) |

### 4.3 Document Date Formats

Update help text for all time-related flags:
```
--start   Start time. Formats: "2024-01-15 14:00", "2024-01-15T14:00",
          "01/02/2006 15:04", "01/02/2006 3:04 PM"
```

### 4.4 File Refactoring

Split monolithic files:

```
cmd/
├── mail.go           → cmd/mail/
│                       ├── mail.go (root command)
│                       ├── list.go
│                       ├── read.go
│                       ├── send.go
│                       ├── search.go
│                       ├── archive.go
│                       └── label.go
├── calendar.go       → cmd/calendar/
│                       ├── calendar.go
│                       ├── list.go
│                       ├── create.go
│                       ├── update.go
│                       └── rsvp.go
```

Or simpler: keep single files but extract helpers to `cmd/mail_helpers.go`, `cmd/calendar_helpers.go`.

---

## Phase 5: Testing

### 5.1 Test Structure

```
tests/
├── unit/
│   ├── gmail_test.go        # Mock API responses
│   ├── calendar_test.go
│   ├── auth_test.go
│   └── helpers_test.go
├── integration/
│   ├── mail_test.go         # Against test account
│   └── calendar_test.go
└── fixtures/
    ├── messages.json        # Sample API responses
    ├── events.json
    └── html_emails/         # For HTML-to-text testing
        ├── simple.html
        ├── complex.html
        └── malformed.html
```

### 5.2 Priority Test Cases

1. **HTML-to-text conversion** - Many edge cases
2. **Date parsing** - All supported formats
3. **OAuth flow** - State parameter validation
4. **Email validation** - RFC 5322 compliance
5. **Batch operations** - Confirmation logic

### 5.3 CI/CD

`.github/workflows/test.yml`:
```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - run: go test ./...
      - run: go build ./...
```

---

## Implementation Order

| Priority | Item | Effort | Impact |
|----------|------|--------|--------|
| 1 | OAuth state parameter | 1h | Critical security |
| 2 | --dry-run and --yes flags | 2h | Safety |
| 3 | Silent failures fix | 1h | Reliability |
| 4 | Cross-platform browser | 1h | Compatibility |
| 5 | Email validation | 30m | UX |
| 6 | Result limit indicator | 30m | UX |
| 7 | Remove custom min() | 5m | Cleanup |
| 8 | Fix -q flag | 15m | Consistency |
| 9 | Label management | 3h | Feature |
| 10 | Mark read/unread batch | 1h | Feature |
| 11 | Pagination support | 2h | Feature |
| 12 | Verbosity flags | 1h | UX |
| 13 | Document date formats | 30m | Docs |
| 14 | Unit tests (core) | 4h | Quality |
| 15 | File refactoring | 2h | Maintainability |
| 16 | CI/CD setup | 1h | Quality |

**Total estimated effort:** ~20 hours

---

## Command Safety Matrix

| Command | Destructive | Batch | Needs --dry-run | Needs confirm |
|---------|-------------|-------|-----------------|---------------|
| mail list | No | No | No | No |
| mail read | No | No | No | No |
| mail search | No | No | No | No |
| mail count | No | No | No | No |
| mail send | No | No | Yes (preview) | No |
| mail reply | No | No | Yes (preview) | No |
| mail archive | Yes | Yes | Yes | If >10 items |
| mail trash | Yes | Yes | Yes | If >10 items |
| mail label add | No | Yes | Yes | No |
| mail label remove | Yes | Yes | Yes | If >10 items |
| cal create | No | No | No | No |
| cal update | Yes | No | Yes | No |
| cal delete | Yes | Yes | Yes | Always |
| cal rsvp | No | No | No | No |
