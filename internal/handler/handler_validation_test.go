package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/team-task-api/internal/middleware"
)

func postJSON(url string, body any) *http.Request {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func putJSON(url string, body any) *http.Request {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	var m map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&m); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return m
}

func catchPanic(fn func()) (panicked bool) {
	panicked = false
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	fn()
	return
}

func withChiURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func withUserID(r *http.Request, userID string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, userID))
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"user@example.com", true},
		{"a@b.c", true},
		{"test.email@domain.co.uk", true},
		{"user+tag@example.com", true},
		{"", false},
		{"foo", false},
		{"foo@", false},
		{"@bar.com", false},
		{"a@b", false},
		{"a@.com", false},
		{"a@b.", false},
		{"a@b..com", true},
		{"@@", false},
		{"not-an-email", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := isValidEmail(tc.input)
			if got != tc.valid {
				t.Errorf("isValidEmail(%q) = %v, want %v", tc.input, got, tc.valid)
			}
		})
	}
}

func TestIsValidTaskStatus(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"pending", true},
		{"in_progress", true},
		{"done", true},
		{"", false},
		{"PENDING", false},
		{"invalid", false},
		{"in progress", false},
		{"cancel", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := isValidTaskStatus(tc.input)
			if got != tc.valid {
				t.Errorf("isValidTaskStatus(%q) = %v, want %v", tc.input, got, tc.valid)
			}
		})
	}
}

func TestRegisterValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		wantStatus int
		wantErr    string
	}{
		{
			name:       "empty body",
			body:       struct{}{},
			wantStatus: 400,
			wantErr:    "email, password, and name are required",
		},
		{
			name:       "missing password",
			body:       map[string]string{"email": "a@b.com", "name": "A"},
			wantStatus: 400,
			wantErr:    "email, password, and name are required",
		},
		{
			name:       "missing name",
			body:       map[string]string{"email": "a@b.com", "password": "secret"},
			wantStatus: 400,
			wantErr:    "email, password, and name are required",
		},
		{
			name:       "missing email",
			body:       map[string]string{"password": "secret", "name": "A"},
			wantStatus: 400,
			wantErr:    "email, password, and name are required",
		},
		{
			name:       "blank email",
			body:       map[string]string{"email": "  ", "password": "secret", "name": "A"},
			wantStatus: 400,
			wantErr:    "email, password, and name are required",
		},
		{
			name:       "invalid email no at",
			body:       map[string]string{"email": "foo", "password": "secret", "name": "A"},
			wantStatus: 400,
			wantErr:    "invalid email format",
		},
		{
			name:       "invalid email no dot",
			body:       map[string]string{"email": "foo@bar", "password": "secret", "name": "A"},
			wantStatus: 400,
			wantErr:    "invalid email format",
		},
		{
			name:       "invalid email no domain",
			body:       map[string]string{"email": "foo@", "password": "secret", "name": "A"},
			wantStatus: 400,
			wantErr:    "invalid email format",
		},
		{
			name:       "invalid json",
			body:       "not json",
			wantStatus: 400,
			wantErr:    "invalid request body",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewAuthHandler(nil)
			req := postJSON("/api/v1/register", tc.body)
			rec := httptest.NewRecorder()

			panicked := catchPanic(func() { h.Register(rec, req) })

			if panicked {
				t.Fatal("handler panicked: validation should reject before service call")
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantErr != "" {
				m := decodeJSON(t, rec)
				if m["error"] != tc.wantErr {
					t.Errorf("error = %q, want %q", m["error"], tc.wantErr)
				}
			}
		})
	}
}

func TestLoginValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		wantStatus int
		wantErr    string
	}{
		{
			name:       "empty body",
			body:       struct{}{},
			wantStatus: 400,
			wantErr:    "email and password are required",
		},
		{
			name:       "missing password",
			body:       map[string]string{"email": "a@b.com"},
			wantStatus: 400,
			wantErr:    "email and password are required",
		},
		{
			name:       "missing email",
			body:       map[string]string{"password": "secret"},
			wantStatus: 400,
			wantErr:    "email and password are required",
		},
		{
			name:       "blank email",
			body:       map[string]string{"email": "  ", "password": "secret"},
			wantStatus: 400,
			wantErr:    "email and password are required",
		},
		{
			name:       "invalid email",
			body:       map[string]string{"email": "foo", "password": "secret"},
			wantStatus: 400,
			wantErr:    "invalid email format",
		},
		{
			name:       "invalid json",
			body:       "not json",
			wantStatus: 400,
			wantErr:    "invalid request body",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewAuthHandler(nil)
			req := postJSON("/api/v1/login", tc.body)
			rec := httptest.NewRecorder()

			panicked := catchPanic(func() { h.Login(rec, req) })

			if panicked {
				t.Fatal("handler panicked: validation should reject before service call")
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantErr != "" {
				m := decodeJSON(t, rec)
				if m["error"] != tc.wantErr {
					t.Errorf("error = %q, want %q", m["error"], tc.wantErr)
				}
			}
		})
	}
}

func TestTeamCreateValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		wantStatus int
		wantErr    string
	}{
		{
			name:       "empty body",
			body:       struct{}{},
			wantStatus: 400,
			wantErr:    "team name is required",
		},
		{
			name:       "blank name",
			body:       map[string]string{"name": "   "},
			wantStatus: 400,
			wantErr:    "team name is required",
		},
		{
			name:       "invalid json",
			body:       "not json",
			wantStatus: 400,
			wantErr:    "invalid request body",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewTeamHandler(nil)
			req := postJSON("/api/v1/teams", tc.body)
			rec := httptest.NewRecorder()

			panicked := catchPanic(func() { h.Create(rec, req) })

			if panicked {
				t.Fatal("handler panicked: validation should reject before service call")
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantErr != "" {
				m := decodeJSON(t, rec)
				if m["error"] != tc.wantErr {
					t.Errorf("error = %q, want %q", m["error"], tc.wantErr)
				}
			}
		})
	}
}

func TestTeamInviteValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		wantStatus int
		wantErr    string
	}{
		{
			name:       "empty body",
			body:       struct{}{},
			wantStatus: 400,
			wantErr:    "user_id is required",
		},
		{
			name:       "blank user_id",
			body:       map[string]string{"user_id": "  "},
			wantStatus: 400,
			wantErr:    "user_id is required",
		},
		{
			name:       "invalid role",
			body:       map[string]string{"user_id": "u1", "role": "superadmin"},
			wantStatus: 400,
			wantErr:    "invalid role: must be admin or member",
		},
		{
			name:       "invalid role owner",
			body:       map[string]string{"user_id": "u1", "role": "owner"},
			wantStatus: 400,
			wantErr:    "invalid role: must be admin or member",
		},
		{
			name:       "invalid json",
			body:       "not json",
			wantStatus: 400,
			wantErr:    "invalid request body",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewTeamHandler(nil)
			req := withChiURLParam(postJSON("/api/v1/teams/abc123/invite", tc.body), "id", "abc123")
			rec := httptest.NewRecorder()

			panicked := catchPanic(func() { h.Invite(rec, req) })

			if panicked {
				t.Fatal("handler panicked: validation should reject before service call")
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantErr != "" {
				m := decodeJSON(t, rec)
				if m["error"] != tc.wantErr {
					t.Errorf("error = %q, want %q", m["error"], tc.wantErr)
				}
			}
		})
	}
}

func TestTaskCreateValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		wantStatus int
		wantErr    string
	}{
		{
			name:       "empty body",
			body:       struct{}{},
			wantStatus: 400,
			wantErr:    "team_id and title are required",
		},
		{
			name:       "missing title",
			body:       map[string]string{"team_id": "t1"},
			wantStatus: 400,
			wantErr:    "team_id and title are required",
		},
		{
			name:       "missing team_id",
			body:       map[string]string{"title": "Task"},
			wantStatus: 400,
			wantErr:    "team_id and title are required",
		},
		{
			name:       "blank team_id",
			body:       map[string]string{"team_id": "  ", "title": "Task"},
			wantStatus: 400,
			wantErr:    "team_id and title are required",
		},
		{
			name:       "blank title",
			body:       map[string]string{"team_id": "t1", "title": "  "},
			wantStatus: 400,
			wantErr:    "team_id and title are required",
		},
		{
			name:       "invalid json",
			body:       "not json",
			wantStatus: 400,
			wantErr:    "invalid request body",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewTaskHandler(nil, nil, nil)
			req := postJSON("/api/v1/tasks", tc.body)
			rec := httptest.NewRecorder()

			panicked := catchPanic(func() { h.Create(rec, req) })

			if panicked {
				t.Fatal("handler panicked: validation should reject before service call")
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantErr != "" {
				m := decodeJSON(t, rec)
				if m["error"] != tc.wantErr {
					t.Errorf("error = %q, want %q", m["error"], tc.wantErr)
				}
			}
		})
	}
}

func TestTaskUpdateValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		wantStatus int
		wantErr    string
	}{
		{
			name:       "empty body",
			body:       struct{}{},
			wantStatus: 400,
			wantErr:    "at least one field must be provided",
		},
		{
			name:       "invalid status",
			body:       map[string]string{"status": "invalid"},
			wantStatus: 400,
			wantErr:    "invalid status: must be pending, in_progress, or done",
		},
		{
			name:       "invalid status uppercase",
			body:       map[string]string{"status": "PENDING"},
			wantStatus: 400,
			wantErr:    "invalid status: must be pending, in_progress, or done",
		},
		{
			name:       "invalid json",
			body:       "not json",
			wantStatus: 400,
			wantErr:    "invalid request body",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewTaskHandler(nil, nil, nil)
			req := withChiURLParam(putJSON("/api/v1/tasks/abc123", tc.body), "id", "abc123")
			rec := httptest.NewRecorder()

			panicked := catchPanic(func() { h.Update(rec, req) })

			if panicked {
				t.Fatal("handler panicked: validation should reject before service call")
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantErr != "" {
				m := decodeJSON(t, rec)
				if m["error"] != tc.wantErr {
					t.Errorf("error = %q, want %q", m["error"], tc.wantErr)
				}
			}
		})
	}
}

func TestTaskListValidation(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantErr    string
	}{
		{
			name:       "invalid status filter",
			query:      "?status=invalid",
			wantStatus: 400,
			wantErr:    "invalid status: must be pending, in_progress, or done",
		},
		{
			name:       "uppercase status rejected",
			query:      "?status=DONE",
			wantStatus: 400,
			wantErr:    "invalid status: must be pending, in_progress, or done",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewTaskHandler(nil, nil, nil)
			req := withUserID(
				httptest.NewRequest(http.MethodGet, "/api/v1/tasks"+tc.query, nil),
				"user-1",
			)
			rec := httptest.NewRecorder()

			panicked := catchPanic(func() { h.List(rec, req) })

			if panicked {
				t.Fatal("handler panicked: validation should reject before service call")
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantErr != "" {
				m := decodeJSON(t, rec)
				if m["error"] != tc.wantErr {
					t.Errorf("error = %q, want %q", m["error"], tc.wantErr)
				}
			}
		})
	}
}

func TestCommentCreateValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		wantStatus int
		wantErr    string
	}{
		{
			name:       "empty body",
			body:       struct{}{},
			wantStatus: 400,
			wantErr:    "comment content is required",
		},
		{
			name:       "blank content",
			body:       map[string]string{"content": "   "},
			wantStatus: 400,
			wantErr:    "comment content is required",
		},
		{
			name:       "invalid json",
			body:       "not json",
			wantStatus: 400,
			wantErr:    "invalid request body",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewTaskCommentHandler(nil)
			req := withChiURLParam(
				postJSON("/api/v1/tasks/abc123/comments", tc.body),
				"id", "abc123",
			)
			rec := httptest.NewRecorder()

			panicked := catchPanic(func() { h.Create(rec, req) })

			if panicked {
				t.Fatal("handler panicked: validation should reject before service call")
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantErr != "" {
				m := decodeJSON(t, rec)
				if m["error"] != tc.wantErr {
					t.Errorf("error = %q, want %q", m["error"], tc.wantErr)
				}
			}
		})
	}
}

