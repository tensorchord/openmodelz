package event

import (
	"github.com/golang/mock/gomock"
	"github.com/tensorchord/openmodelz/agent/pkg/query"
	querymock "github.com/tensorchord/openmodelz/agent/pkg/query/mock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("event", func() {
	var (
		ctrl *gomock.Controller
		mock *querymock.MockQuerier
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mock = querymock.NewMockQuerier(ctrl)
	})
	AfterEach(func() {
		ctrl.Finish()
	})
	It("create deployment event", func() {
		tcs := []struct {
			desp        string
			namespace   string
			deployment  string
			event       string
			message     string
			expectedErr bool
		}{
			{
				desp:        "basic deployment event",
				namespace:   "modelz-00000000-1111-1111-1111-000000000000",
				deployment:  "00000000-1111-1111-1111-000000000000",
				event:       "mock-event",
				message:     "mock-message",
				expectedErr: false,
			},
			{
				desp:        "namespace too short",
				namespace:   "modelz-",
				deployment:  "00000000-1111-1111-1111-000000000000",
				event:       "mock-event",
				message:     "mock-message",
				expectedErr: true,
			},
			{
				desp:        "namespace not start with defined prefix",
				namespace:   "bad-777888999000",
				deployment:  "00000000-1111-1111-1111-000000000000",
				event:       "mock-event",
				message:     "mock-message",
				expectedErr: true,
			},
			{
				desp:        "bad deployment id",
				namespace:   "modelz-00000000-1111-1111-1111-000000000000",
				deployment:  "bad-deployment",
				event:       "mock-event",
				message:     "mock-message",
				expectedErr: true,
			},
			{
				desp:        "bad user id",
				namespace:   "modelz-bad-user-id",
				deployment:  "bad-deployment",
				event:       "mock-event",
				message:     "mock-message",
				expectedErr: true,
			},
		}

		for _, tc := range tcs {
			mock.EXPECT().CreateDeploymentEvent(gomock.Any(), gomock.Any()).AnyTimes().Return(
				query.DeploymentEvent{}, nil)
			recorder := NewEventRecorder(mock)
			err := recorder.CreateDeploymentEvent(tc.namespace, tc.deployment, tc.event, tc.message)
			if tc.expectedErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		}
	})
})
