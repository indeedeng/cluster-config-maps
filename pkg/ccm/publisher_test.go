package ccm

import (
	"context"
	"os"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_NodePublishVolume(t *testing.T) {
	type testcase struct {
		description string
		req         *csi.NodePublishVolumeRequest

		publisher func() *mockVolumePublisher
	}
	tests := []testcase{
		{
			description: "node publish volume should succeed with default permissions",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:   "test-volume-id",
				TargetPath: "/tmp/test-path",
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
				},
				VolumeContext: map[string]string{
					"name": "test-cluster-config-maps",
				},
			},
			publisher: func() *mockVolumePublisher {
				mockPublisher := &mockVolumePublisher{}
				mockPublisher.On("Populate", context.TODO(), mock.AnythingOfType("*ccm.ClusterConfigMapMeta")).Return(func(_ context.Context, meta *ClusterConfigMapMeta) error {
					require.Equal(t, "test-volume-id", meta.VolumeID)
					require.Equal(t, "/tmp/test-path", meta.TargetPath)
					require.Equal(t, "test-cluster-config-maps", meta.Name)
					return nil
				})
				mockPublisher.On("Mount", context.TODO(), mock.AnythingOfType("*ccm.ClusterConfigMapMeta")).Return(nil)
				return mockPublisher
			},
		},
		{
			description: "node publish volume should succeed with custom permissions",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:   "test-volume-id",
				TargetPath: "/tmp/test-path",
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
				},
				VolumeContext: map[string]string{
					"name": "test-cluster-config-maps",
					"mode": "0777",
				},
			},
			publisher: func() *mockVolumePublisher {
				mockPublisher := &mockVolumePublisher{}
				mockPublisher.On("Populate", context.TODO(), mock.AnythingOfType("*ccm.ClusterConfigMapMeta")).Return(func(_ context.Context, meta *ClusterConfigMapMeta) error {
					require.Equal(t, "test-volume-id", meta.VolumeID)
					require.Equal(t, "/tmp/test-path", meta.TargetPath)
					require.Equal(t, "test-cluster-config-maps", meta.Name)
					require.Equal(t, "0777", meta.Mode)
					mode, err := meta.FileMode()
					require.NoError(t, err)
					require.Equal(t, os.ModePerm, mode)
					return nil
				})
				mockPublisher.On("Mount", context.TODO(), mock.AnythingOfType("*ccm.ClusterConfigMapMeta")).Return(nil)
				return mockPublisher
			},
		},
	}

	for _, test := range tests {
		mockPublisher := test.publisher()
		driver := newDriver("test", "", mockPublisher)

		_, err := driver.NodePublishVolume(context.TODO(), test.req)
		require.NoError(t, err, test.description)
		mockPublisher.AssertExpectations(t)
	}
}

func Test_NodePublishVolume_Error(t *testing.T) {
	type testcase struct {
		description string
		req         *csi.NodePublishVolumeRequest

		err string
	}
	tests := []testcase{
		{
			description: "node publish volume should ensure a volume name",
			req: &csi.NodePublishVolumeRequest{
				VolumeId:   "test-volume-id",
				TargetPath: "/tmp/test-path",
				VolumeCapability: &csi.VolumeCapability{
					AccessType: &csi.VolumeCapability_Mount{
						Mount: &csi.VolumeCapability_MountVolume{},
					},
				},
				VolumeContext: map[string]string{},
			},
			err: "NodePublishVolume volume context name field should be set",
		},
	}

	for _, test := range tests {
		mockPublisher := &mockVolumePublisher{}
		driver := newDriver("test", "", mockPublisher)

		_, err := driver.NodePublishVolume(context.TODO(), test.req)
		require.Error(t, err, test.description)
		require.Contains(t, err.Error(), test.err, test.description)
		mockPublisher.AssertExpectations(t)
	}
}
