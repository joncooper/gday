# gday - Implementation Notes

## Architecture

### Project Structure

```
gday/
├── main.go                    # Entry point
├── cmd/                       # Cobra commands
│   ├── root.go               # Root command and help
│   ├── auth.go               # Authentication commands
│   ├── mail.go               # Gmail commands
│   └── calendar.go           # Calendar commands
└── internal/
    ├── auth/
    │   └── auth.go           # OAuth2 flow handling
    ├── config/
    │   └── config.go         # Config file management
    ├── gmail/
    │   └── gmail.go          # Gmail API wrapper
    └── calendar/
        └── calendar.go       # Calendar API wrapper
```

### Design Decisions

1. **Single Binary**: Go compiles to a single binary with no runtime dependencies, making installation simple.

2. **Cobra CLI Framework**: Industry standard for Go CLIs. Provides subcommands, flags, help generation, and shell completion.

3. **Local OAuth Flow**: Uses localhost callback for OAuth2 instead of device flow. This works better in development environments where a browser is available.

4. **Token Caching**: OAuth tokens are cached in `~/.gday/token.json` and automatically refreshed when expired.

5. **Minimal External Dependencies**: Only essential dependencies:
   - `github.com/spf13/cobra` - CLI framework
   - `golang.org/x/oauth2` - OAuth2 handling
   - `google.golang.org/api` - Google APIs client

### OAuth2 Setup

The tool requires users to create their own OAuth credentials because:
1. Google's verification process for OAuth apps with mail/calendar scope is extensive
2. Using personal credentials keeps data access transparent
3. No need to manage a central OAuth app or worry about rate limits

The setup process:
1. User creates OAuth credentials in Google Cloud Console
2. Downloads credentials JSON
3. Pastes JSON into `gday auth setup`
4. Runs `gday auth login` to complete OAuth flow
5. Token is cached and refreshed automatically

### Gmail Implementation

- Uses Gmail API v1
- Messages fetched with metadata format by default (faster)
- Full format used when reading message body
- HTML emails converted to plain text for terminal display
- Attachments downloaded via separate API call

### Calendar Implementation

- Uses Calendar API v3
- Events sorted by start time
- Supports both timed and all-day events
- Quick Add uses Google's natural language parsing
- Multi-calendar support via calendar ID flag

## Known Limitations

1. **Browser Required for Auth**: Initial login requires a browser for OAuth consent. Could add device flow as fallback.

2. **No Offline Support**: Requires internet connection for all operations.

3. **HTML Email Display**: HTML to text conversion is basic. Complex emails may not render well.

4. **No Draft Editing**: Can create drafts but not edit existing ones.

5. **Single Account**: Only supports one Google account at a time.

## Future Improvements

See ICEBOX.md for planned enhancements.

## Testing

Manual testing has been performed for:
- Authentication flow
- Email listing and reading
- Email sending and replying
- Calendar event listing
- Event creation with various formats

Automated tests could be added with mocked API responses.

## Security Considerations

1. **Credentials Storage**: OAuth credentials and tokens stored with 0600 permissions
2. **No Secrets in Code**: All sensitive data from config files
3. **Token Refresh**: Tokens auto-refresh, no long-term storage of refresh tokens exposed
4. **HTTPS Only**: All API calls use HTTPS
