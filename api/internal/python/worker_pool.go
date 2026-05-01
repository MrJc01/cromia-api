package python

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// ─────────────────────────── Tipos de Dados ────────────────────────────────

type Request struct {
	Model    string                 `json:"model"`
	Messages []map[string]string    `json:"messages"`
	Params   map[string]interface{} `json:"params,omitempty"`
	Stream   bool                   `json:"stream,omitempty"`
}

type Response struct {
	Output string         `json:"output"`
	Usage  map[string]int `json:"usage"`
	Error  string         `json:"error,omitempty"`
}

// Job representa um trabalho de inferência sem streaming.
type Job struct {
	Input  Request
	Result chan JobResult
}

type JobResult struct {
	Response Response
	Err      error
}

// StreamJob representa um trabalho onde os chunks são escritos em um io.Writer.
type StreamJob struct {
	Input  Request
	Writer io.Writer
	Done   chan error
}

// internalJob é o tipo interno que unifica ambos os modos.
type internalJob struct {
	req       Request
	result    chan JobResult // != nil para modo normal
	streamOut io.Writer     // != nil para modo streaming
	streamDone chan error
}

// ─────────────────────────── Worker ────────────────────────────────────────

type Worker struct {
	id         int
	jobs       chan internalJob
	socketPath string
	mu         sync.Mutex
	cmd        *exec.Cmd
	ready      bool
}

// WorkerPool gerencia N workers independentes com supervisão e auto-restart.
type WorkerPool struct {
	workers  []*Worker
	jobQueue chan internalJob
}

// NewWorkerPool cria e inicia um pool de workers supervisionados.
func NewWorkerPool(numWorkers int, queueSize int) *WorkerPool {
	pool := &WorkerPool{
		jobQueue: make(chan internalJob, queueSize),
	}
	for i := 0; i < numWorkers; i++ {
		socketPath := fmt.Sprintf("/tmp/cromia_worker_%d.sock", i)
		w := &Worker{
			id:         i,
			jobs:       pool.jobQueue,
			socketPath: socketPath,
		}
		pool.workers = append(pool.workers, w)
		go w.supervise()
	}
	return pool
}

// supervise mantém o processo Python vivo, reiniciando com backoff em caso de falha.
func (w *Worker) supervise() {
	backoff := 500 * time.Millisecond
	const maxBackoff = 30 * time.Second

	for {
		log.Printf("[Worker %d] Iniciando processo Python...", w.id)
		err := w.run()
		if err != nil {
			log.Printf("[Worker %d] Processo encerrou com erro: %v. Reiniciando em %v", w.id, err, backoff)
		} else {
			log.Printf("[Worker %d] Processo encerrou. Reiniciando em %v", w.id, backoff)
		}

		w.mu.Lock()
		w.ready = false
		w.mu.Unlock()

		time.Sleep(backoff)
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

// run inicia o processo Python, espera o sinal "READY" e processa jobs.
// Retorna quando o processo Python morre.
func (w *Worker) run() error {
	// Remove socket antigo se existir
	os.Remove(w.socketPath)

	cmd := exec.Command("python3", "worker/crom/socket_server.py", w.socketPath)
	cmd.Env = append(os.Environ(), "PYTHONPATH="+filepath.Join(os.Getenv("PWD"), "worker"))
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cmd.Start: %w", err)
	}

	w.mu.Lock()
	w.cmd = cmd
	w.mu.Unlock()

	// Espera o sinal "READY\n" do Python
	scanner := bufio.NewScanner(stdout)
	readyCh := make(chan bool, 1)
	go func() {
		if scanner.Scan() && scanner.Text() == "READY" {
			readyCh <- true
		} else {
			readyCh <- false
		}
	}()

	select {
	case ok := <-readyCh:
		if !ok {
			cmd.Process.Kill()
			cmd.Wait()
			return fmt.Errorf("Python worker não enviou READY")
		}
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		cmd.Wait()
		return fmt.Errorf("timeout esperando READY do worker %d", w.id)
	}

	w.mu.Lock()
	w.ready = true
	w.mu.Unlock()
	log.Printf("[Worker %d] Pronto e aguardando jobs.", w.id)

	// Processa jobs até o processo morrer
	doneCh := make(chan error, 1)
	go func() { doneCh <- cmd.Wait() }()

	for {
		select {
		case exitErr := <-doneCh:
			return exitErr
		case job := <-w.jobs:
			w.processJob(job)
		}
	}
}

// processJob executa um job com timeout de 60 segundos.
func (w *Worker) processJob(job internalJob) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	errResult := func(err error) {
		if job.result != nil {
			job.result <- JobResult{Err: err}
		}
		if job.streamDone != nil {
			job.streamDone <- err
		}
	}

	conn, err := net.DialTimeout("unix", w.socketPath, 5*time.Second)
	if err != nil {
		errResult(fmt.Errorf("[Worker %d] socket connect error: %w", w.id, err))
		return
	}

	// Define deadline conforme timeout do context
	deadline, _ := ctx.Deadline()
	conn.SetDeadline(deadline)
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(job.req); err != nil {
		errResult(fmt.Errorf("[Worker %d] encode error: %w", w.id, err))
		return
	}

	// ─── Modo Streaming ───────────────────────────────────────────────────────
	if job.req.Stream && job.streamOut != nil {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}
			// Escreve cada chunk no writer (SSE ou outro)
			fmt.Fprintf(job.streamOut, "data: %s\n\n", line)
		}
		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			job.streamDone <- fmt.Errorf("stream read error: %w", err)
			return
		}
		job.streamDone <- nil
		return
	}

	// ─── Modo Normal ──────────────────────────────────────────────────────────
	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		errResult(fmt.Errorf("[Worker %d] decode error: %w", w.id, err))
		return
	}
	if resp.Error != "" {
		errResult(fmt.Errorf("python error: %s", resp.Error))
		return
	}
	job.result <- JobResult{Response: resp}
}

// ─────────────────────────── Pool API ──────────────────────────────────────

// Submit envia um job de inferência normal (sem streaming).
func (p *WorkerPool) Submit(job Job) {
	p.jobQueue <- internalJob{
		req:    job.Input,
		result: job.Result,
	}
}

// SubmitStream envia um job de streaming com escrita direta em um io.Writer.
func (p *WorkerPool) SubmitStream(job StreamJob) {
	p.jobQueue <- internalJob{
		req:        job.Input,
		streamOut:  job.Writer,
		streamDone: job.Done,
	}
}
