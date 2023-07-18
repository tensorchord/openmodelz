package runtime

import (
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/errdefs"
	v1 "k8s.io/api/core/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func (r Runtime) InferenceExec(ctx *gin.Context, namespace, instance string,
	commands []string, tty bool,
) error {
	req := r.restClient.Post().
		Resource("pods").
		Name(instance).
		Namespace(namespace).
		SubResource("exec")
	req.VersionedParams(&v1.PodExecOptions{
		Command: commands,
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     tty,
	}, clientsetscheme.ParameterCodec)

	result := req.Do(ctx)
	logrus.Infof("result: %v", result.Error())

	logrus.Info(req.URL())
	_, err := remotecommand.NewSPDYExecutor(r.clientConfig, "POST", req.URL())
	if err != nil {
		return errdefs.System(err)
	}

	dt, err := rest.TransportFor(r.clientConfig)
	if err != nil {
		return errdefs.System(err)
	}
	proxyServer := httputil.NewSingleHostReverseProxy(req.URL())
	proxyServer.Transport, err = rest.HTTPWrappersForConfig(r.clientConfig, dt)
	if err != nil {
		return errdefs.System(err)
	}

	proxyServer.ServeHTTP(ctx.Writer, ctx.Request)
	return nil

	// remoteOptions := remotecommand.StreamOptions{
	// 	Tty: tty,
	// }

	// if tty {
	// 	// Setting up the streaming http interface.
	// 	inStream, outStream, err := HijackConnection(ctx.Writer)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer inStream.Close()

	// 	remoteOptions.Stdin = inStream
	// 	remoteOptions.Stdout = outStream
	// 	remoteOptions.Stderr = outStream

	// 	if _, ok := ctx.Request.Header["Upgrade"]; ok {
	// 		contentType := "application/vnd.openmodelz.stream"
	// 		logrus.Debugf("Upgrading to %s", contentType)
	// 		ctx.String(http.StatusSwitchingProtocols, contentType, "HTTP/1.1 101 UPGRADED\r\nContent-Type: "+contentType+"\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n")
	// 	} else {
	// 		fmt.Fprint(outStream, "HTTP/1.1 200 OK\r\nContent-Type: application/vnd.openmodelz.stream\r\n")
	// 	}

	// 	if err := ctx.Writer.Header().WriteSubset(outStream, nil); err != nil {
	// 		return err
	// 	}
	// 	fmt.Fprint(outStream, "\r\n")

	// 	if err := exec.StreamWithContext(ctx.Request.Context(), remoteOptions); err != nil {
	// 		return errdefs.System(err)
	// 	}
	// } else {
	// 	if err := exec.StreamWithContext(ctx.Request.Context(), remoteOptions); err != nil {
	// 		return errdefs.System(err)
	// 	}
	// }
	// return nil
}

// HijackConnection interrupts the http response writer to get the
// underlying connection and operate with it.
func HijackConnection(w http.ResponseWriter) (io.ReadCloser, io.Writer, error) {
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		return nil, nil, err
	}
	// Flush the options to make sure the client sets the raw mode
	_, _ = conn.Write([]byte{})
	return conn, conn, nil
}
