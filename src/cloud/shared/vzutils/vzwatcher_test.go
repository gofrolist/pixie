package vzutils_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pixielabs.ai/pixielabs/src/cloud/shared/messages"
	messagespb "pixielabs.ai/pixielabs/src/cloud/shared/messagespb"
	"pixielabs.ai/pixielabs/src/cloud/shared/vzutils"
	"pixielabs.ai/pixielabs/src/cloud/vzmgr/vzmgrpb"
	mock_vzmgrpb "pixielabs.ai/pixielabs/src/cloud/vzmgr/vzmgrpb/mock"
	"pixielabs.ai/pixielabs/src/utils"
	"pixielabs.ai/pixielabs/src/utils/testingutils"
)

func TestVzWatcher(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "no error",
			expectError: false,
		},
		{
			name:        "error",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockVZMgr := mock_vzmgrpb.NewMockVZMgrServiceClient(ctrl)

			existingVzID := uuid.Must(uuid.NewV4())
			existingOrgID := uuid.Must(uuid.NewV4())
			existingK8sUID := "testUID"

			mockVZMgr.
				EXPECT().
				GetViziersByShard(gomock.Any(), &vzmgrpb.GetViziersByShardRequest{
					FromShardID: "00",
					ToShardID:   "bb",
				}).
				Return(&vzmgrpb.GetViziersByShardResponse{
					Viziers: []*vzmgrpb.GetViziersByShardResponse_VizierInfo{
						{
							VizierID: utils.ProtoFromUUID(existingVzID),
							OrgID:    utils.ProtoFromUUID(existingOrgID),
							K8sUID:   existingK8sUID,
						},
					},
				}, nil)

			nc, natsCleanup := testingutils.MustStartTestNATS(t)
			defer natsCleanup()

			w, err := vzutils.NewWatcher(nc, mockVZMgr, "00", "bb")
			require.NoError(t, err)

			var wg sync.WaitGroup
			wg.Add(2)

			if test.expectError {
				wg.Add(1)
			}

			defer wg.Wait()

			newVzID := uuid.Must(uuid.NewV4())
			newOrgID := uuid.Must(uuid.NewV4())
			newK8sUID := "testUID"

			w.RegisterErrorHandler(func(id uuid.UUID, orgID uuid.UUID, uid string, err error) {
				defer wg.Done()
				assert.Equal(t, existingVzID, id)
				assert.Equal(t, existingOrgID, orgID)
				assert.Equal(t, existingK8sUID, uid)
			})

			w.RegisterVizierHandler(func(id uuid.UUID, orgID uuid.UUID, uid string) error {
				defer wg.Done()

				switch id {
				case existingVzID:
					assert.Equal(t, existingOrgID, orgID)
					assert.Equal(t, existingK8sUID, uid)
					if test.expectError {
						return errors.New("Some error")
					}
				case newVzID:
					assert.Equal(t, newOrgID, orgID)
					assert.Equal(t, newK8sUID, uid)
				default:
					t.Fatal("Called Vizier handler with unexpected vizier")
				}
				return nil
			})

			msg := &messagespb.VizierConnected{
				VizierID: utils.ProtoFromUUID(newVzID),
				OrgID:    utils.ProtoFromUUID(newOrgID),
				K8sUID:   newK8sUID,
			}
			b, err := msg.Marshal()
			require.NoError(t, err)
			nc.Publish(messages.VizierConnectedChannel, b)
		})
	}
}
