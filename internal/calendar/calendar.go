package calendar

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Service wraps the Google Calendar API service
type Service struct {
	srv *calendar.Service
}

// Event represents a simplified calendar event
type Event struct {
	ID           string
	CalendarID   string
	Summary      string
	Description  string
	Location     string
	Start        time.Time
	End          time.Time
	AllDay       bool
	Attendees    []string
	Status       string
	HtmlLink     string
	Recurring    bool
	RecurrenceID string
}

// Calendar represents a calendar
type Calendar struct {
	ID          string
	Summary     string
	Description string
	Primary     bool
	Color       string
}

// NewService creates a new Calendar service
func NewService(ctx context.Context, client *http.Client) (*Service, error) {
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Calendar service: %w", err)
	}
	return &Service{srv: srv}, nil
}

// ListCalendars returns all calendars the user has access to
func (s *Service) ListCalendars(ctx context.Context) ([]*Calendar, error) {
	resp, err := s.srv.CalendarList.List().Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list calendars: %w", err)
	}

	calendars := make([]*Calendar, 0, len(resp.Items))
	for _, c := range resp.Items {
		calendars = append(calendars, &Calendar{
			ID:          c.Id,
			Summary:     c.Summary,
			Description: c.Description,
			Primary:     c.Primary,
			Color:       c.BackgroundColor,
		})
	}

	return calendars, nil
}

// ListEvents lists events from a calendar
func (s *Service) ListEvents(ctx context.Context, calendarID string, timeMin, timeMax time.Time, maxResults int64) ([]*Event, error) {
	if calendarID == "" {
		calendarID = "primary"
	}

	req := s.srv.Events.List(calendarID).
		SingleEvents(true).
		OrderBy("startTime").
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339))

	if maxResults > 0 {
		req = req.MaxResults(maxResults)
	}

	resp, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	events := make([]*Event, 0, len(resp.Items))
	for _, e := range resp.Items {
		events = append(events, parseEvent(e, calendarID))
	}

	return events, nil
}

// ListEventsFromAllCalendars lists events from all calendars
func (s *Service) ListEventsFromAllCalendars(ctx context.Context, timeMin, timeMax time.Time, maxResults int64) ([]*Event, error) {
	calendars, err := s.ListCalendars(ctx)
	if err != nil {
		return nil, err
	}

	var allEvents []*Event
	for _, cal := range calendars {
		events, err := s.ListEvents(ctx, cal.ID, timeMin, timeMax, 0)
		if err != nil {
			// Skip calendars that fail (e.g., no access)
			continue
		}
		allEvents = append(allEvents, events...)
	}

	// Sort by start time
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Start.Before(allEvents[j].Start)
	})

	// Apply max results limit
	if maxResults > 0 && int64(len(allEvents)) > maxResults {
		allEvents = allEvents[:maxResults]
	}

	return allEvents, nil
}

