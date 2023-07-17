package runtime

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/errdefs"
	"k8s.io/client-go/tools/remotecommand"
)

func (r Runtime) InferenceExec(ctx *gin.Context, namespace, instance string,
	commands []string,
) error {
	// req := r.kubeClient.(*kubernetes.Clientset).RESTClient().Post().
	// 	Resource("pods").
	// 	Name(instance).
	// 	Namespace(namespace).
	// 	SubResource("exec")

	req := r.restClient.Post().
		Resource("pods").
		Name(instance).
		Namespace(namespace).
		SubResource("exec")
	// req.VersionedParams(&v1.PodExecOptions{
	// 	Command: []string{"bash"},
	// 	Stdin:   false,
	// 	Stdout:  true,
	// 	Stderr:  true,
	// 	TTY:     false,
	// }, clientsetscheme.ParameterCodec)

	result := req.Do(ctx)
	logrus.Infof("result: %v", result.Error())

	logrus.Info(req.URL())
	exec, err := remotecommand.NewSPDYExecutor(r.clientConfig, "POST", req.URL())
	if err != nil {
		return errdefs.System(err)
	}

	if err := exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  ctx.Request.Body,
		Stdout: ctx.Writer,
		Stderr: ctx.Writer,
		Tty:    false,
	}); err != nil {
		return errdefs.System(err)
	}
	return nil
}
