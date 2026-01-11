# gw - Future Enhancements Backlog

## High Priority

### Multi-Account Support
- Support multiple Google accounts
- Switch between accounts with `gw auth switch <account>`
- Per-account token storage

### Device Flow Authentication
- Add `gw auth login --device` for headless environments
- Useful for servers and containers without browsers

### JSON Output Mode
- Add `--json` flag for machine-readable output
- Enable piping to jq and other tools
- Better integration with scripts

### Email Composition with Editor
- `gw mail compose` opens $EDITOR for message body
- Support for Markdown to HTML conversion
- Template support

## Medium Priority

### Gmail Features

#### Label Management
- `gw mail label add <id> <label>`
- `gw mail label remove <id> <label>`
- `gw mail label create <name>`

#### Archive/Delete
- `gw mail archive <id>`
- `gw mail trash <id>`
- `gw mail spam <id>`

#### Batch Operations
- `gw mail list --unread | gw mail mark-read`
- Pipeline-friendly IDs

#### Draft Editing
- `gw mail drafts` list drafts
- `gw mail draft edit <id>`
- `gw mail draft send <id>`

#### Signature Support
- Configure default signature
- `gw mail send --no-signature`

### Calendar Features

#### Event Updates
- `gw cal update <id> --start "new time"`
- `gw cal update <id> --title "new title"`

#### Recurring Events
- Create recurring events
- `--recurrence daily/weekly/monthly`
- Handle individual instances

#### RSVP Management
- `gw cal rsvp <id> accept/decline/tentative`
- View attendee responses

#### Free/Busy Query
- `gw cal busy "tomorrow 2pm-4pm"`
- Find available slots

#### Working Hours
- Configure default working hours
- `gw cal available` shows free slots

### UX Improvements

#### Interactive Mode
- `gw shell` for REPL-style interface
- Tab completion for IDs
- Fuzzy search

#### Rich Terminal Output
- Colors for read/unread
- Table formatting
- Progress bars for downloads

#### Notifications
- `gw watch mail` for new email alerts
- `gw watch cal` for upcoming event reminders
- Desktop notifications

## Low Priority

### Google Tasks Integration
- `gw tasks list`
- `gw tasks add "Task description"`
- Link tasks to calendar events

### Google Contacts Integration
- `gw contacts search "John"`
- Auto-complete email addresses
- Contact groups

### Google Meet Integration
- `gw cal create --meet` adds Meet link
- `gw meet join <event-id>` opens Meet

### Export/Import
- `gw mail export --query "..." --format mbox`
- `gw cal export --format ics`
- Backup functionality

### Configuration File
- `~/.gw/config.yaml`
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
