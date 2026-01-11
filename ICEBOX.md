# gday - Future Enhancements Backlog

## High Priority

### Multi-Account Support
- Support multiple Google accounts
- Switch between accounts with `gday auth switch <account>`
- Per-account token storage

### Device Flow Authentication
- Add `gday auth login --device` for headless environments
- Useful for servers and containers without browsers

### JSON Output Mode
- Add `--json` flag for machine-readable output
- Enable piping to jq and other tools
- Better integration with scripts

### Email Composition with Editor
- `gday mail compose` opens $EDITOR for message body
- Support for Markdown to HTML conversion
- Template support

## Medium Priority

### Gmail Features

#### Label Management
- `gday mail label add <id> <label>`
- `gday mail label remove <id> <label>`
- `gday mail label create <name>`

#### Archive/Delete
- `gday mail archive <id>`
- `gday mail trash <id>`
- `gday mail spam <id>`

#### Batch Operations
- `gday mail list --unread | gday mail mark-read`
- Pipeline-friendly IDs

#### Draft Editing
- `gday mail drafts` list drafts
- `gday mail draft edit <id>`
- `gday mail draft send <id>`

#### Signature Support
- Configure default signature
- `gday mail send --no-signature`

### Calendar Features

#### Event Updates
- `gday cal update <id> --start "new time"`
- `gday cal update <id> --title "new title"`

#### Recurring Events
- Create recurring events
- `--recurrence daily/weekly/monthly`
- Handle individual instances

#### RSVP Management
- `gday cal rsvp <id> accept/decline/tentative`
- View attendee responses

#### Free/Busy Query
- `gday cal busy "tomorrow 2pm-4pm"`
- Find available slots

#### Working Hours
- Configure default working hours
- `gday cal available` shows free slots

### UX Improvements

#### Interactive Mode
- `gday shell` for REPL-style interface
- Tab completion for IDs
- Fuzzy search

#### Rich Terminal Output
- Colors for read/unread
- Table formatting
- Progress bars for downloads

#### Notifications
- `gday watch mail` for new email alerts
- `gday watch cal` for upcoming event reminders
- Desktop notifications

## Low Priority

### Google Tasks Integration
- `gday tasks list`
- `gday tasks add "Task description"`
- Link tasks to calendar events

### Google Contacts Integration
- `gday contacts search "John"`
- Auto-complete email addresses
- Contact groups

### Google Meet Integration
- `gday cal create --meet` adds Meet link
- `gday meet join <event-id>` opens Meet

### Export/Import
- `gday mail export --query "..." --format mbox`
- `gday cal export --format ics`
- Backup functionality

### Configuration File
- `~/.gday/config.yaml`
- Default calendar, timezone
- Output format preferences
- Aliases for common searches

### Shell Completions
- Bash/Zsh/Fish completions
- Complete message IDs from cache
- Complete calendar names

## Technical Debt

### Testing
- Add unit tests with mocked API responses
- Integration tests with test account
- CI/CD pipeline

### Error Handling
- Better error messages
- Retry logic for transient failures
- Rate limit handling

### Performance
- Cache message list for faster browsing
- Parallel API calls where possible
- Lazy loading of message bodies

### Documentation
- Man pages
- More examples in README
- Video walkthrough
