package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/joncooper/gday/internal/auth"
	gdaycal "github.com/joncooper/gday/internal/calendar"
	"github.com/spf13/cobra"
)

var calCmd = &cobra.Command{
	Use:     "cal",
	Aliases: []string{"c", "calendar"},
	Short:   "Google Calendar commands",
	Long:    `Commands for interacting with Google Calendar.`,
}

var calListCmd = &cobra.Command{
	Use:   "list",
	Short: "List upcoming events",
	Long: `List upcoming calendar events.

Examples:
  gday cal list                    # List next 10 events
  gday cal list -n 20              # List next 20 events
  gday cal list --days 30          # Events in next 30 days
  gday cal list --calendar work    # Events from specific calendar
  gday cal list --all-calendars    # Events from all calendars`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, err := auth.GetClient(ctx)
		if err != nil {
			exitError("%v", err)
		}

		srv, err := gdaycal.NewService(ctx, client)
		if err != nil {
			exitError("%v", err)
		}

		n, _ := cmd.Flags().GetInt64("number")
		days, _ := cmd.Flags().GetInt("days")
		calID, _ := cmd.Flags().GetString("calendar")
		allCals, _ := cmd.Flags().GetBool("all-calendars")

		now := time.Now()
		timeMin := now
		timeMax := now.AddDate(0, 0, days)

		var events []*gdaycal.Event
		if allCals {
			events, err = srv.ListEventsFromAllCalendars(ctx, timeMin, timeMax, n)
		} else {
			events, err = srv.ListEvents(ctx, calID, timeMin, timeMax, n)
		}
		if err != nil {
			exitError("%v", err)
		}

		if isJSONOutput() {
			outputJSON(eventsToJSON(events))
			return
		}

		if len(events) == 0 {
			fmt.Println("No upcoming events")
			return
		}

		printEvents(events)
	},
}

var calTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's events",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, err := auth.GetClient(ctx)
		if err != nil {
			exitError("%v", err)
		}

		srv, err := gdaycal.NewService(ctx, client)
		if err != nil {
			exitError("%v", err)
		}

		calID, _ := cmd.Flags().GetString("calendar")
		events, err := srv.Today(ctx, calID)
		if err != nil {
			exitError("%v", err)
		}

		if isJSONOutput() {
			outputJSON(eventsToJSON(events))
			return
		}

		if len(events) == 0 {
			fmt.Println("No events today")
			return
		}

		fmt.Println("Today's events:")
		fmt.Println()
		printEvents(events)
	},
}

var calTomorrowCmd = &cobra.Command{
	Use:   "tomorrow",
	Short: "Show tomorrow's events",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, err := auth.GetClient(ctx)
		if err != nil {
			exitError("%v", err)
		}

		srv, err := gdaycal.NewService(ctx, client)
		if err != nil {
			exitError("%v", err)
		}

		calID, _ := cmd.Flags().GetString("calendar")
		events, err := srv.Tomorrow(ctx, calID)
		if err != nil {
			exitError("%v", err)
		}

		if isJSONOutput() {
			outputJSON(eventsToJSON(events))
			return
		}

		if len(events) == 0 {
			fmt.Println("No events tomorrow")
			return
		}

		fmt.Println("Tomorrow's events:")
		fmt.Println()
		printEvents(events)
	},
}

var calWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Show this week's events",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, err := auth.GetClient(ctx)
		if err != nil {
			exitError("%v", err)
		}

		srv, err := gdaycal.NewService(ctx, client)
		if err != nil {
			exitError("%v", err)
		}

		calID, _ := cmd.Flags().GetString("calendar")
		events, err := srv.Week(ctx, calID)
		if err != nil {
			exitError("%v", err)
		}

		if isJSONOutput() {
			outputJSON(eventsToJSON(events))
			return
		}

		if len(events) == 0 {
			fmt.Println("No events this week")
			return
		}

		fmt.Println("This week's events:")
		fmt.Println()
		printEvents(events)
	},
}

