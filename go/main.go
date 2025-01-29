package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

var (
	templates      *template.Template
	streamedTextCh chan string
)

func init() {
	// Parse all templates in the templates folder.
	templates = template.Must(template.ParseGlob("templates/*.html"))

	streamedTextCh = make(chan string)
}

// generateText calls FastAPI and returns every token received on the fly through
// a dedicated channel (streamedTextCh).
// If the EOF character is received from FastAPI, it means that text generation is over.
func generateText(streamedTextCh chan<- string) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:8000", nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Stream finished")
				break
			}
			log.Println("Error reading stream:", err)
			break
		}

		text := strings.TrimSpace(line)
		log.Println("Received from FastAPI:", text)
		streamedTextCh <- text
	}

	close(streamedTextCh)
	log.Println("Closed streamedTextCh")
}

// formatServerSentEvent creates a proper SSE compatible body.
// Server sent events need to follow a specific formatting that
// uses "event:" and "data:" prefixes.
func formatServerSentEvent(event, data string) (string, error) {
	sb := strings.Builder{}

	_, err := sb.WriteString(fmt.Sprintf("event: %s\n", event))
	if err != nil {
		return "", err
	}
	_, err = sb.WriteString(fmt.Sprintf("data: %v\n\n", data))
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

// generate is an infinite loop that waits for new tokens received
// from the streamedTextCh. Once a new token is received,
// it is automatically pushed to the frontend as a server sent event.
func generate(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	streamedTextCh := make(chan string)

	// Start background text generation
	go generateText(streamedTextCh)

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for text := range streamedTextCh {
		event := fmt.Sprintf("event: streamed-text\ndata: %s\n\n", text)

		log.Println("Sending SSE to client:", event)
		_, err := fmt.Fprint(w, event)
		if err != nil {
			log.Println("Error sending SSE to client:", err)
			break
		}

		flusher.Flush()
	}
	log.Println("Stream ended for client")
}

// start starts an asynchronous request to the AI engine.
func start(w http.ResponseWriter, r *http.Request) {
	go generateText(streamedTextCh)
}

func home(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "home.html", nil); err != nil {
		log.Println(err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/generate", generate)
	http.HandleFunc("/start", start)
	http.HandleFunc("/", home)

	r := mux.NewRouter()
	r.HandleFunc("/generate", generate).Methods("GET")
	r.HandleFunc("/start", start).Methods("POST")
	r.HandleFunc("/", home).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", r))
}
