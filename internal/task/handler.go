package task

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/raffy-io/taskifyv3/internal/db"
	"github.com/raffy-io/taskifyv3/internal/middleware"
)

type taskHandler struct {
	Queries *db.Queries
	Session *scs.SessionManager
}

func (h *taskHandler) createTask(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w,"Bad request",http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		slog.Error("Unauthorized attempt to create task")
		http.Error(w, "Unauthorized", http.StatusUnauthorized) 
		return
	}

	title := strings.TrimSpace(r.PostFormValue("title"))
	form := &taskForm{
		Title: title,
		UserID: &userID,
	}
	if !form.Validate(){
		slog.Error("Failed to add task", "err", form.Errors)
		http.Error(w, "Validation failed", http.StatusBadRequest)
		return
	}

	task,err := h.Queries.CreateTask(ctx,db.CreateTaskParams{
		Title: form.Title,
		UserID: userID,
	})
	if err != nil {
		slog.Error("Failed","error",err)
		http.Error(w,"Failed to add task",http.StatusInternalServerError)
		return
	}

	slog.Info("Successfully added task","task",task.Title)
}

func (h *taskHandler) toggleTask(w http.ResponseWriter, r *http.Request){
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w,"Bad request",http.StatusBadRequest)
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

	task,err := h.Queries.ToggleTaskStatus(ctx,db.ToggleTaskStatusParams{
		ID: taskID,
		UserID: userID,
	})

	slog.Info("Successfully updated task","task",task.Title)

}
func (h *taskHandler) deleteTask(w http.ResponseWriter, r *http.Request){
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

    slog.Info("Successfully deleted task", "task_id", taskID, "user_id", userID)
}