package mcpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/zagvozdeen/again/internal/httpx"
	"github.com/zagvozdeen/again/internal/models"
	"github.com/zagvozdeen/again/internal/store"
)

type Server struct {
	log   *slog.Logger
	store *store.Store
}

func New(log *slog.Logger, storage *store.Store) http.Handler {
	s := &Server{log: log, store: storage}

	server := mcp.NewServer(&mcp.Implementation{Name: "again", Version: "0.1.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "whoami",
		Description: "Return the account the API token belongs to.",
	}, s.whoami)

	handler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true, Logger: log},
	)
	return s.withUser(handler)
}

type userKey struct{}

func (s *Server) withUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token == "" {
			w.Header().Set("WWW-Authenticate", "Bearer")
			httpx.WriteError(w, r, s.log, httpx.ErrUnauthorized)
			return
		}
		user, err := s.store.GetUserByAPIToken(r.Context(), token)
		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				w.Header().Set("WWW-Authenticate", "Bearer error=\"invalid_token\"")
				err = httpx.ErrUnauthorized
			}
			httpx.WriteError(w, r, s.log, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey{}, user)))
	})
}

type whoamiInput struct{}

type whoamiOutput struct {
	Email      string `json:"email"`
	UILanguage string `json:"ui_language"`
}

func (s *Server) whoami(ctx context.Context, _ *mcp.CallToolRequest, _ whoamiInput) (*mcp.CallToolResult, whoamiOutput, error) {
	user, ok := ctx.Value(userKey{}).(*models.User)
	if !ok {
		return nil, whoamiOutput{}, errors.New("unauthorized")
	}
	return nil, whoamiOutput{Email: user.Email, UILanguage: user.UILanguage}, nil
}