var calShowCmd = &cobra.Command{
	Use:   "show <event-id>",
	Short: "Show event details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, err := auth.GetClient(ctx)
		if err != nil {
			exitError("%v", err)
		}

		srv, err := gdaycal.NewService(ctx, client)
		if err != nil {
			exitError("%v", err)
		}

		eventID := args[0]
		calID, _ := cmd.Flags().GetString("calendar")

		event, err := srv.GetEvent(ctx, calID, eventID)
		if err != nil {
			exitError("%v", err)
		}

		if isJSONOutput() {
			outputJSON(eventToJSON(event))
			return
		}

		printEventDetails(event)
	},
}

var calCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new event",
	Long: `Create a new calendar event.

Examples:
  gday cal create --title "Meeting" --start "2024-01-15 14:00" --end "2024-01-15 15:00"
  gday cal create --title "Birthday" --date "2024-01-20" --all-day
  gday cal create --quick "Lunch with John tomorrow at noon"`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, err := auth.GetClient(ctx)
		if err != nil {
			exitError("%v", err)
		}

		srv, err := gdaycal.NewService(ctx, client)
		if err != nil {
			exitError("%v", err)
		}

		calID, _ := cmd.Flags().GetString("calendar")
		quick, _ := cmd.Flags().GetString("quick")

		// Quick add mode
		if quick != "" {
			event, err := srv.QuickAdd(ctx, calID, quick)
			if err != nil {
				exitError("%v", err)
			}
			if isJSONOutput() {
				outputJSON(EventCreatedJSON{ID: event.ID, Summary: event.Summary, HtmlLink: event.HtmlLink, Status: "created"})
				return
			}
			fmt.Printf("Event created: %s\n", event.Summary)
			fmt.Printf("ID: %s\n", event.ID)
			if !event.AllDay {
				fmt.Printf("Time: %s - %s\n",
					event.Start.Format("Mon Jan 2, 3:04 PM"),
					event.End.Format("3:04 PM"))
			} else {
				fmt.Printf("Date: %s\n", event.Start.Format("Mon Jan 2, 2006"))
			}
			return
		}

		// Manual event creation
		title, _ := cmd.Flags().GetString("title")
		startStr, _ := cmd.Flags().GetString("start")
		endStr, _ := cmd.Flags().GetString("end")
		dateStr, _ := cmd.Flags().GetString("date")
		allDay, _ := cmd.Flags().GetBool("all-day")
		location, _ := cmd.Flags().GetString("location")
		description, _ := cmd.Flags().GetString("description")
		attendees, _ := cmd.Flags().GetStringSlice("attendees")

		if title == "" {
			exitError("--title or --quick is required")
		}

		event := &gdaycal.Event{
			Summary:     title,
			Location:    location,
			Description: description,
			Attendees:   attendees,
		}

		if allDay || dateStr != "" {
			event.AllDay = true
			if dateStr != "" {
				t, err := parseDate(dateStr)
				if err != nil {
					exitError("invalid date format: %v", err)
				}
				event.Start = t
				event.End = t.AddDate(0, 0, 1)
			} else {
				exitError("--date is required for all-day events")
			}
		} else {
			if startStr == "" {
				exitError("--start is required (or use --quick)")
			}
			start, err := parseDateTime(startStr)
			if err != nil {
				exitError("invalid start time: %v", err)
			}
			event.Start = start

			if endStr != "" {
				end, err := parseDateTime(endStr)
				if err != nil {
					exitError("invalid end time: %v", err)
				}
				event.End = end
			} else {
				// Default to 1 hour duration
				event.End = start.Add(time.Hour)
			}
		}

		created, err := srv.CreateEvent(ctx, calID, event)
		if err != nil {
			exitError("%v", err)
		}

		if isJSONOutput() {
			outputJSON(EventCreatedJSON{ID: created.ID, Summary: created.Summary, HtmlLink: created.HtmlLink, Status: "created"})
			return
		}

		fmt.Printf("Event created: %s\n", created.Summary)
		fmt.Printf("ID: %s\n", created.ID)
		if created.HtmlLink != "" {
			fmt.Printf("Link: %s\n", created.HtmlLink)
		}
	},
}

