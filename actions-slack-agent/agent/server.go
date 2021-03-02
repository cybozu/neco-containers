package agent

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cybozu-go/well"
	"github.com/cybozu/neco-containers/actions-slack-agent/slack"
	"github.com/gin-gonic/gin"
)

// Server is a Slack agent server.
type Server struct {
	listenAddr  string
	listenPort  int
	slackClient slack.Notifier
}

func NewServer(
	listenAddr string,
	listenPort int,
	slackClient slack.Notifier,
) *Server {
	return &Server{
		listenAddr,
		listenPort,
		slackClient,
	}
}

func (s *Server) Start(_ context.Context) error {
	serv := &well.HTTPServer{
		Server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", s.listenAddr, s.listenPort),
			Handler: s.prepareRouter(),
		},
	}
	return serv.ListenAndServe()
}

func (s *Server) prepareRouter() http.Handler {
	router := gin.Default()
	router.POST("/slack/success", s.postSlackSuccess)
	router.POST("/slack/fail", s.postSlackFail)
	router.PATCH("/runner/label", s.patchRunnerLabel)
	return router
}

type slackPayload struct {
	PodName      string `json:"pod_name"`
	PodNamespace string `json:"pod_namespace"`
	JobName      string `json:"job_name"`
}

func (s *Server) postSlack(c *gin.Context, isSucceeded bool) {
	var p slackPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return

	}

	s.slackClient.Notify(
		context.Background(),
		p.JobName,
		p.PodNamespace,
		p.PodName,
		isSucceeded,
		time.Now(),
	)
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (s *Server) postSlackSuccess(c *gin.Context) {
	s.postSlack(c, true)
}

func (s *Server) postSlackFail(c *gin.Context) {
	s.postSlack(c, false)
}

func (s *Server) patchRunnerLabel(c *gin.Context) {
	panic("not implemented")
}
