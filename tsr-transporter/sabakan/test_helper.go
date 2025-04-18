package sabakan

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Sabakan Mock
type sabakanMock struct {
	host   string
	path   string
	resDir string
	mutex  sync.Mutex
}

func (sm *sabakanMock) startMock() {
	server := http.NewServeMux()
	server.HandleFunc(sm.path+"/", sm.reqHandler)
	go func() {
		slog.Error("error at ListenAndServe", "err", http.ListenAndServe(sm.host, server))
	}()
}

func (sm *sabakanMock) getEndpoint() string {
	return "http://" + sm.host + sm.path
}

func (sm *sabakanMock) reqHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve serial number from URL
	items := strings.Split(r.URL.Path, "/")
	fn := items[len(items)-1]
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Create HTTP response from the response file
	fd, err := os.Open(sm.resDir + "/" + fn)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "")
		return
	}
	defer fd.Close()

	// Working time
	time.Sleep(1 * time.Second)

	stringJSON, err := io.ReadAll(fd)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprint(w, string(stringJSON))
}
