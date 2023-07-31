package client // import "docker.io/go-docker"

import (
	"net/url"

	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
)

// HijackedResponse holds connection information for a hijacked request.
type HijackedResponse struct {
	Conn *websocket.Conn
}

// Close closes the hijacked connection and reader.
func (h *HijackedResponse) Close() {
	h.Conn.Close()
}

// postHijacked sends a POST request and hijacks the connection.
func (cli *Client) websocket(ctx context.Context, path string, query url.Values, headers map[string][]string) (HijackedResponse, error) {
	apiPath := cli.getAPIPath(ctx, path, nil)

	scheme := "ws"
	if cli.scheme == "https" {
		scheme = "wss"
	}

	apiURL := url.URL{
		Scheme:   scheme,
		Host:     cli.addr,
		Path:     apiPath,
		RawQuery: query.Encode(),
	}
	c, _, err := websocket.DefaultDialer.DialContext(ctx, apiURL.String(), nil)
	if err != nil {
		return HijackedResponse{}, err
	}

	return HijackedResponse{Conn: c}, err
}

func (h HijackedResponse) Read(p []byte) (int, error) {
	// Read message from websocket connection.
	tm := &TerminalMessage{}
	if err := h.Conn.ReadJSON(tm); err != nil {
		return 0, err
	}
	if tm.Op != "stdout" {
		return 0, nil
	}
	return copy(p, tm.Data), nil
}

func (h HijackedResponse) Write(p []byte) (int, error) {
	// Write message to websocket connection.
	tm := &TerminalMessage{
		Op:   "stdin",
		Data: string(p),
	}
	if err := h.Conn.WriteJSON(tm); err != nil {
		return 0, err
	}
	return len(p), nil
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
