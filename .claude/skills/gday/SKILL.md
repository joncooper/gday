---
name: gday
description: Gmail and Google Calendar CLI for checking inbox, reading/sending emails, viewing calendar events, and creating meetings
---

# gday - Gmail & Google Calendar CLI

Use this skill when users ask about their Gmail or Google Calendar. This includes:
- Finding, searching, or reading emails
- Checking inbox or unread messages
- Sending emails or replies
- Viewing calendar events (today, tomorrow, this week)
- Creating or deleting calendar events
- Searching for specific emails or events

## Prerequisites

The `gday` CLI is installed at `~/.local/bin/gday`.

The user must have authenticated:
```bash
gday auth status  # Check if authenticated
gday auth login   # Authenticate if needed
```

## Gmail Commands

```bash
# Count emails (fast - single API call)
gday mail count                   # Total email count
gday mail count --unread          # Unread count
gday mail count -q "from:boss"    # Count from specific sender
gday mail count --json            # JSON output

# List recent emails
gday mail list                    # 10 recent emails
gday mail list -n 25              # 25 emails
gday mail list --unread           # Only unread
gday mail list --json             # JSON output for parsing

# Read an email
gday mail read <message-id>       # Read message
gday mail read <id> --json        # JSON output

# Search emails (Gmail search syntax)
gday mail search "from:someone@example.com"
gday mail search "subject:invoice"
gday mail search "gift card"
gday mail search "has:attachment filename:pdf"
gday mail search "is:unread after:2024/01/01"

# Read a thread
gday mail thread <thread-id>

# Send email
gday mail send --to user@example.com --subject "Subject" --body "Message"

# Reply to email
gday mail reply <message-id> --body "Reply text"

# List/download attachments
gday mail attachment <message-id>           # List attachments
gday mail attachment <message-id> --all     # Download all

# List labels
gday mail labels
```

## Calendar Commands

```bash
# View events
gday cal today                    # Today's events
gday cal tomorrow                 # Tomorrow's events
gday cal week                     # This week's events
gday cal list                     # Next 10 events
gday cal list -n 20 --days 30     # Next 20 events in 30 days
gday cal list --json              # JSON output

# Show event details
gday cal show <event-id>

# Create events
gday cal create --quick "Lunch tomorrow at noon"
gday cal create --title "Meeting" --start "2024-01-15 14:00"

# Search events
gday cal search "meeting"
gday cal search "John" --days 90

# Delete event
gday cal delete <event-id>

# List calendars
gday cal calendars
```

## JSON Output

All commands support `--json` for machine-readable output. Use this when you need to parse or analyze the data:

```bash
gday mail list --json | jq '.messages[] | select(.is_unread)'
gday mail search "gift card" --json
gday cal today --json
```

## Common Patterns

**Check inbox scale before fetching (important for large inboxes!):**
```bash
gday mail count --unread          # Might return 47,382!
gday mail count -q "is:unread newer_than:7d"  # Narrow down first
```

**Find specific emails:**
```bash
gday mail search "from:amazon subject:gift card" --json
```

**Check for unread from specific sender:**
```bash
gday mail search "from:boss@company.com is:unread"
```

**Get today's schedule:**
```bash
gday cal today --json
```

**Quick calendar event:**
```bash
gday cal create --quick "Call with Sarah Friday 2pm"
```
