package controllers

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/H3Cki/Plotor/clients/binance"
	"github.com/H3Cki/Plotor/plotor"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var sessions = &sessionRegistry{
	sessions: map[string]*session{},
	mu:       &sync.Mutex{},
}

type sessionRegistry struct {
	sessions map[string]*session
	mu       *sync.Mutex
}

func (ss *sessionRegistry) add(s *session) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.sessions[s.Token()] = s
}

func (ss *sessionRegistry) delete(token string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	_, ok := ss.sessions[token]
	if !ok {
		return fmt.Errorf("session does not exist")
	}

	delete(ss.sessions, token)

	return nil
}

func (ss *sessionRegistry) get(token string) (*session, bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	s, ok := ss.sessions[token]
	return s, ok
}

func (ss *sessionRegistry) getAll(token string) ([]*session, bool) {
	s, ok := ss.get(token)
	if !ok {
		return []*session{}, false
	}

	return ss.hash(s.Hash()), true
}

func (ss *sessionRegistry) hash(hash []byte) []*session {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	sessions := []*session{}

	for _, s := range ss.sessions {
		if bytes.Equal(s.Hash(), hash) {
			sessions = append(sessions, s)
		}
	}

	return sessions
}

type session struct {
	token       string
	hash        []byte
	PlotOrderer *plotor.PlotOrderer
}

func newSession(hash []byte, po *plotor.PlotOrderer) *session {
	return &session{
		hash:        hash,
		token:       uuid.NewString(),
		PlotOrderer: po,
	}
}

func (s *session) Token() string {
	return s.token
}

func (s *session) Hash() []byte {
	return s.hash
}

type createSessionRequest struct {
	Client string          `json:"client"`
	Auth   json.RawMessage `json:"auth"`
}

type createSessionResponse struct {
	Token string
	Error string
}

func CreateSession() func(c *gin.Context) {
	return func(c *gin.Context) {
		// Parse body
		ea := createSessionRequest{}
		if err := c.BindJSON(&ea); err != nil {
			c.IndentedJSON(http.StatusBadRequest, createSessionResponse{Error: err.Error()})
			return
		}

		hash, err := hash(ea)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, createSessionResponse{Error: err.Error()})
			return
		}

		exchange, err := client(ea.Client, ea.Auth)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, createSessionResponse{Error: err.Error()})
			return
		}

		session := newSession(hash, plotor.NewPlotOrderer(exchange))
		if _, ok := sessions.get(session.Token()); ok {
			c.IndentedJSON(http.StatusInternalServerError, createSessionResponse{Error: "error creating session: duplicate token"})
			return
		}

		sessions.add(session)

		c.IndentedJSON(http.StatusOK, createSessionResponse{
			Token: session.Token(),
		})
	}
}

type getSessionsResponse struct {
	Sessions []string
	Error    string
}

func GetSessions() func(c *gin.Context) {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

		ss, ok := sessions.getAll(token)
		if !ok {
			c.IndentedJSON(http.StatusBadRequest, getSessionsResponse{
				Error: "session not found",
			})
			return
		}

		tokens := []string{}

		for _, s := range ss {
			tokens = append(tokens, s.Token())
		}

		c.IndentedJSON(http.StatusOK, getSessionsResponse{
			Sessions: tokens,
		})
	}
}

func client(name string, auth []byte) (plotor.Client, error) {
	switch name {
	case "BINANCE_SPOT":
		creds := binance.SpotCredentials{}

		if err := json.Unmarshal(auth, &creds); err != nil {
			return nil, fmt.Errorf("error unmarshalling credentials: %w", err)
		}

		client := &binance.SpotClient{}

		if err := client.SetUp(creds); err != nil {
			return nil, fmt.Errorf("error setting up client: %w", err)
		}

		return client, nil
	case "BINANCE_FUTURES":
		creds := binance.FuturesCredentials{}

		if err := json.Unmarshal(auth, &creds); err != nil {
			return nil, fmt.Errorf("error unmarshalling credentials: %w", err)
		}

		client := &binance.FuturesClient{}

		if err := client.SetUp(creds); err != nil {
			return nil, fmt.Errorf("error setting up client: %w", err)
		}

		return client, nil
	}

	return nil, fmt.Errorf("unsupported client: %s", name)
}

type deleteSessionResponse struct {
	Error string
}

func DeleteSession() func(c *gin.Context) {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

		if err := sessions.delete(token); err != nil {
			c.IndentedJSON(http.StatusBadRequest, createSessionResponse{Error: err.Error()})
			return
		}

		c.IndentedJSON(http.StatusOK, deleteSessionResponse{})
	}
}

func hash(obj any) ([]byte, error) {
	h := sha256.New()
	_, err := h.Write([]byte(fmt.Sprint(obj)))
	if err != nil {
		return []byte{}, err
	}

	return h.Sum(nil), nil
}
