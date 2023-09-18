package server

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/pkg/server/validator"
)

var _ = Describe("namespace delete", func() {
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
		err := server.handleNamespaceDelete(c)
		Expect(err).To(HaveOccurred())
	})
	It("invalid request - mock error", func() {
		mockRuntime.EXPECT().NamespaceDelete(gomock.Any(), gomock.Any()).Times(1).Return(errors.New("mock-error"))
		c := mkJsonBodyContext("GET", "/", nil, types.NamespaceRequest{
			Name: "mock-ns",
		})
		err := server.handleNamespaceDelete(c)
		Expect(err).To(HaveOccurred())
	})
	It("good request", func() {
		mockRuntime.EXPECT().NamespaceDelete(gomock.Any(), gomock.Any()).Times(1).Return(nil)
		c := mkJsonBodyContext("GET", "/", nil, types.NamespaceRequest{
			Name: "mock-ns",
		})
		err := server.handleNamespaceDelete(c)
		Expect(err).NotTo(HaveOccurred())
	})
})
