package server

import (
	"encoding/json"
	"github.com/Av1shay/timers-scheduler-demo/logx"
	"github.com/Av1shay/timers-scheduler-demo/task"
	"github.com/gorilla/mux"
	"gopkg.in/dealancer/validate.v2"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	taskService *task.Service
}

func New(taskService *task.Service) *Server {
	s := &Server{taskService}
	return s
}

func (s *Server) MountHandlers(router *mux.Router) {
	router.Use(traceIdMiddleware)
	router.Use(logMiddleware)
	router.HandleFunc("/timers", s.NewTimer).Methods(http.MethodPost)
	router.HandleFunc("/timers/{id}", s.GetTimer).Methods(http.MethodGet)
	router.HandleFunc("/test-webhook/{id}", s.Test).Methods(http.MethodPost) // for testing purposes
}

func (s *Server) NewTimer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	defer r.Body.Close()
	var reqBody SetTimerReq
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logx.Error(ctx, "failed to parse request body:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validate.Validate(reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	n := time.Now().UTC()
	dueDate := n.Add(time.Hour*time.Duration(reqBody.Hours) + time.Minute*time.Duration(reqBody.Minutes) + time.Second*time.Duration(reqBody.Seconds))
	createdTask, err := s.taskService.SaveTask(ctx, dueDate, reqBody.URL)
	if err != nil {
		logx.Error(ctx, "failed to save task:", err)
		msg, code := parseError(err)
		http.Error(w, msg, code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SetTimerResp{ID: createdTask.ID})
}

func (s *Server) GetTimer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := mux.Vars(r)
	idParam := params["id"]

	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "id must be be a numeric number", http.StatusBadRequest)
		return
	}

	t, err := s.taskService.GetTask(ctx, id)
	if err != nil {
		logx.Error(ctx, "failed to get task:", err)
		msg, code := parseError(err)
		http.Error(w, msg, code)
		return
	}

	n := time.Now().UTC().Truncate(time.Second)
	timeLeft := t.DueDate.Sub(n)
	secs := int64(timeLeft.Seconds())
	if secs < 0 {
		secs = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetTimerResp{ID: t.ID, TimeLeft: secs})
}

// Test dummy route just to check webhooks
func (s *Server) Test(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	logx.Info(r.Context(), "webhook called for task", id)
	w.WriteHeader(http.StatusOK)
}

func parseError(err error) (string, int) {
	msg := err.Error()
	code := http.StatusInternalServerError
	if e, ok := err.(*task.ApiError); ok {
		msg = e.ClientMessage
		code = e.Code
	}
	return msg, code
}

func traceIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := logx.ContextWithTraceID(r.Context())
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logx.Info(r.Context(), r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