var calDeleteCmd = &cobra.Command{
	Use:   "delete <event-id>",
	Short: "Delete an event",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, err := auth.GetClient(ctx)
		if err != nil {
			exitError("%v", err)
		}

		srv, err := gdaycal.NewService(ctx, client)
		if err != nil {
			exitError("%v", err)
		}

		eventID := args[0]
		calID, _ := cmd.Flags().GetString("calendar")

		if err := srv.DeleteEvent(ctx, calID, eventID); err != nil {
			exitError("%v", err)
		}

		if isJSONOutput() {
			outputJSON(StatusJSON{Status: "deleted", Message: "Event deleted successfully"})
			return
		}

		fmt.Println("Event deleted")
	},
}

var calSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for events",
	Long: `Search for calendar events matching a query.

Examples:
  gday cal search "meeting"
  gday cal search "John" --days 90`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, err := auth.GetClient(ctx)
		if err != nil {
			exitError("%v", err)
		}

		srv, err := gdaycal.NewService(ctx, client)
		if err != nil {
			exitError("%v", err)
		}

		query := strings.Join(args, " ")
		calID, _ := cmd.Flags().GetString("calendar")
		days, _ := cmd.Flags().GetInt("days")
		n, _ := cmd.Flags().GetInt64("number")

		now := time.Now()
		timeMin := now.AddDate(0, 0, -30) // Search past 30 days too
		timeMax := now.AddDate(0, 0, days)

		events, err := srv.SearchEvents(ctx, calID, query, timeMin, timeMax, n)
		if err != nil {
			exitError("%v", err)
		}

		if isJSONOutput() {
			outputJSON(eventsToJSON(events))
			return
		}

		if len(events) == 0 {
			fmt.Println("No events found")
			return
		}

		fmt.Printf("Found %d events matching: %s\n\n", len(events), query)
		printEvents(events)
	},
}

var calCalendarsCmd = &cobra.Command{
	Use:   "calendars",
	Short: "List all calendars",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client, err := auth.GetClient(ctx)
		if err != nil {
			exitError("%v", err)
		}

		srv, err := gdaycal.NewService(ctx, client)
		if err != nil {
			exitError("%v", err)
		}

		calendars, err := srv.ListCalendars(ctx)
		if err != nil {
			exitError("%v", err)
		}

		if isJSONOutput() {
			jsonCals := make([]CalendarJSON, 0, len(calendars))
			for _, c := range calendars {
				jsonCals = append(jsonCals, CalendarJSON{
					ID:          c.ID,
					Summary:     c.Summary,
					Description: c.Description,
					Primary:     c.Primary,
				})
			}
			outputJSON(CalendarsListJSON{Calendars: jsonCals})
			return
		}

		fmt.Println("Calendars:")
		for _, c := range calendars {
			primary := ""
			if c.Primary {
				primary = " (primary)"
			}
			fmt.Printf("  %-40s %s%s\n", c.Summary, c.ID[:min(30, len(c.ID))], primary)
		}
	},
}

