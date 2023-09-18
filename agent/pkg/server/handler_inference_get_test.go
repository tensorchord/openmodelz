package server

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/pkg/server/validator"
	. "github.com/tensorchord/openmodelz/modelzetes/pkg/pointer"
)

var _ = Describe("inference get", func() {
	BeforeEach(func() {
		server = &Server{
			router:        gin.New(),
			metricsRouter: gin.New(),
			runtime:       mockRuntime,
			validator:     validator.New(),
		}
	})
	It("invalid request - no namespace", func() {
		c := mkContext("GET", "/", nil, nil)
		err := server.handleInferenceGet(c)
		Expect(err).To(HaveOccurred())
	})
	It("invalid request - no name", func() {
		c := mkJsonBodyContext("GET", "/", nil, nil)
		setQuery(c, map[string]string{"namespace": "mock-namespace"})
		err := server.handleInferenceGet(c)
		Expect(err).To(HaveOccurred())
	})
	It("invalid request - mock error", func() {
		mockRuntime.EXPECT().InferenceGet(gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("mock-error"))
		c := mkJsonBodyContext("GET", "/", nil, nil)
		setQuery(c, map[string]string{"namespace": "mock-namespace"})
		setParam(c, map[string]string{"name": "mock-name"})
		err := server.handleInferenceGet(c)
		Expect(err).To(HaveOccurred())
	})
	It("good request", func() {
		mockRuntime.EXPECT().InferenceGet(gomock.Any(), gomock.Any()).Times(1).Return(Ptr(types.InferenceDeployment{}), nil)
		c := mkJsonBodyContext("GET", "/", nil, nil)
		setQuery(c, map[string]string{"namespace": "mock-namespace"})
		setParam(c, map[string]string{"name": "mock-name"})
		err := server.handleInferenceGet(c)
		Expect(err).NotTo(HaveOccurred())
	})
})
