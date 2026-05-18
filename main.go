package main

/*
#cgo LDFLAGS: -lsqlite3
#include <sqlite3.h>
#include <stdlib.h>

static int bind_text_transient(sqlite3_stmt *stmt, int idx, const char *text) {
    return sqlite3_bind_text(stmt, idx, text, -1, SQLITE_TRANSIENT);
}
*/
import "C"

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

type Application struct {
	ID              int64
	Company         string
	Role            string
	Location        string
	PayMin          *float64
	PayMax          *float64
	Link            string
	Notes           string
	AppOpenDate     string
	AppCloseDate    string
	InternStartDate string
	InternEndDate   string
	LatestStatus    string
}

type ApplicationEvent struct {
	ID            int64
	ApplicationID int64
	EventType     string
	EventDate     string
	Details       string
}

type Stats struct {
	Total      int
	Rejected   int
	Offers     int
	InProgress int
}

type DashboardPage struct {
	Apps  []Application
	Stats Stats
}

type AppPage struct {
	App        Application
	Events     []ApplicationEvent
	EventTypes []string
	Today      string
}

type Server struct {
	db        *DB
	templates *template.Template
}

var eventTypes = []string{
	"APPLIED",
	"OA_SENT",
	"OA_DONE",
	"HIREVUE",
	"PHONE_SCREEN",
	"INTERVIEW",
	"OFFER",
	"REJECTED",
	"WITHDRAWN",
	"CLOSED",
}

func main() {
	dbPath := os.Getenv("DATABASE_URL")
	if dbPath == "" {
		dbPath = "internship-tracker.sqlite"
	}

	db, err := OpenDB(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := db.InitSchema(); err != nil {
		log.Fatalf("initialize database schema: %v", err)
	}

	srv := &Server{
		db:        db,
		templates: loadTemplates(),
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", srv.handleIndex)
	mux.HandleFunc("/add", srv.handleAdd)
	mux.HandleFunc("/app/", srv.handleApp)
	mux.HandleFunc("/calendar.ics", srv.handleCalendar)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":5173"
	}

	log.Printf("Internship Tracker listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, securityHeaders(mux)); err != nil {
		log.Fatal(err)
	}
}

func loadTemplates() *template.Template {
	funcs := template.FuncMap{
		"payText":       payText,
		"payClass":      payClass,
		"statusClass":   statusClass,
		"dateRange":     dateRange,
		"fallback":      fallback,
		"formatDollars": formatDollars,
	}

	return template.Must(template.New("app").Funcs(funcs).ParseGlob(filepath.Join("templates", "*.gohtml")))
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) render(w http.ResponseWriter, status int, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("render %s: %v", name, err)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	apps, err := s.db.ListApplicationsWithLatestStatus()
	if err != nil {
		serverError(w, err)
		return
	}
	stats, err := s.db.Stats()
	if err != nil {
		serverError(w, err)
		return
	}

	s.render(w, http.StatusOK, "index", DashboardPage{Apps: apps, Stats: stats})
}

func (s *Server) handleAdd(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.render(w, http.StatusOK, "add", nil)
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			badRequest(w, "Could not parse form")
			return
		}

		app := Application{
			Company:         strings.TrimSpace(r.FormValue("company")),
			Role:            strings.TrimSpace(r.FormValue("role")),
			Location:        strings.TrimSpace(r.FormValue("location")),
			Link:            strings.TrimSpace(r.FormValue("link")),
			Notes:           strings.TrimSpace(r.FormValue("notes")),
			PayMin:          parseOptionalFloat(r.FormValue("payMin")),
			PayMax:          parseOptionalFloat(r.FormValue("payMax")),
			AppOpenDate:     r.FormValue("appOpenDate"),
			AppCloseDate:    r.FormValue("appCloseDate"),
			InternStartDate: r.FormValue("internStartDate"),
			InternEndDate:   r.FormValue("internEndDate"),
		}

		if app.Company == "" || app.Role == "" {
			badRequest(w, "Company and role are required")
			return
		}

		if _, err := s.db.CreateApplication(app); err != nil {
			serverError(w, err)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		methodNotAllowed(w)
	}
}

