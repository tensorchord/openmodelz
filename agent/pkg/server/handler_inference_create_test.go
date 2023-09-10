package server

import (
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/pkg/server/validator"
	. "github.com/tensorchord/openmodelz/modelzetes/pkg/pointer"
)

var _ = Describe("inference create", func() {
	BeforeEach(func() {
		server = &Server{
			router:        gin.New(),
			metricsRouter: gin.New(),
			runtime:       mockRuntime,
			validator:     validator.New(),
		}
	})
	It("invalid request - nil", func() {
		c := mkContext("GET", "/", nil, nil)
		err := server.handleInferenceCreate(c)
		Expect(err).To(HaveOccurred())
	})
	It("invalid request - empty", func() {
		c := mkJsonBodyContext("GET", "/", nil, types.InferenceDeployment{})
		err := server.handleInferenceCreate(c)
		Expect(err).To(HaveOccurred())
	})
	It("good request", func() {
		mockRuntime.EXPECT().InferenceCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
		c := mkJsonBodyContext("GET", "/", nil, types.InferenceDeployment{
			Spec: types.InferenceDeploymentSpec{
				Name:  "abc",
				Image: "mock-image",
				Port:  Ptr(int32(123)),
			},
		})
		err := server.handleInferenceCreate(c)
		Expect(err).NotTo(HaveOccurred())
	})
})
