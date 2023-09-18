package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runtimemock "github.com/tensorchord/openmodelz/agent/pkg/runtime/mock"
)

var (
	ctrl        *gomock.Controller
	mockRuntime *runtimemock.MockRuntime
	server      *Server
)

func mkContext(method string, path string, header map[string][]string, body io.Reader) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	if c == nil {
		panic(c)
	}
	req, _ := http.NewRequest(method, path, body)
	for k, vs := range header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	c.Request = req
	return c
}

func mkJsonBodyContext(method string, path string, header map[string][]string, body any) *gin.Context {
	jsonValue, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	return mkContext(method, path, header, bytes.NewBuffer(jsonValue))
}

func setQuery(c *gin.Context, query map[string]string) {
	params, _ := url.ParseQuery(c.Request.URL.RawQuery)
	for k, v := range query {
		params.Set(k, v)
	}
	c.Request.URL.RawQuery = params.Encode()
}

func setParam(c *gin.Context, param map[string]string) {
	for k, v := range param {
		c.Params = []gin.Param{{Key: k, Value: v}}
	}
}

func TestBuilder(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	RegisterFailHandler(Fail)
	RunSpecs(t, "server")
}

var _ = BeforeSuite(func() {
	ctrl = gomock.NewController(GinkgoT())
	mockRuntime = runtimemock.NewMockRuntime(ctrl)
})
