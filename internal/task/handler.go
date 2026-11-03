package task

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/raffy-io/taskifyv3/internal/db"
	"github.com/raffy-io/taskifyv3/internal/middleware"
	"github.com/raffy-io/taskifyv3/internal/web"
)

type taskHandler struct {
	Queries  *db.Queries
	Session  *scs.SessionManager
	Renderer *web.Renderer
}

func (h *taskHandler) ServeUserTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	form := newTaskForm("", 0)
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		slog.Error("Unauthorized attempt to fetch tasks")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 1. Query the exact user by ID
	user, err := h.Queries.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("Failed to fetch user by ID", "error", err, "user_id", userID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 2. Query the user's tasks
	tasks, err := h.Queries.ListTasks(ctx, userID)
	if err != nil {
		slog.Error("Failed to list tasks", "error", err, "user_id", userID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 3. Pass the username string to your component
	component := tasksListPage(time.Now(), tasks, user.Username, form)
	if err := h.Renderer.Render(w, r, component); err != nil {
		slog.Error("Failed to render template", "error", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

func (h *taskHandler) createTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		slog.Error("Unauthorized")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	form := newTaskForm(r.PostFormValue("title"), userID)
	if !form.validate() {
		// 1. Return 422 UnprocessableEntity for validation failures
		w.WriteHeader(http.StatusUnprocessableEntity)
		// 2. Render the Templ component with the form errors attached
		if err := createTaskForm(form, false).Render(ctx, w); err != nil {
			slog.Error("Failed to render template", "err", err)
		}
		// 3. Stop execution so it doesn't proceed to user creation!
		return
	}

	task, err := h.Queries.CreateTask(ctx, db.CreateTaskParams{
		Title:  form.Title,
		UserID: userID,
	})
	if err != nil {
		slog.Error("Failed", "error", err)
		http.Error(w, "Failed to add task", http.StatusInternalServerError)
		return
	}

	slog.Info("Successfully added task", "title", task.Title)

	// 1. Query the exact user by ID
	user, err := h.Queries.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("Failed to fetch user by ID", "error", err, "user_id", userID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 2. Query the user's tasks
	tasks, err := h.Queries.ListTasks(ctx, userID)
	if err != nil {
		slog.Error("Failed to list tasks", "error", err, "user_id", userID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render the OOB templates
	w.WriteHeader(http.StatusOK)
	if err := tasksHeading(time.Now(), user.Username, true).Render(ctx, w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
	if err := createTaskForm(form, true).Render(ctx, w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
	if err := tasksCounter(tasks, true).Render(ctx, w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
	if err := tasksList(tasks, true).Render(ctx, w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}

}

func (h *taskHandler) toggleTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	idStr := chi.URLParam(r, "id")
	// 2. Convert it from string to int64 safely using strconv
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.Error("Invalid task ID format", "id", idStr, "error", err)
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		slog.Error("Unauthorized attempt to create task")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	task, err := h.Queries.ToggleTaskStatus(ctx, db.ToggleTaskStatusParams{
		ID:     taskID,
		UserID: userID,
	})

	slog.Info("Successfully updated task", "task", task.Title)

	// 2. Query the user's tasks
	tasks, err := h.Queries.ListTasks(ctx, userID)
	if err != nil {
		slog.Error("Failed to list tasks", "error", err, "user_id", userID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render the OOB templates
	w.WriteHeader(http.StatusOK)
	if err := tasksList(tasks, true).Render(ctx, w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}

}
func (h *taskHandler) deleteTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Get and verify the logged-in user ID from context
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		slog.Error("Unauthorized attempt to delete task")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Grab the task ID from the Chi URL parameter
	idStr := chi.URLParam(r, "id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.Error("Invalid task ID format for deletion", "id", idStr, "error", err)
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// 3. Execute the SQLC query with both taskID and userID
	err = h.Queries.DeleteTask(ctx, db.DeleteTaskParams{
		ID:     taskID,
		UserID: userID,
	})
	if err != nil {
		slog.Error("Failed to delete task", "error", err)
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	// 1. Query the exact user by ID
	user, err := h.Queries.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("Failed to fetch user by ID", "error", err, "user_id", userID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 2. Query the user's tasks
	tasks, err := h.Queries.ListTasks(ctx, userID)
	if err != nil {
		slog.Error("Failed to list tasks", "error", err, "user_id", userID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Render the OOB templates
	w.WriteHeader(http.StatusOK)
	if err := tasksHeading(time.Now(), user.Username, true).Render(ctx, w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
	if err := tasksCounter(tasks, true).Render(ctx, w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
	if err := tasksList(tasks, true).Render(ctx, w); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