func TestCommentListValidation(t *testing.T) {
	h := NewTaskCommentHandler(nil)
	req := withUserID(
		withChiURLParam(
			httptest.NewRequest(http.MethodGet, "/api/v1/tasks/", nil),
			"id", "",
		),
		"user-1",
	)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	m := decodeJSON(t, rec)
	if m["error"] != "task id is required" {
		t.Errorf("error = %q, want %q", m["error"], "task id is required")
	}
}

func TestCommentDeleteValidation(t *testing.T) {
	tests := []struct {
		name       string
		taskID     string
		commentID  string
		wantStatus int
		wantErr    string
	}{
		{
			name:       "missing task id",
			taskID:     "",
			commentID:  "c1",
			wantStatus: 400,
			wantErr:    "task id is required",
		},
		{
			name:       "missing comment id",
			taskID:     "t1",
			commentID:  "",
			wantStatus: 400,
			wantErr:    "comment id is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewTaskCommentHandler(nil)
			req := httptest.NewRequest(http.MethodDelete, "/", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.taskID)
			rctx.URLParams.Add("commentID", tc.commentID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rec := httptest.NewRecorder()

			h.Delete(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			m := decodeJSON(t, rec)
			if m["error"] != tc.wantErr {
				t.Errorf("error = %q, want %q", m["error"], tc.wantErr)
			}
		})
	}
}

func TestTaskDeleteValidation(t *testing.T) {
	h := NewTaskHandler(nil, nil, nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/", nil)
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	m := decodeJSON(t, rec)
	if m["error"] != "task id is required" {
		t.Errorf("error = %q, want %q", m["error"], "task id is required")
	}
}

func TestTaskGetByIDValidation(t *testing.T) {
	h := NewTaskHandler(nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/", nil)
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	m := decodeJSON(t, rec)
	if m["error"] != "task id is required" {
		t.Errorf("error = %q, want %q", m["error"], "task id is required")
	}
}

func TestTeamGetByIDValidation(t *testing.T) {
	h := NewTeamHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/", nil)
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	m := decodeJSON(t, rec)
	if m["error"] != "team id is required" {
		t.Errorf("error = %q, want %q", m["error"], "team id is required")
	}
}

func TestTeamUpdateValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		wantStatus int
		wantErr    string
	}{
		{
			name:       "empty body",
			body:       struct{}{},
			wantStatus: 400,
			wantErr:    "at least one field must be provided",
		},
		{
			name:       "invalid json",
			body:       "not json",
			wantStatus: 400,
			wantErr:    "invalid request body",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewTeamHandler(nil)
			req := withChiURLParam(putJSON("/api/v1/teams/abc123", tc.body), "id", "abc123")
			rec := httptest.NewRecorder()

			h.Update(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantErr != "" {
				m := decodeJSON(t, rec)
				if m["error"] != tc.wantErr {
					t.Errorf("error = %q, want %q", m["error"], tc.wantErr)
				}
			}
		})
	}
}

func TestTeamUpdateValidation_EmptyName(t *testing.T) {
	h := NewTeamHandler(nil)
	req := withChiURLParam(
		putJSON("/api/v1/teams/abc123", map[string]string{"name": "  "}),
		"id", "abc123",
	)
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != 400 {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	m := decodeJSON(t, rec)
	if m["error"] != "team name is required" {
		t.Errorf("error = %q, want %q", m["error"], "team name is required")
	}
}

func TestTeamDeleteValidation(t *testing.T) {
	h := NewTeamHandler(nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/", nil)
	rec := httptest.NewRecorder()

	h.Delete(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	m := decodeJSON(t, rec)
	if m["error"] != "team id is required" {
		t.Errorf("error = %q, want %q", m["error"], "team id is required")
	}
}

func TestHistoryListValidation(t *testing.T) {
	h := NewTaskHistoryHandler(nil)
	req := withUserID(
		httptest.NewRequest(http.MethodGet, "/api/v1/tasks/", nil),
		"user-1",
	)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	m := decodeJSON(t, rec)
	if m["error"] != "task id is required" {
		t.Errorf("error = %q, want %q", m["error"], "task id is required")
	}
}

func TestAuthHandlers_AllCases(t *testing.T) {
	t.Run("Register", TestRegisterValidation)
	t.Run("Login", TestLoginValidation)
}
