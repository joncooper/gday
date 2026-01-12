# gday Toolchain: Creative Applications Brainstorm

A collection of useful, clever, and unexpected things you can do with gday.

---

## 📊 Quick Counts (Start Here!)

Before diving into large datasets, get the lay of the land:

```bash
gday mail count                          # Total emails
gday mail count --unread                 # Unread count (tens of thousands? time for inbox zero!)
gday mail count -q "from:linkedin.com"   # How many LinkedIn notifications?
gday mail count -q "is:unread has:attachment"  # Unread with attachments
gday mail count -q "older_than:1y"       # Emails over a year old
```

These are single-API-call operations - instant results even with massive inboxes.

---

## 🤖 AI-Powered Workflows (Claude Code Integration)

### 1. **Intelligent Email Triage** (Count-Aware)
```bash
# First, check the scale
gday mail count --unread  # → "47,382 messages (unread)"

# If count is huge, narrow down first
gday mail count -q "is:unread newer_than:7d"  # → "127 messages"

# Now triage the manageable set
gday mail list -q "is:unread newer_than:7d" -n 50 --json | \
  claude "Categorize these by urgency (critical/important/low)"
```
Let AI read your inbox, but be smart about scope first!

### 2. **Auto-Draft Replies**
"Read my last 5 emails and draft contextual replies for each. Use my writing style from the sent folder."
```bash
gday mail list -n 5 --unread --json > inbox.json
gday mail search "in:sent" -n 20 --json > sent.json
# Claude analyzes style and drafts replies
```

### 3. **Meeting Prep Briefings**
Before every meeting, automatically pull:
- The calendar event details
- All email threads with attendees
- Summarize context and action items
```bash
gday cal today --json | claude "For each meeting, find related emails and create a 1-paragraph briefing"
```

### 4. **Smart Scheduling Assistant**
"Find a time this week when I have 2 consecutive free hours and schedule a focus block"
```bash
gday cal week --json | claude "Find 2-hour gaps and suggest the best focus time slots"
```

---

## 🔧 Developer Workflow Integration

### 5. **PR/Issue → Calendar Blocking**
When assigned a complex GitHub issue, auto-create a calendar block:
```bash
gh issue view 123 --json title,body | \
  gday cal create --quick "Work on: $(jq -r .title)"
```

### 6. **Deploy Notifications to Stakeholders**
After successful deploy, email stakeholders with changelog:
```bash
git log --oneline HEAD~5..HEAD | \
  gday mail send --to team@company.com \
    --subject "Deployed: $(git describe --tags)" \
    --body-stdin
```

### 7. **Incident Response Automation**
On-call alert triggers email thread to incident channel:
```bash
gday mail send --to incident@company.com \
  --subject "[P1] Service degradation detected" \
  --body "Alert triggered at $(date). Dashboard: https://..."
```

### 8. **Stand-up Report Generator**
Pull yesterday's sent emails + today's calendar to auto-generate standup:
```bash
gday mail search "in:sent after:$(date -d yesterday +%Y/%m/%d)" --json > sent.json
gday cal today --json > today.json
# Claude generates: "Yesterday I... Today I'll... Blockers:..."
```

---

## 📊 Analytics & Insights

### 9. **Inbox Health Dashboard**
```bash
# One-liner inbox overview
echo "Total: $(gday mail count --json | jq .estimated_total)"
echo "Unread: $(gday mail count --unread --json | jq .estimated_total)"
echo "This week: $(gday mail count -q 'newer_than:7d' --json | jq .estimated_total)"
echo "Newsletters: $(gday mail count -q 'unsubscribe' --json | jq .estimated_total)"
echo "With attachments: $(gday mail count -q 'has:attachment' --json | jq .estimated_total)"
```

### 10. **Email Source Analysis**
```bash
# Who's filling your inbox?
for sender in linkedin.com github.com slack.com jira.atlassian.com; do
  count=$(gday mail count -q "from:$sender" --json | jq .estimated_total)
  echo "$sender: $count"
done
```

### 11. **Email Response Time Tracker**
```bash
gday mail search "in:inbox after:2024/01/01" --json | \
  jq 'map({from, date})' | \
  claude "Calculate average response time by sender"
```

### 10. **Meeting Load Analysis**
```bash
gday cal list --days 30 --json | \
  claude "How many hours/week am I in meetings? Which days are overloaded? Suggest calendar hygiene improvements"
```

### 11. **Communication Pattern Detection**
"Who haven't I emailed in 3 months that I used to email weekly?"
```bash
gday mail search "in:sent after:2023/01/01" --json | \
  claude "Find contacts whose email frequency has dropped significantly"
```

---

## 🎯 Productivity Hacks

### 12. **Email-to-Calendar Extraction**
Someone sends "Let's meet Tuesday at 3pm" - auto-create the event:
```bash
gday mail read <id> --json | \
  claude "Extract any meeting requests and output as gday cal create commands" | \
  bash
```

### 13. **Attachment Auto-Organization**
Download all attachments from a thread and organize by type:
```bash
gday mail attachment <thread-id> --all -o ./downloads
# Then organize: move *.pdf to docs/, *.png to images/, etc.
```

### 14. **Newsletter Digest**
Aggregate all newsletters into a single summary:
```bash
gday mail search "label:newsletters after:$(date -d '1 week ago' +%Y/%m/%d)" --json | \
  claude "Create a bullet-point digest of the key takeaways from these newsletters"
```

### 15. **Focus Time Guardian**
Script that checks calendar before scheduling:
```bash
# Before creating any meeting, check for focus blocks
gday cal list --days 7 --json | \
  jq '.[] | select(.title | contains("Focus"))'
```

---

## 🔄 Automation & Cron Jobs

