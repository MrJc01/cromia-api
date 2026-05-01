package handlers

import (
	"cromia/api/internal/db"
	"cromia/api/internal/middleware"
	"cromia/api/internal/python"
	"encoding/json"
	"log"
	"net/http"
)

type ChatHandler struct {
	Pool *python.WorkerPool
	DB   db.DB
}

// Completions atende POST /v1/chat/completions compatível com OpenAI.
// Detecta o campo "stream" e alterna entre resposta direta e SSE.
func (h *ChatHandler) Completions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Limita o tamanho do payload a 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req python.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
		return
	}

	// Validação básica
	if req.Model == "" {
		http.Error(w, `{"error":"model field required"}`, http.StatusBadRequest)
		return
	}
	if len(req.Messages) == 0 {
		http.Error(w, `{"error":"messages list is empty"}`, http.StatusBadRequest)
		return
	}
	if len(req.Messages) > 100 {
		http.Error(w, `{"error":"too many messages (max 100)"}`, http.StatusBadRequest)
		return
	}

	// ─── Modo Streaming (SSE) ────────────────────────────────────────────────
	if req.Stream {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, `{"error":"streaming not supported"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		doneCh := make(chan error, 1)
		job := python.StreamJob{
			Input:  req,
			Writer: &flushWriter{w: w, f: flusher},
			Done:   doneCh,
		}

		h.Pool.SubmitStream(job)
		if err := <-doneCh; err != nil {
			log.Printf("[ChatHandler] Stream error: %v", err)
			// Envia um evento de erro SSE
			w.Write([]byte("data: {\"error\":\"" + err.Error() + "\"}\n\n"))
			flusher.Flush()
		}
		// Sinaliza fim do stream
		w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
		return
	}

	// ─── Modo Normal (JSON único) ────────────────────────────────────────────
	resultChan := make(chan python.JobResult, 1)
	job := python.Job{
		Input:  req,
		Result: resultChan,
	}

	h.Pool.Submit(job)
	result := <-resultChan

	if result.Err != nil {
		log.Printf("[ChatHandler] Internal Worker Error: %v | Model: %s", result.Err, req.Model)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "internal worker error",
			"details": result.Err.Error(),
		})
		return
	}

	// Tenta registrar a requisição no DB de forma assíncrona
	apiKey, ok := r.Context().Value(middleware.APIKeyContextKey).(*db.APIKey)
	if ok {
		go func() {
			prompt := lastUserMessage(req.Messages)
			h.DB.CreateRequest(apiKey.ID, req.Model, prompt)
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.Response)
}

func lastUserMessage(messages []map[string]string) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i]["role"] == "user" {
			return messages[i]["content"]
		}
	}
	return ""
}

// flushWriter envolve um http.ResponseWriter e chama Flush após cada escrita,
// garantindo que os chunks SSE chegam ao cliente imediatamente.
type flushWriter struct {
	w http.ResponseWriter
	f http.Flusher
}

func (fw *flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	fw.f.Flush()
	return
}
