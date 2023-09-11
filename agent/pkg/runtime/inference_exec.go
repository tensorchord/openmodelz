package runtime

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/tensorchord/openmodelz/agent/errdefs"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// ctrl+d to close terminal.
	endOfTransmission = "\u0004"
)

func (r generalRuntime) InferenceExec(ctx *gin.Context, namespace, instance string,
	commands []string, tty bool) error {
	pod, err := r.kubeClient.CoreV1().Pods(namespace).Get(
		ctx.Request.Context(), instance, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return errdefs.NotFound(errors.New("inference instance not found"))
		}
		return errdefs.System(err)
	}

	if pod.Status.Phase != v1.PodRunning {
		return errdefs.Unavailable(errors.New("inference instance is not running"))
	}

	req := r.kubeClient.CoreV1().RESTClient().Post().
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

	if tty {
		t, err := newTerminalSession(fmt.Sprintf("exec/%s/%s/%s", namespace, instance, rand.String(5)),
			ctx.Request, ctx.Writer)
		if err != nil {
			return err
		}
		defer t.Close()

		logrus.WithField("exec", exec).Debugf("executing command")
		if err = exec.StreamWithContext(ctx.Request.Context(), remotecommand.StreamOptions{
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
		logrus.Debugf("running without tty")
		if err := exec.StreamWithContext(ctx.Request.Context(),
			remotecommand.StreamOptions{
				Stdout: ctx.Writer,
				Stderr: ctx.Writer,
				Tty:    tty,
			}); err != nil {
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
	ID   string `json:"id,omitempty"`
	Op   string `json:"op,omitempty"`
	Data string `json:"data,omitempty"`
	Rows uint16 `json:"rows,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
}

// TerminalSession
type TerminalSession struct {
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
		return copy(p, endOfTransmission), err
	}
	logrus.Debugf("%s: read json: %v", t.ID, msg)
	switch msg.Op {
	case "stdin":
		logrus.WithField("remote", t.wsConn.RemoteAddr()).Debugf("%s: read %d bytes: %s", t.ID, len(msg.Data), msg.Data)
		size := copy(p, msg.Data)
		logrus.WithField("remote", t.wsConn.RemoteAddr()).Debugf("%s: copied %d bytes: %s", t.ID, size, p)
		return size, nil
	case "resize":
		t.sizeChan <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	default:
		logrus.WithField("remote", t.wsConn.RemoteAddr()).Debugf("%s: unknown message type '%s'", t.ID, msg.Op)
		return copy(p, endOfTransmission), fmt.Errorf("unknown message type '%s'", msg.Op)
	}
}

// Write handles process->pty stdout
// Called from remotecommand whenever there is any output
func (t *TerminalSession) Write(p []byte) (int, error) {
	msg := TerminalMessage{
		Op:   "stdout",
		Data: string(p),
	}

	logrus.WithField("remote", t.wsConn.RemoteAddr()).Debugf("%s: write %d bytes: %s", t.ID, len(p), string(p))

	if err := t.wsConn.WriteJSON(msg); err != nil {
		logrus.WithField("remote", t.wsConn.RemoteAddr()).Debugf("write message failed: %v", err)
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
