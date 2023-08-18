package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/pkg/log"
)

// @Summary     Get the inference logs.
// @Description Get the inference logs.
// @Tags        log
// @Accept      json
// @Produce     json
// @Param       namespace query    string true  "Namespace"
// @Param       name      query    string true  "Name"
// @Param       instance  query    string false "Instance"
// @Param       tail      query    int    false "Tail"
// @Param       follow    query    bool   false "Follow"
// @Param       since     query    string false "Since"
// @Param       end       query    string false "End"
// @Success     200       {object} []types.Message
// @Router      /system/logs/inference [get]
func (s *Server) handleInferenceLogs(c *gin.Context) error {
	return s.getLogsFromRequester(c, s.deploymentLogRequester)
}

func (s Server) getLogsFromRequester(c *gin.Context, requester log.Requester) error {
	cn, ok := c.Writer.(http.CloseNotifier)
	if !ok {
		return NewError(http.StatusNotFound, errors.New("LogHandler: response is not a CloseNotifier, required for streaming response"), "log-get")
	}
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return NewError(http.StatusNotFound, errors.New("LogHandler: response is not a Flusher, required for streaming response"), "log-get")
	}

	var req types.LogRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return NewError(http.StatusBadRequest, err, "log-get")
	}
	_ = cn

	timeout := s.config.Inference.LogTimeout
	if req.Follow {
		// use a much larger timeout for streaming log
		timeout = time.Hour
	}

	ctx, cancelQuery := context.WithTimeout(c.Request.Context(), timeout)
	defer cancelQuery()

	messages, err := requester.Query(ctx, req)
	if err != nil {
		return errFromErrDefs(err, "log-get")
	}

	// Send the initial headers saying we're gonna stream the response.
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("Connection", "Keep-Alive")
	flusher.Flush()

	defer flusher.Flush()
	defer c.Writer.Write([]byte{})
	defer flusher.Flush()

	jsonEncoder := json.NewEncoder(c.Writer)
	for messages != nil {
		select {
		case <-cn.CloseNotify():
			s.logger.WithField("req", req).
				Debug("client closed connection")
			return nil
		case msg, ok := <-messages:
			if !ok {
				s.logger.WithField("req", req).
					Debug("log stream closed")
				messages = nil
				return nil
			}

			// serialize and write the msg to the http ResponseWriter
			err := jsonEncoder.Encode(msg)
			if err != nil {
				// can't actually write the status header here so we should json serialize an error
				// and return that because we have already sent the content type and status code
				s.logger.WithError(err).Error("LogHandler: failed to serialize log message")
				// write json error message here ?
				jsonEncoder.Encode(types.Message{Text: "failed to serialize log message"})
				flusher.Flush()
				return nil
			}

			flusher.Flush()
		}
	}
	return nil
}
