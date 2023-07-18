package runtime

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/errdefs"
	v1 "k8s.io/api/core/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func (r Runtime) InferenceExec(ctx *gin.Context, namespace, instance string,
	commands []string, tty bool) error {
	req := r.restClient.Post().
		Resource("pods").
		Name(instance).
		Namespace(namespace).
		SubResource("exec")
	req.VersionedParams(&v1.PodExecOptions{
		Command: commands,
		Stdin:   tty,
		Stdout:  true,
		Stderr:  true,
		TTY:     tty,
	}, clientsetscheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(r.clientConfig, http.MethodPost, req.URL())
	if err != nil {
		return errdefs.System(err)
	}

	remoteOptions := remotecommand.StreamOptions{
		Stdout: ctx.Writer,
		Stderr: ctx.Writer,
		Tty:    tty,
	}

	if tty {
		t, err := newTerminalSession(fmt.Sprintf("exec/%s/%s", namespace, instance),
			ctx.Request, ctx.Writer)
		if err != nil {
			log.Println(err)
			return err
		}
		defer t.Close()

		if err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
			Stdin:             t,
			Stdout:            t,
			Stderr:            t,
			TerminalSizeQueue: t,
			Tty:               true,
		}); err != nil {
			// The response is already hijacked, so we can't return an error.
			logrus.Warnf("exec stream failed: %v", err)
			return nil
		}
	} else {
		if err := exec.StreamWithContext(ctx.Request.Context(), remoteOptions); err != nil {
			return errdefs.System(err)
		}
	}
	return nil
}

type PtyHandler interface {
	io.Reader
	io.Writer
	remotecommand.TerminalSizeQueue
}

const END_OF_TRANSMISSION = "\u0004"

// TerminalMessage is the messaging protocol between ShellController and TerminalSession.
//
// OP      DIRECTION  FIELD(S) USED  DESCRIPTION
// ---------------------------------------------------------------------
// bind    fe->be     SessionID      Id sent back from TerminalResponse
// stdin   fe->be     Data           Keystrokes/paste buffer
// resize  fe->be     Rows, Cols     New terminal size
// stdout  be->fe     Data           Output from the process
// toast   be->fe     Data           OOB message to be shown to the user
type TerminalMessage struct {
	Op   string `json:"op,omitempty"`
	Data string `json:"data,omitempty"`
	Rows uint16 `json:"rows,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
}

// TerminalSession
type TerminalSession struct {
	sync.Mutex
	ID       string
	wsConn   *websocket.Conn
	sizeChan chan remotecommand.TerminalSize
	doneChan chan struct{}
}

// TerminalSize handles pty->process resize events
// Called in a loop from remotecommand as long as the process is running
func (t *TerminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeChan:
		return &size
	case <-t.doneChan:
		return nil
	}
}

// Read handles pty->process messages (stdin, resize)
// Called in a loop from remotecommand as long as the process is running
func (t *TerminalSession) Read(p []byte) (int, error) {
	var msg TerminalMessage
	if err := t.wsConn.ReadJSON(&msg); err != nil {
		logrus.Debugf("%s: read json failed: %v", t.ID, err)
		return copy(p, END_OF_TRANSMISSION), err
	}
	switch msg.Op {
	case "stdin":
		logrus.Debugf("%s: read %d bytes: %s", t.ID, len(msg.Data), msg.Data)
		return copy(p, msg.Data), nil
	case "resize":
		t.sizeChan <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	default:
		logrus.Debugf("%s: unknown message type '%s'", t.ID, msg.Op)
		return copy(p, END_OF_TRANSMISSION), fmt.Errorf("unknown message type '%s'", msg.Op)
	}
}

// Write handles process->pty stdout
// Called from remotecommand whenever there is any output
func (t *TerminalSession) Write(p []byte) (int, error) {
	msg := TerminalMessage{
		Op:   "stdout",
		Data: string(p),
	}

	logrus.Debugf("%s: write %d bytes: %s", t.ID, len(p), string(p))

	if err := t.wsConn.WriteJSON(msg); err != nil {
		log.Printf("write message failed: %v", err)
		return 0, err
	}
	return len(p), nil
}

func (t *TerminalSession) Close() error {
	close(t.doneChan)
	return t.wsConn.Close()
}

func newTerminalSession(id string, r *http.Request, w http.ResponseWriter) (*TerminalSession, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &TerminalSession{
		ID:       id,
		wsConn:   conn,
		sizeChan: make(chan remotecommand.TerminalSize),
		doneChan: make(chan struct{}),
	}, nil
}