func init() {
	rootCmd.AddCommand(calCmd)

	// Global calendar flag
	calCmd.PersistentFlags().StringP("calendar", "c", "", "Calendar ID (default: primary)")

	// List command
	calCmd.AddCommand(calListCmd)
	calListCmd.Flags().Int64P("number", "n", 10, "Maximum number of events")
	calListCmd.Flags().Int("days", 14, "Number of days to look ahead")
	calListCmd.Flags().Bool("all-calendars", false, "Include events from all calendars")

	// Today command
	calCmd.AddCommand(calTodayCmd)

	// Tomorrow command
	calCmd.AddCommand(calTomorrowCmd)

	// Week command
	calCmd.AddCommand(calWeekCmd)

	// Show command
	calCmd.AddCommand(calShowCmd)

	// Create command
	calCmd.AddCommand(calCreateCmd)
	calCreateCmd.Flags().StringP("title", "t", "", "Event title")
	calCreateCmd.Flags().StringP("start", "s", "", "Start time (YYYY-MM-DD HH:MM)")
	calCreateCmd.Flags().StringP("end", "e", "", "End time (YYYY-MM-DD HH:MM)")
	calCreateCmd.Flags().String("date", "", "Date for all-day events (YYYY-MM-DD)")
	calCreateCmd.Flags().Bool("all-day", false, "Create all-day event")
	calCreateCmd.Flags().StringP("location", "l", "", "Event location")
	calCreateCmd.Flags().StringP("description", "d", "", "Event description")
	calCreateCmd.Flags().StringSlice("attendees", nil, "Event attendees (emails)")
	calCreateCmd.Flags().StringP("quick", "q", "", "Quick add using natural language")

	// Delete command
	calCmd.AddCommand(calDeleteCmd)

	// Search command
	calCmd.AddCommand(calSearchCmd)
	calSearchCmd.Flags().Int("days", 90, "Number of days to search")
	calSearchCmd.Flags().Int64P("number", "n", 20, "Maximum number of results")

	// Calendars command
	calCmd.AddCommand(calCalendarsCmd)
}

// Helper functions

func printEvents(events []*gdaycal.Event) {
	currentDate := ""
	for _, e := range events {
		dateStr := e.Start.Format("Mon Jan 2")
		if dateStr != currentDate {
			if currentDate != "" {
				fmt.Println()
			}
			fmt.Printf("%s\n", dateStr)
			currentDate = dateStr
		}

		if e.AllDay {
			fmt.Printf("  All day    %s\n", e.Summary)
		} else {
			fmt.Printf("  %s - %s  %s\n",
				e.Start.Format("15:04"),
				e.End.Format("15:04"),
				e.Summary)
		}
	}
}

func printEventDetails(e *gdaycal.Event) {
	fmt.Printf("Event: %s\n", e.Summary)
	fmt.Printf("ID: %s\n", e.ID)

	if e.AllDay {
		fmt.Printf("Date: %s (all day)\n", e.Start.Format("Mon Jan 2, 2006"))
	} else {
		fmt.Printf("Start: %s\n", e.Start.Format("Mon Jan 2, 2006 at 3:04 PM"))
		fmt.Printf("End: %s\n", e.End.Format("Mon Jan 2, 2006 at 3:04 PM"))
	}

	if e.Location != "" {
		fmt.Printf("Location: %s\n", e.Location)
	}

	if len(e.Attendees) > 0 {
		fmt.Printf("Attendees: %s\n", strings.Join(e.Attendees, ", "))
	}

	if e.Description != "" {
		fmt.Printf("\nDescription:\n%s\n", e.Description)
	}

	if e.HtmlLink != "" {
		fmt.Printf("\nLink: %s\n", e.HtmlLink)
	}
}

func parseDateTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"01/02/2006 15:04",
		"01/02/2006 3:04 PM",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, s, time.Local); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse datetime: %s", s)
}

func parseDate(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"Jan 2, 2006",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, s, time.Local); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// eventToJSON converts a calendar.Event to EventJSON
func eventToJSON(e *gdaycal.Event) EventJSON {
	return EventJSON{
		ID:          e.ID,
		CalendarID:  e.CalendarID,
		Summary:     e.Summary,
		Description: e.Description,
		Location:    e.Location,
		Start:       e.Start,
		End:         e.End,
		AllDay:      e.AllDay,
		Attendees:   e.Attendees,
		Status:      e.Status,
		HtmlLink:    e.HtmlLink,
		Recurring:   e.Recurring,
	}
}

// eventsToJSON converts a slice of events to JSON format
func eventsToJSON(events []*gdaycal.Event) EventsListJSON {
	jsonEvents := make([]EventJSON, 0, len(events))
	for _, e := range events {
		jsonEvents = append(jsonEvents, eventToJSON(e))
	}
	return EventsListJSON{Count: len(jsonEvents), Events: jsonEvents}
}
