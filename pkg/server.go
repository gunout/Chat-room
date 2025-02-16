package barertc

import (
	"io"
	"net/http"
	"sync"

	"git.kirsle.net/apps/barertc/pkg/config"
	"git.kirsle.net/apps/barertc/pkg/log"
	"git.kirsle.net/apps/barertc/pkg/models"
)

// Server is the primary back-end server struct for BareRTC, see main.go
type Server struct {
	// HTTP router.
	mux *http.ServeMux

	// Max number of messages we'll buffer for a subscriber
	subscriberMessageBuffer int

	// Connected WebSocket subscribers.
	subscribersMu sync.RWMutex
	subscribers   map[*Subscriber]struct{}

	// Cached filehandles for channel logging.
	logfh map[string]io.WriteCloser
}

// NewServer initializes the Server.
func NewServer() *Server {
	return &Server{
		subscriberMessageBuffer: 32,
		subscribers:             make(map[*Subscriber]struct{}),
	}
}

// Setup the server: configure HTTP routes, etc.
func (s *Server) Setup() error {
	// Enable the SQLite database for DM history?
	if config.Current.DirectMessageHistory.Enabled {
		if err := models.Initialize(config.Current.DirectMessageHistory.SQLiteDatabase); err != nil {
			log.Error("Error initializing SQLite database: %s", err)
		}
	}

	var mux = http.NewServeMux()

	mux.Handle("/", IndexPage())
	mux.Handle("/about", AboutPage())
	mux.Handle("/logout", LogoutPage())
	mux.Handle("/ws", s.WebSocket())
	mux.Handle("/poll", s.PollingAPI())
	mux.Handle("/api/statistics", s.Statistics())
	mux.Handle("/api/blocklist", s.BlockList())
	mux.Handle("/api/block/now", s.BlockNow())
	mux.Handle("/api/disconnect/now", s.DisconnectNow())
	mux.Handle("/api/authenticate", s.Authenticate())
	mux.Handle("/api/shutdown", s.ShutdownAPI())
	mux.Handle("/api/profile", s.UserProfile())
	mux.Handle("/api/message/history", s.MessageHistory())
	mux.Handle("/api/message/clear", s.ClearMessages())
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("dist/assets"))))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("dist/static"))))

	s.mux = mux

	return nil
}

// ListenAndServe starts the web server.
func (s *Server) ListenAndServe(address string) error {
	// Run the polling user idle kicker.
	go s.KickIdlePollUsers()
	return http.ListenAndServe(address, s.mux)
}
