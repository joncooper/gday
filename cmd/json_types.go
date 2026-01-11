package cmd

import "time"

// JSON output types for Gmail

// MessageJSON represents a message in JSON output
type MessageJSON struct {
	ID        string           `json:"id"`
	ThreadID  string           `json:"thread_id"`
	Date      time.Time        `json:"date"`
	From      string           `json:"from"`
	To        string           `json:"to"`
	Subject   string           `json:"subject"`
	Snippet   string           `json:"snippet,omitempty"`
	Body      string           `json:"body,omitempty"`
	Labels    []string         `json:"labels,omitempty"`
	IsUnread  bool             `json:"is_unread"`
	Attachments []AttachmentJSON `json:"attachments,omitempty"`
}

// AttachmentJSON represents an attachment in JSON output
type AttachmentJSON struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

// MessagesListJSON represents a list of messages
type MessagesListJSON struct {
	Count    int           `json:"count"`
	Messages []MessageJSON `json:"messages"`
}

// ThreadJSON represents a thread in JSON output
type ThreadJSON struct {
	ThreadID string        `json:"thread_id"`
	Count    int           `json:"count"`
	Messages []MessageJSON `json:"messages"`
}

// SearchResultJSON represents search results
type SearchResultJSON struct {
	Query    string        `json:"query"`
	Count    int           `json:"count"`
	Messages []MessageJSON `json:"messages"`
}

// SendResultJSON represents the result of sending a message
type SendResultJSON struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"`
}

// LabelsJSON represents labels list
type LabelsJSON struct {
	Labels []string `json:"labels"`
}

// JSON output types for Calendar

// EventJSON represents a calendar event in JSON output
type EventJSON struct {
	ID          string    `json:"id"`
	CalendarID  string    `json:"calendar_id,omitempty"`
	Summary     string    `json:"summary"`
	Description string    `json:"description,omitempty"`
	Location    string    `json:"location,omitempty"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	AllDay      bool      `json:"all_day"`
	Attendees   []string  `json:"attendees,omitempty"`
	Status      string    `json:"status,omitempty"`
	HtmlLink    string    `json:"html_link,omitempty"`
	Recurring   bool      `json:"recurring"`
}

// EventsListJSON represents a list of events
type EventsListJSON struct {
	Count  int         `json:"count"`
	Events []EventJSON `json:"events"`
}

// CalendarJSON represents a calendar in JSON output
type CalendarJSON struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"`
	Primary     bool   `json:"primary"`
}

// CalendarsListJSON represents a list of calendars
type CalendarsListJSON struct {
	Calendars []CalendarJSON `json:"calendars"`
}

// EventCreatedJSON represents the result of creating an event
type EventCreatedJSON struct {
	ID       string `json:"id"`
	Summary  string `json:"summary"`
	HtmlLink string `json:"html_link,omitempty"`
	Status   string `json:"status"`
}

// StatusJSON for simple status messages
type StatusJSON struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