// GetEvent retrieves a single event
func (s *Service) GetEvent(ctx context.Context, calendarID, eventID string) (*Event, error) {
	if calendarID == "" {
		calendarID = "primary"
	}

	e, err := s.srv.Events.Get(calendarID, eventID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return parseEvent(e, calendarID), nil
}

// CreateEvent creates a new calendar event
func (s *Service) CreateEvent(ctx context.Context, calendarID string, event *Event) (*Event, error) {
	if calendarID == "" {
		calendarID = "primary"
	}

	e := &calendar.Event{
		Summary:     event.Summary,
		Description: event.Description,
		Location:    event.Location,
	}

	if event.AllDay {
		e.Start = &calendar.EventDateTime{
			Date: event.Start.Format("2006-01-02"),
		}
		e.End = &calendar.EventDateTime{
			Date: event.End.Format("2006-01-02"),
		}
	} else {
		e.Start = &calendar.EventDateTime{
			DateTime: event.Start.Format(time.RFC3339),
			TimeZone: event.Start.Location().String(),
		}
		e.End = &calendar.EventDateTime{
			DateTime: event.End.Format(time.RFC3339),
			TimeZone: event.End.Location().String(),
		}
	}

	// Add attendees
	for _, email := range event.Attendees {
		e.Attendees = append(e.Attendees, &calendar.EventAttendee{
			Email: email,
		})
	}

	created, err := s.srv.Events.Insert(calendarID, e).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return parseEvent(created, calendarID), nil
}

// UpdateEvent updates an existing event
func (s *Service) UpdateEvent(ctx context.Context, calendarID, eventID string, event *Event) (*Event, error) {
	if calendarID == "" {
		calendarID = "primary"
	}

	e := &calendar.Event{
		Summary:     event.Summary,
		Description: event.Description,
		Location:    event.Location,
	}

	if event.AllDay {
		e.Start = &calendar.EventDateTime{
			Date: event.Start.Format("2006-01-02"),
		}
		e.End = &calendar.EventDateTime{
			Date: event.End.Format("2006-01-02"),
		}
	} else {
		e.Start = &calendar.EventDateTime{
			DateTime: event.Start.Format(time.RFC3339),
		}
		e.End = &calendar.EventDateTime{
			DateTime: event.End.Format(time.RFC3339),
		}
	}

	updated, err := s.srv.Events.Update(calendarID, eventID, e).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return parseEvent(updated, calendarID), nil
}

// DeleteEvent deletes an event
func (s *Service) DeleteEvent(ctx context.Context, calendarID, eventID string) error {
	if calendarID == "" {
		calendarID = "primary"
	}

	if err := s.srv.Events.Delete(calendarID, eventID).Do(); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	return nil
}

// SearchEvents searches for events matching a query
func (s *Service) SearchEvents(ctx context.Context, calendarID, query string, timeMin, timeMax time.Time, maxResults int64) ([]*Event, error) {
	if calendarID == "" {
		calendarID = "primary"
	}

	req := s.srv.Events.List(calendarID).
		SingleEvents(true).
		OrderBy("startTime").
		Q(query).
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339))

	if maxResults > 0 {
		req = req.MaxResults(maxResults)
	}

	resp, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}

	events := make([]*Event, 0, len(resp.Items))
	for _, e := range resp.Items {
		events = append(events, parseEvent(e, calendarID))
	}

	return events, nil
}

// QuickAdd creates an event using natural language
func (s *Service) QuickAdd(ctx context.Context, calendarID, text string) (*Event, error) {
	if calendarID == "" {
		calendarID = "primary"
	}

	created, err := s.srv.Events.QuickAdd(calendarID, text).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to quick add event: %w", err)
	}

	return parseEvent(created, calendarID), nil
}

// Today returns events for today
func (s *Service) Today(ctx context.Context, calendarID string) ([]*Event, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	return s.ListEvents(ctx, calendarID, startOfDay, endOfDay, 0)
}

// Tomorrow returns events for tomorrow
func (s *Service) Tomorrow(ctx context.Context, calendarID string) ([]*Event, error) {
	now := time.Now()
	startOfTomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	endOfTomorrow := startOfTomorrow.Add(24 * time.Hour)
	return s.ListEvents(ctx, calendarID, startOfTomorrow, endOfTomorrow, 0)
}

// Week returns events for the next 7 days
func (s *Service) Week(ctx context.Context, calendarID string) ([]*Event, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfWeek := startOfDay.Add(7 * 24 * time.Hour)
	return s.ListEvents(ctx, calendarID, startOfDay, endOfWeek, 0)
}

// parseEvent converts a calendar.Event to our Event type
func parseEvent(e *calendar.Event, calendarID string) *Event {
	event := &Event{
		ID:          e.Id,
		CalendarID:  calendarID,
		Summary:     e.Summary,
		Description: e.Description,
		Location:    e.Location,
		Status:      e.Status,
		HtmlLink:    e.HtmlLink,
	}

	// Parse start time
	if e.Start != nil {
		if e.Start.Date != "" {
			// All-day event
			event.AllDay = true
			t, _ := time.Parse("2006-01-02", e.Start.Date)
			event.Start = t
		} else {
			t, _ := time.Parse(time.RFC3339, e.Start.DateTime)
			event.Start = t
		}
	}

	// Parse end time
	if e.End != nil {
		if e.End.Date != "" {
			t, _ := time.Parse("2006-01-02", e.End.Date)
			event.End = t
		} else {
			t, _ := time.Parse(time.RFC3339, e.End.DateTime)
			event.End = t
		}
	}

	// Parse attendees
	for _, a := range e.Attendees {
		event.Attendees = append(event.Attendees, a.Email)
	}

	// Check if recurring
	if e.RecurringEventId != "" {
		event.Recurring = true
		event.RecurrenceID = e.RecurringEventId
	}

	return event
}
