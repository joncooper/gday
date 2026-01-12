# gday - Gmail & Google Calendar CLI

A fast, single-binary CLI for accessing Gmail and Google Calendar from the terminal. Designed for developers, automation, and seamless integration with Claude Code.

## Features

**Gmail:**
- List, search, and read emails
- Send new emails and replies
- Download attachments
- Thread support
- Full Gmail search syntax

**Google Calendar:**
- List events across all calendars
- Quick views: today, tomorrow, week
- Create events with natural language
- Search and manage events
- Multi-calendar support

**Developer Features:**
- JSON output mode for scripting and automation
- Device flow authentication for headless environments
- Claude Code skill integration

## Quick Start

### 1. Build

```bash
go build -o gday .
```

### 2. Set Up OAuth Credentials

You need to create OAuth2 credentials in Google Cloud Console:

1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Create a new project (or select existing)
3. Enable the **Gmail API** and **Google Calendar API**
4. Configure OAuth consent screen:
   - Choose "External" user type
   - Add your email as a test user
5. Create OAuth 2.0 Client ID:
   - Application type: **Desktop application**
   - Download the JSON file
6. Run `gday auth setup` and paste the JSON contents

### 3. Authenticate

```bash
gday auth login
```

This opens your browser for Google authentication. Token is cached at `~/.gday/token.json`.

**For headless environments** (SSH, containers, servers):

```bash
gday auth login --device
```

This uses OAuth2 device flow - you'll get a URL and code to enter on any device.

### 4. Use

```bash
# Gmail
gday mail list              # Recent emails
gday mail list --unread     # Unread only
gday mail read <id>         # Read email
gday mail send --to user@example.com --subject "Hi" --body "Hello!"

# Calendar
gday cal today              # Today's events
gday cal week               # This week
gday cal create --quick "Lunch tomorrow at noon"
```

## JSON Output Mode

All commands support `--json` flag for machine-readable output:

```bash
# Get emails as JSON
gday mail list --json
gday mail read <id> --json
gday mail search "from:boss" --json

# Get calendar events as JSON
gday cal today --json
gday cal list --json
gday cal show <event-id> --json
```

This is useful for:
- Scripting and automation
- Piping to `jq` for processing
- Integration with other tools
- Claude Code integration

## Gmail Commands

### List Emails

```bash
gday mail list                    # List 10 recent emails
gday mail list -n 25              # List 25 emails
gday mail list --unread           # Only unread
gday mail list -q "from:boss"     # With search query
gday mail list --json             # JSON output
```

### Count Emails

Get email counts efficiently (single API call, no message fetching):

```bash
gday mail count                   # Count all emails
gday mail count --unread          # Count unread emails
gday mail count -q "from:boss"    # Count emails from specific sender
gday mail count -q "has:attachment" --json  # JSON output
```

### Read Email

```bash
gday mail read <message-id>       # Read message
gday mail read <id> --raw         # Raw format
gday mail read <id> --mark-read   # Mark as read
gday mail read <id> --json        # JSON output
gday mail thread <thread-id>      # Read full thread
```

### Search

```bash
gday mail search "from:alice@example.com"
gday mail search "subject:urgent is:unread"
gday mail search "has:attachment larger:5M"
gday mail search "after:2024/01/01 before:2024/02/01"
```

### Send Email

```bash
gday mail send --to user@example.com --subject "Hello" --body "Message"
gday mail send --to user@example.com --subject "Hello" --body-file msg.txt
echo "Message" | gday mail send --to user@example.com --subject "Hello" --body-stdin
gday mail send --to user@example.com --subject "Hello" --body "Hi" --cc other@example.com
gday mail send --to user@example.com --subject "Hello" --body "Hi" --draft  # Create draft only
```

### Reply

```bash
gday mail reply <message-id> --body "Thanks for your message"
gday mail reply <message-id> --body-file reply.txt
```

### Attachments

```bash
gday mail attachment <message-id>                  # List attachments
gday mail attachment <message-id> <attachment-id>  # Download one
gday mail attachment <message-id> --all            # Download all
gday mail attachment <message-id> --all -o ./downloads
```

### Labels

```bash
gday mail labels  # List all labels
```

## Calendar Commands

### View Events

```bash
gday cal list                 # Next 10 events
gday cal list -n 20           # Next 20 events
gday cal list --days 30       # Next 30 days
gday cal list --all-calendars # From all calendars

gday cal today                # Today's events
gday cal tomorrow             # Tomorrow's events
gday cal week                 # This week's events

gday cal show <event-id>      # Event details
```

### Create Events

```bash
# Specific times
gday cal create --title "Meeting" --start "2024-01-15 14:00" --end "2024-01-15 15:00"
gday cal create --title "Meeting" --start "2024-01-15 14:00"  # 1 hour default

# All-day events
gday cal create --title "Vacation" --date "2024-01-20" --all-day

# With location and attendees
gday cal create --title "Team Sync" --start "2024-01-15 10:00" \
  --location "Conference Room A" \
  --attendees alice@company.com,bob@company.com

# Natural language (Quick Add)
gday cal create --quick "Lunch with John tomorrow at noon"
gday cal create --quick "Project deadline January 31st"
gday cal create --quick "1:1 with manager Friday 3-4pm"
```

### Search and Delete

```bash
gday cal search "team meeting"
gday cal search "John" --days 90

gday cal delete <event-id>
```

### Calendars

```bash
gday cal calendars                          # List all calendars
gday cal list --calendar <calendar-id>      # Events from specific calendar
```

## Authentication Commands

```bash
gday auth setup    # Configure OAuth credentials
gday auth login    # Authenticate with Google (browser)
gday auth login --device  # Authenticate (device flow, for SSH/headless)
gday auth logout   # Clear cached token
gday auth status   # Check auth status
```

## Configuration

All configuration is stored in `~/.gday/`:

```
~/.gday/
├── credentials.json   # OAuth client credentials
└── token.json         # Cached access token
```

## Integration with Claude Code

This tool is designed to work with Claude Code through the `gday` skill. The skill is included in this repository at `.claude/skills/gday/SKILL.md`.

When the skill is active, Claude Code can naturally interact with your Gmail and Calendar - just ask about your emails or schedule and it will use gday automatically.

## Gmail Search Syntax

The `gday` tool supports Gmail's full search syntax:

| Query | Description |
|-------|-------------|
| `from:user@example.com` | From specific sender |
| `to:user@example.com` | To specific recipient |
| `subject:keyword` | Subject contains keyword |
| `is:unread` | Unread messages |
| `is:starred` | Starred messages |
| `has:attachment` | Has attachments |
| `filename:pdf` | Has PDF attachments |
| `larger:5M` | Larger than 5MB |
| `smaller:1M` | Smaller than 1MB |
| `after:2024/01/01` | After date |
| `before:2024/02/01` | Before date |
| `older_than:7d` | Older than 7 days |
| `newer_than:1d` | From last day |
| `label:important` | Has label |
| `in:inbox` | In inbox |
| `in:sent` | In sent |

Combine with `AND`, `OR`, `NOT`, and parentheses:
```bash
gday mail search "(from:alice OR from:bob) AND is:unread"
```

## Building from Source

Requirements:
- Go 1.21 or later

```bash
git clone https://github.com/joncooper/gday
cd gday
go build -o gday .

# Install to PATH
sudo mv gday /usr/local/bin/
```

## License

MIT