### 16. **Daily Briefing Email**
Every morning at 7am, email yourself:
- Today's calendar
- Unread email count by category
- Weather (from another API)
```bash
# Cron: 0 7 * * * /path/to/daily-brief.sh
gday cal today --json > /tmp/cal.json
gday mail list --unread --json > /tmp/mail.json
# Combine and email to self
gday mail send --to me@example.com --subject "Daily Brief $(date +%A)" --body-stdin
```

### 17. **Auto-RSVP for Certain Events**
```bash
gday cal list --days 7 --json | \
  jq '.[] | select(.title | contains("Team Standup"))' | \
  # Auto-accept logic
```

### 18. **Stale Thread Detector**
Weekly report of email threads you started but never got a response:
```bash
gday mail search "in:sent -in:inbox older_than:7d" --json | \
  claude "Find threads where I'm waiting for a response"
```

### 19. **Calendar Conflict Monitor**
Hourly check for double-bookings:
```bash
gday cal list --days 3 --all-calendars --json | \
  claude "Identify any overlapping events and alert if found"
```

---

## 🎭 Unexpected/Creative Uses

### 20. **Personal CRM**
Track relationships via email frequency:
```bash
gday mail search "from:important-contact@example.com" --json | \
  claude "When did we last communicate? What were we discussing? Remind me of context."
```

### 21. **Email-Based Time Tracking**
Use sent emails as work log:
```bash
gday mail search "in:sent after:$(date -d 'monday' +%Y/%m/%d)" --json | \
  claude "Estimate hours spent on each project based on email activity"
```

### 22. **Natural Language Calendar Management**
```bash
# Delete all "Canceled" events
gday cal search "Canceled" --json | jq -r '.[].id' | xargs -I {} gday cal delete {}
```

### 23. **Email Template System**
Store templates as files, use with `--body-file`:
```bash
gday mail send --to client@example.com \
  --subject "Project Update" \
  --body-file ~/.gday/templates/weekly-update.md
```

### 24. **Meeting Cost Calculator**
```bash
gday cal week --json | \
  claude "Assuming average salary of $X/hour and Y attendees, calculate the 'cost' of each meeting. Which meetings are most expensive?"
```

### 25. **Async Standup via Email**
Cron job that emails team for standup updates, collects responses:
```bash
gday mail send --to team@example.com \
  --subject "Standup $(date +%Y-%m-%d)" \
  --body "Reply with: Yesterday / Today / Blockers"
```

---

## 🔗 Pipeline Compositions

### 26. **Inbox Zero Automation**
```bash
# Archive all read, non-starred emails older than 30 days
gday mail search "is:read -is:starred older_than:30d" --json | \
  jq -r '.[].id' | \
  xargs -I {} gday mail archive {}  # (if archive command added)
```

### 27. **Cross-Platform Notifications**
Email → Slack/Discord via webhook:
```bash
gday mail list --unread --json | \
  jq '.[] | select(.from | contains("urgent"))' | \
  curl -X POST -d @- $SLACK_WEBHOOK
```

### 28. **Backup Important Threads**
```bash
gday mail search "from:lawyer@firm.com" --json | \
  jq -r '.[].id' | \
  while read id; do gday mail read "$id" >> legal-archive.txt; done
```

### 29. **Smart Email Forwarding**
```bash
gday mail list --unread --json | \
  jq '.[] | select(.subject | contains("Invoice"))' | \
  # Forward to accounting
```

### 30. **Calendar → Timesheet Export**
```bash
gday cal list --days 30 --json | \
  jq 'map({title, start, end, duration: (...)})' | \
  claude "Format as timesheet CSV with project codes"
```

---

## 🧠 AI-Native Features (Future)

### 31. **Email Sentiment Analysis**
"Flag emails where the sender seems frustrated"

### 32. **Smart Unsubscribe Suggestions**
"Which newsletters haven't I opened in 3 months?"

### 33. **Calendar Optimization**
"Rearrange my meetings to minimize context-switching"

### 34. **Auto-Snooze**
"Resurface this email when the sender's project launches"

### 35. **Relationship Health Dashboard**
"Which important contacts am I neglecting?"

---

## 🔐 Security & Compliance

### 36. **Sensitive Email Audit**
```bash
gday mail search "password OR credentials OR SSN" --json | \
  claude "Flag emails with potential security concerns"
```

### 37. **External Email Tracker**
Monitor for data exfiltration patterns:
```bash
gday mail search "in:sent has:attachment to:@gmail.com OR to:@yahoo.com" --json
```

### 38. **Calendar Privacy Check**
```bash
gday cal list --days 30 --json | \
  jq '.[] | select(.attendees | length > 10)' | \
  claude "Large meetings that might need privacy review"
```

---

## Key Enablers

These features make gday especially powerful:

1. **`mail count`** - Fast inbox analytics (single API call, no message fetching)
2. **`--json` output** - Machine-readable for pipelines
3. **`--body-stdin`** - Compose emails from any source
4. **`--body-file`** - Template-based emails
5. **Gmail search syntax** - Precise filtering (`from:`, `after:`, `has:attachment`)
6. **`--quick` natural language** - "Meeting with Bob tomorrow at 3pm"
7. **Thread support** - Full conversation context
8. **Multi-calendar** - Aggregate personal + work calendars

---

## What's Missing for These Use Cases?

To enable more of these workflows, consider adding:

1. **`mail archive`** - Move to archive
2. **`mail label add/remove`** - Label management
3. **`mail forward`** - Forward emails
4. **`cal update`** - Modify existing events
5. **`cal rsvp`** - Accept/decline invitations
6. **`--watch` mode** - Stream new emails/events
7. **Webhook triggers** - Push notifications
8. **`mail batch`** - Operate on multiple IDs

---

*Generated by Claude Code exploring the gday toolchain.*
