package server

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("healthz", func() {
	BeforeEach(func() {
		server = &Server{
			router:        gin.New(),
			metricsRouter: gin.New(),
			runtime:       mockRuntime,
		}
	})
	It("healthz", func() {
		c := mkContext("GET", "/", nil, nil)
		err := server.handleHealthz(c)
		Expect(err).NotTo(HaveOccurred())
	})
})