func (s *Server) handleApp(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.Trim(strings.TrimPrefix(r.URL.Path, "/app/"), "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	if r.Method == http.MethodGet && action == "" {
		s.showApp(w, r, id)
		return
	}
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	switch action {
	case "update":
		s.updateApp(w, r, id)
	case "event":
		s.addEvent(w, r, id)
	case "delete":
		s.deleteApp(w, r, id)
	case "delete-event":
		s.deleteEvent(w, r, id)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) showApp(w http.ResponseWriter, r *http.Request, id int64) {
	app, err := s.db.GetApplication(id)
	if err != nil {
		if errors.Is(err, errNotFound) {
			http.NotFound(w, r)
			return
		}
		serverError(w, err)
		return
	}
	events, err := s.db.ListEvents(id)
	if err != nil {
		serverError(w, err)
		return
	}

	s.render(w, http.StatusOK, "app_detail", AppPage{
		App:        app,
		Events:     events,
		EventTypes: eventTypes,
		Today:      time.Now().UTC().Format("2006-01-02"),
	})
}

func (s *Server) updateApp(w http.ResponseWriter, r *http.Request, id int64) {
	if err := r.ParseForm(); err != nil {
		badRequest(w, "Could not parse form")
		return
	}
	app := Application{
		ID:              id,
		Company:         strings.TrimSpace(r.FormValue("company")),
		Role:            strings.TrimSpace(r.FormValue("role")),
		Location:        strings.TrimSpace(r.FormValue("location")),
		Link:            strings.TrimSpace(r.FormValue("link")),
		Notes:           strings.TrimSpace(r.FormValue("notes")),
		PayMin:          parseOptionalFloat(r.FormValue("payMin")),
		PayMax:          parseOptionalFloat(r.FormValue("payMax")),
		AppOpenDate:     r.FormValue("appOpenDate"),
		AppCloseDate:    r.FormValue("appCloseDate"),
		InternStartDate: r.FormValue("internStartDate"),
		InternEndDate:   r.FormValue("internEndDate"),
	}
	if app.Company == "" || app.Role == "" {
		badRequest(w, "Company and role are required")
		return
	}
	if err := s.db.UpdateApplication(app); err != nil {
		serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/app/%d", id), http.StatusSeeOther)
}

func (s *Server) addEvent(w http.ResponseWriter, r *http.Request, id int64) {
	if err := r.ParseForm(); err != nil {
		badRequest(w, "Could not parse form")
		return
	}
	eventType := r.FormValue("eventType")
	if !isEventType(eventType) {
		badRequest(w, "Invalid event type")
		return
	}
	eventDate := r.FormValue("eventDate")
	if eventDate == "" {
		eventDate = time.Now().UTC().Format("2006-01-02")
	}

	err := s.db.CreateEvent(ApplicationEvent{
		ApplicationID: id,
		EventType:     eventType,
		EventDate:     eventDate,
		Details:       strings.TrimSpace(r.FormValue("details")),
	})
	if err != nil {
		serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/app/%d", id), http.StatusSeeOther)
}

func (s *Server) deleteApp(w http.ResponseWriter, r *http.Request, id int64) {
	if err := s.db.DeleteApplication(id); err != nil {
		serverError(w, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) deleteEvent(w http.ResponseWriter, r *http.Request, id int64) {
	if err := r.ParseForm(); err != nil {
		badRequest(w, "Could not parse form")
		return
	}
	eventID, err := strconv.ParseInt(r.FormValue("eventId"), 10, 64)
	if err != nil || eventID < 1 {
		badRequest(w, "Invalid event ID")
		return
	}
	if err := s.db.DeleteEvent(eventID, id); err != nil {
		serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/app/%d", id), http.StatusSeeOther)
}

func (s *Server) handleCalendar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	interviews, err := s.db.ListInterviewEvents()
	if err != nil {
		serverError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "inline; filename=calendar.ics")
	_, _ = w.Write([]byte(buildICS(interviews)))
}

func buildICS(events []ApplicationEvent) string {
	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\n")
	b.WriteString("VERSION:2.0\r\n")
	b.WriteString("PRODID:-//Internship Tracker//Go//EN\r\n")
	b.WriteString("CALSCALE:GREGORIAN\r\n")
	b.WriteString("METHOD:PUBLISH\r\n")
	now := time.Now().UTC().Format("20060102T150405Z")
	for _, ev := range events {
		date, err := time.Parse("2006-01-02", ev.EventDate)
		if err != nil {
			continue
		}
		start := time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, time.UTC)
		end := start.Add(time.Hour)
		uid := stableUID(ev.ID)
		title := "Interview: Internship"
		if ev.Details != "" {
			title = "Interview: " + ev.Details
		}
		b.WriteString("BEGIN:VEVENT\r\n")
		b.WriteString("UID:" + escapeICal(uid) + "\r\n")
		b.WriteString("DTSTAMP:" + now + "\r\n")
		b.WriteString("DTSTART:" + start.Format("20060102T150405Z") + "\r\n")
		b.WriteString("DTEND:" + end.Format("20060102T150405Z") + "\r\n")
		b.WriteString("SUMMARY:" + escapeICal(title) + "\r\n")
		b.WriteString("DESCRIPTION:" + escapeICal("Event details: "+ev.Details) + "\r\n")
		b.WriteString("END:VEVENT\r\n")
	}
	b.WriteString("END:VCALENDAR\r\n")
	return b.String()
}

func stableUID(id int64) string {
	if id > 0 {
		return fmt.Sprintf("application-event-%d@internship-tracker", id)
	}
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err == nil {
		return hex.EncodeToString(buf) + "@internship-tracker"
	}
	return fmt.Sprintf("%d@internship-tracker", time.Now().UnixNano())
}

func escapeICal(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\r\n", "\\n")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func methodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Allow", "GET, POST")
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func badRequest(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusBadRequest)
}

func serverError(w http.ResponseWriter, err error) {
	log.Printf("server error: %v", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func parseOptionalFloat(raw string) *float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil
	}
	return &v
}

func isEventType(v string) bool {
	for _, typ := range eventTypes {
		if v == typ {
			return true
		}
	}
	return false
}

func fallback(v, fb string) string {
	if strings.TrimSpace(v) == "" {
		return fb
	}
	return v
}

func payText(min, max *float64) string {
	if min != nil && max != nil {
		return "$" + formatDollars(*min) + " - $" + formatDollars(*max)
	}
	if min != nil {
		return "$" + formatDollars(*min)
	}
	if max != nil {
		return "$" + formatDollars(*max)
	}
	return "???"
}

func payClass(min, max *float64) string {
	if min != nil && max != nil {
		return "pay-good"
	}
	if min != nil || max != nil {
		return "pay-partial"
	}
	return "pay-missing"
}

func formatDollars(v any) string {
	var n float64
	switch typed := v.(type) {
	case float64:
		n = typed
	case *float64:
		if typed == nil {
			return ""
		}
		n = *typed
	default:
		return ""
	}
	if n == float64(int64(n)) {
		return strconv.FormatInt(int64(n), 10)
	}
	return strconv.FormatFloat(n, 'f', 2, 64)
}

func statusClass(status string) string {
	switch status {
	case "REJECTED":
		return "status-rejected"
	case "INTERVIEW":
		return "status-interview"
	case "OFFER":
		return "status-offer"
	case "":
		return "status-unknown"
	default:
		return "status-active"
	}
}

func dateRange(start, end string) string {
	if start == "" {
		start = "?"
	}
	if end == "" {
		end = "?"
	}
	return start + " - " + end
}

var errNotFound = errors.New("not found")

type DB struct {
	mu   sync.Mutex
	conn *C.sqlite3
}

func OpenDB(path string) (*DB, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var conn *C.sqlite3
	flags := C.SQLITE_OPEN_READWRITE | C.SQLITE_OPEN_CREATE | C.SQLITE_OPEN_FULLMUTEX
	if rc := C.sqlite3_open_v2(cPath, &conn, C.int(flags), nil); rc != C.SQLITE_OK {
		msg := "unable to open database"
		if conn != nil {
			msg = C.GoString(C.sqlite3_errmsg(conn))
			C.sqlite3_close(conn)
		}
		return nil, errors.New(msg)
	}
	return &DB{conn: conn}, nil
}

func (db *DB) Close() {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.conn != nil {
		C.sqlite3_close(db.conn)
		db.conn = nil
	}
}

func (db *DB) InitSchema() error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS application_events (
	id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	application_id integer NOT NULL,
	event_type text NOT NULL,
	event_date text NOT NULL,
	details text,
	FOREIGN KEY (application_id) REFERENCES applications(id) ON UPDATE no action ON DELETE no action
)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS applications (
	id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	company text NOT NULL,
	role text NOT NULL,
	location text,
	pay_min real,
	pay_max real,
	link text,
	notes text,
	app_open_date text,
	app_close_date text,
	intern_start_date text,
	intern_end_date text
)`)
	return err
}

func (db *DB) Exec(query string, args ...any) (int64, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.execLocked(query, args...)
}

func (db *DB) execLocked(query string, args ...any) (int64, error) {
	stmt, err := db.prepareLocked(query)
	if err != nil {
		return 0, err
	}
	defer C.sqlite3_finalize(stmt)

	if err := bindArgs(stmt, args...); err != nil {
		return 0, err
	}

	rc := C.sqlite3_step(stmt)
	if rc != C.SQLITE_DONE && rc != C.SQLITE_ROW {
		return 0, db.errLocked()
	}
	return int64(C.sqlite3_last_insert_rowid(db.conn)), nil
}

func (db *DB) prepareLocked(query string) (*C.sqlite3_stmt, error) {
	cQuery := C.CString(query)
	defer C.free(unsafe.Pointer(cQuery))

	var stmt *C.sqlite3_stmt
	if rc := C.sqlite3_prepare_v2(db.conn, cQuery, -1, &stmt, nil); rc != C.SQLITE_OK {
		return nil, db.errLocked()
	}
	return stmt, nil
}

func (db *DB) errLocked() error {
	if db.conn == nil {
		return errors.New("database is closed")
	}
	return errors.New(C.GoString(C.sqlite3_errmsg(db.conn)))
}

func bindArgs(stmt *C.sqlite3_stmt, args ...any) error {
	for i, arg := range args {
		idx := C.int(i + 1)
		var rc C.int
		switch v := arg.(type) {
		case nil:
			rc = C.sqlite3_bind_null(stmt, idx)
		case string:
			cText := C.CString(v)
			rc = C.bind_text_transient(stmt, idx, cText)
			C.free(unsafe.Pointer(cText))
		case int:
			rc = C.sqlite3_bind_int64(stmt, idx, C.sqlite3_int64(v))
		case int64:
			rc = C.sqlite3_bind_int64(stmt, idx, C.sqlite3_int64(v))
		case float64:
			rc = C.sqlite3_bind_double(stmt, idx, C.double(v))
		case *float64:
			if v == nil {
				rc = C.sqlite3_bind_null(stmt, idx)
			} else {
				rc = C.sqlite3_bind_double(stmt, idx, C.double(*v))
			}
		default:
			return fmt.Errorf("unsupported bind argument %d of type %T", i+1, arg)
		}
		if rc != C.SQLITE_OK {
			return fmt.Errorf("could not bind argument %d", i+1)
		}
	}
	return nil
}

func (db *DB) ListApplicationsWithLatestStatus() ([]Application, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	stmt, err := db.prepareLocked(`SELECT
	id, company, role, location, pay_min, pay_max, link, notes,
	app_open_date, app_close_date, intern_start_date, intern_end_date,
	COALESCE((SELECT event_type FROM application_events WHERE application_id = applications.id ORDER BY event_date DESC, id DESC LIMIT 1), '') AS latest_status
FROM applications
ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer C.sqlite3_finalize(stmt)

	var apps []Application
	for {
		rc := C.sqlite3_step(stmt)
		if rc == C.SQLITE_DONE {
			return apps, nil
		}
		if rc != C.SQLITE_ROW {
			return nil, db.errLocked()
		}
		apps = append(apps, scanApplication(stmt, true))
	}
}

func (db *DB) Stats() (Stats, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	total, err := db.scalarIntLocked(`SELECT COUNT(*) FROM applications`)
	if err != nil {
		return Stats{}, err
	}
	rejected, err := db.scalarIntLocked(`SELECT COUNT(*) FROM application_events WHERE event_type = 'REJECTED'`)
	if err != nil {
		return Stats{}, err
	}
	offers, err := db.scalarIntLocked(`SELECT COUNT(*) FROM application_events WHERE event_type = 'OFFER'`)
	if err != nil {
		return Stats{}, err
	}
	terminal, err := db.scalarIntLocked(`SELECT COUNT(*) FROM application_events WHERE event_type IN ('REJECTED', 'OFFER', 'WITHDRAWN')`)
	if err != nil {
		return Stats{}, err
	}
	return Stats{Total: total, Rejected: rejected, Offers: offers, InProgress: total - terminal}, nil
}

func (db *DB) scalarIntLocked(query string, args ...any) (int, error) {
	stmt, err := db.prepareLocked(query)
	if err != nil {
		return 0, err
	}
	defer C.sqlite3_finalize(stmt)
	if err := bindArgs(stmt, args...); err != nil {
		return 0, err
	}
	if rc := C.sqlite3_step(stmt); rc != C.SQLITE_ROW {
		if rc == C.SQLITE_DONE {
			return 0, nil
		}
		return 0, db.errLocked()
	}
	return int(C.sqlite3_column_int64(stmt, 0)), nil
}

func (db *DB) GetApplication(id int64) (Application, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	stmt, err := db.prepareLocked(`SELECT id, company, role, location, pay_min, pay_max, link, notes,
	app_open_date, app_close_date, intern_start_date, intern_end_date
FROM applications
WHERE id = ?`)
	if err != nil {
		return Application{}, err
	}
	defer C.sqlite3_finalize(stmt)
	if err := bindArgs(stmt, id); err != nil {
		return Application{}, err
	}

	rc := C.sqlite3_step(stmt)
	if rc == C.SQLITE_DONE {
		return Application{}, errNotFound
	}
	if rc != C.SQLITE_ROW {
		return Application{}, db.errLocked()
	}
	return scanApplication(stmt, false), nil
}

func (db *DB) ListEvents(applicationID int64) ([]ApplicationEvent, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	stmt, err := db.prepareLocked(`SELECT id, application_id, event_type, event_date, details
FROM application_events
WHERE application_id = ?
ORDER BY event_date ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer C.sqlite3_finalize(stmt)
	if err := bindArgs(stmt, applicationID); err != nil {
		return nil, err
	}

	var events []ApplicationEvent
	for {
		rc := C.sqlite3_step(stmt)
		if rc == C.SQLITE_DONE {
			return events, nil
		}
		if rc != C.SQLITE_ROW {
			return nil, db.errLocked()
		}
		events = append(events, scanEvent(stmt))
	}
}

func (db *DB) ListInterviewEvents() ([]ApplicationEvent, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	stmt, err := db.prepareLocked(`SELECT id, application_id, event_type, event_date, details
FROM application_events
WHERE event_type = 'INTERVIEW'
ORDER BY event_date ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer C.sqlite3_finalize(stmt)

	var events []ApplicationEvent
	for {
		rc := C.sqlite3_step(stmt)
		if rc == C.SQLITE_DONE {
			return events, nil
		}
		if rc != C.SQLITE_ROW {
			return nil, db.errLocked()
		}
		events = append(events, scanEvent(stmt))
	}
}

func (db *DB) CreateApplication(app Application) (int64, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, err := db.execLocked("BEGIN"); err != nil {
		return 0, err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = db.execLocked("ROLLBACK")
		}
	}()

	id, err := db.execLocked(`INSERT INTO applications (
	company, role, location, pay_min, pay_max, link, notes,
	app_open_date, app_close_date, intern_start_date, intern_end_date
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		app.Company, app.Role, app.Location, app.PayMin, app.PayMax, app.Link, app.Notes,
		app.AppOpenDate, app.AppCloseDate, app.InternStartDate, app.InternEndDate)
	if err != nil {
		return 0, err
	}

	_, err = db.execLocked(`INSERT INTO application_events (application_id, event_type, event_date, details)
VALUES (?, ?, ?, ?)`, id, "APPLIED", time.Now().UTC().Format("2006-01-02"), "Initial application submitted")
	if err != nil {
		return 0, err
	}
	if _, err := db.execLocked("COMMIT"); err != nil {
		return 0, err
	}
	committed = true
	return id, nil
}

func (db *DB) UpdateApplication(app Application) error {
	_, err := db.Exec(`UPDATE applications SET
	company = ?, role = ?, location = ?, link = ?, pay_min = ?, pay_max = ?, notes = ?,
	app_open_date = ?, app_close_date = ?, intern_start_date = ?, intern_end_date = ?
WHERE id = ?`,
		app.Company, app.Role, app.Location, app.Link, app.PayMin, app.PayMax, app.Notes,
		app.AppOpenDate, app.AppCloseDate, app.InternStartDate, app.InternEndDate, app.ID)
	return err
}

func (db *DB) CreateEvent(event ApplicationEvent) error {
	_, err := db.Exec(`INSERT INTO application_events (application_id, event_type, event_date, details)
VALUES (?, ?, ?, ?)`, event.ApplicationID, event.EventType, event.EventDate, event.Details)
	return err
}

func (db *DB) DeleteApplication(id int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, err := db.execLocked("BEGIN"); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = db.execLocked("ROLLBACK")
		}
	}()

	if _, err := db.execLocked(`DELETE FROM application_events WHERE application_id = ?`, id); err != nil {
		return err
	}
	if _, err := db.execLocked(`DELETE FROM applications WHERE id = ?`, id); err != nil {
		return err
	}
	if _, err := db.execLocked("COMMIT"); err != nil {
		return err
	}
	committed = true
	return nil
}

func (db *DB) DeleteEvent(eventID, applicationID int64) error {
	_, err := db.Exec(`DELETE FROM application_events WHERE id = ? AND application_id = ?`, eventID, applicationID)
	return err
}

func scanApplication(stmt *C.sqlite3_stmt, withLatest bool) Application {
	app := Application{
		ID:              colInt64(stmt, 0),
		Company:         colText(stmt, 1),
		Role:            colText(stmt, 2),
		Location:        colText(stmt, 3),
		PayMin:          colFloatPtr(stmt, 4),
		PayMax:          colFloatPtr(stmt, 5),
		Link:            colText(stmt, 6),
		Notes:           colText(stmt, 7),
		AppOpenDate:     colText(stmt, 8),
		AppCloseDate:    colText(stmt, 9),
		InternStartDate: colText(stmt, 10),
		InternEndDate:   colText(stmt, 11),
	}
	if withLatest {
		app.LatestStatus = colText(stmt, 12)
	}
	return app
}

func scanEvent(stmt *C.sqlite3_stmt) ApplicationEvent {
	return ApplicationEvent{
		ID:            colInt64(stmt, 0),
		ApplicationID: colInt64(stmt, 1),
		EventType:     colText(stmt, 2),
		EventDate:     colText(stmt, 3),
		Details:       colText(stmt, 4),
	}
}

func colText(stmt *C.sqlite3_stmt, col int) string {
	if C.sqlite3_column_type(stmt, C.int(col)) == C.SQLITE_NULL {
		return ""
	}
	return C.GoString((*C.char)(unsafe.Pointer(C.sqlite3_column_text(stmt, C.int(col)))))
}

func colInt64(stmt *C.sqlite3_stmt, col int) int64 {
	return int64(C.sqlite3_column_int64(stmt, C.int(col)))
}

func colFloatPtr(stmt *C.sqlite3_stmt, col int) *float64 {
	if C.sqlite3_column_type(stmt, C.int(col)) == C.SQLITE_NULL {
		return nil
	}
	v := float64(C.sqlite3_column_double(stmt, C.int(col)))
	return &v
}
