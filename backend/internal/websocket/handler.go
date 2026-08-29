// Package websocket — handler.go expose le endpoint HTTP GET /ws qui upgrade la connexion
// en WebSocket.
//
// Authentification optionnelle :
//   Le client peut (ou non) envoyer son JWT dans le query param "token".
//   Les clients anonymes peuvent regarder la carte (recevoir ping_created, ping_updated).
//   Seuls les clients authentifiés peuvent envoyer leur position GPS ("position" message).
//
// Upgrade HTTP → WebSocket :
//   gorilla/websocket prend en charge le handshake HTTP Upgrade.
//   On configure l'upgrader avec CheckOrigin permissif en dev (ENV=development).
//   En production, seul le domaine officiel est autorisé.
package websocket

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	gws "github.com/gorilla/websocket"
	"github.com/golang-jwt/jwt/v5"
)

// Handler handles the GET /ws HTTP endpoint.
type Handler struct {
	hub      *Hub
	upgrader gws.Upgrader
	jwtSecret []byte
}

// NewHandler creates a new WebSocket Handler.
func NewHandler(hub *Hub, jwtSecret string) *Handler {
	isDev := os.Getenv("ENV") == "development"

	return &Handler{
		hub: hub,
		upgrader: gws.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// In development, allow all origins. In production, restrict to feedthemall.org.
			CheckOrigin: func(r *http.Request) bool {
				if isDev {
					return true
				}
				origin := r.Header.Get("Origin")
				return strings.HasPrefix(origin, "https://feedthemall.org") ||
					strings.HasPrefix(origin, "https://www.feedthemall.org")
			},
		},
		jwtSecret: []byte(jwtSecret),
	}
}

// ServeWS upgrades the HTTP connection to WebSocket, optionally authenticates the user,
// registers the client with the hub, and starts the read/write pumps.
//
// Route: GET /ws?token=<JWT>   (token is optional)
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Optional JWT authentication via query param.
	userID := h.extractUserID(r)

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// gorilla already wrote an HTTP error response on failure.
		slog.Warn("ws: upgrade failed", "err", err)
		return
	}

	client := NewClient(h.hub, conn, userID)
	h.hub.Register(client)

	// Start write pump in a separate goroutine.
	go client.WritePump()
	// Read pump runs in the current goroutine (blocks until disconnect).
	client.ReadPump()
}

// extractUserID extracts and validates the JWT token from the "token" query param.
// Returns an empty string if the token is absent or invalid (anonymous client).
func (h *Handler) extractUserID(r *http.Request) string {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		return ""
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return h.jwtSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}
	sub, _ := claims["sub"].(string)
	return sub
}
