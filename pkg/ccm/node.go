package ccm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"k8s.io/mount-utils"
)

var _ csi.NodeServer = (*driver)(nil)

const defaultMode = os.FileMode(0644)
const storageDir = "/csi-ccm-data"

func (d *driver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	start := time.Now()
	logger.V(2).Info("node publish volume called, target: " + req.TargetPath)
	if req.VolumeContext == nil {
		publishErr.WithLabelValues("", "missing volume context").Inc()
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume volume context must be provided")
	}
	if req.VolumeContext["name"] == "" {
		publishErr.WithLabelValues("", "missing volume context name field").Inc()
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume volume context name field should be set")
	}
	configMap := req.VolumeContext["name"]

	if req.VolumeId == "" {
		publishErr.WithLabelValues(configMap, "missing volume id").Inc()
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume volume id must be provided")
	}
	if req.TargetPath == "" {
		publishErr.WithLabelValues(configMap, "missing target path").Inc()
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume target path must be provided")
	}
	if req.VolumeCapability == nil {
		publishErr.WithLabelValues(configMap, "missing volume capabilities").Inc()
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume volume capability must be provided")
	}

	mnt := req.VolumeCapability.GetMount()
	options := mnt.MountFlags
	options = append(options, "bind")
	if req.Readonly {
		options = append(options, "ro")
	}

	fsType := "ext4"
	if mnt.FsType != "" {
		fsType = mnt.FsType
	}

	meta := &ClusterConfigMapMeta{
		Name:       configMap,
		Created:    start,
		Mode:       req.VolumeContext["mode"],
		VolumeID:   req.VolumeId,
		TargetPath: req.TargetPath,
		FSType:     fsType,
		BindOpts:   options,
	}
	if _, err := meta.FileMode(); err != nil {
		publishErr.WithLabelValues(configMap, "invalid volume mode").Inc()
		logger.Error(err, fmt.Sprintf("discarding invalid mode %q for volume %q", meta.Mode, req.VolumeId))
		meta.Mode = ""
	}

	if err := d.publisher.Populate(ctx, meta); err != nil {
		publishErr.WithLabelValues(configMap, "failed to populate volume contents").Inc()
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to populate volume %q: %s", req.VolumeId, err.Error()))
	}
	if err := d.publisher.Mount(ctx, meta); err != nil {
		publishErr.WithLabelValues(configMap, "failed to mount volume contents").Inc()
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to mount volume %q: %s", req.VolumeId, err.Error()))
	}
	publish.WithLabelValues(configMap).Inc()
	publishTime.WithLabelValues(configMap).Observe(time.Since(start).Seconds())
	return &csi.NodePublishVolumeResponse{}, nil
}

func (d *driver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	start := time.Now()
	if req.VolumeId == "" {
		unpublishErr.WithLabelValues("", "missing volume id").Inc()
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume Volume ID must be provided")
	}
	if req.TargetPath == "" {
		unpublishErr.WithLabelValues("", "missing target path").Inc()
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume Target Path must be provided")
	}

	logger.V(2).Info(fmt.Sprintf("node unpublish volume called for volume id %q target path %q", req.VolumeId, req.TargetPath))

	configMap := "unknown"
	meta, err := ReadMetadata(req.VolumeId)
	if err != nil {
		// legacy code path
		logger.Info(fmt.Sprintf("missing metadata for volume %q, attempting to handle unpublish of %q gracefully: %s", req.VolumeId, req.TargetPath, err.Error()))
	} else {
		if meta.TargetPath != req.TargetPath {
			unpublishErr.WithLabelValues(meta.Name, "metadata out of sync").Inc()
			logger.Info(fmt.Sprintf("requested unpublish dir %q does not match metadata dir %q for volume %q (configmap: %q), metadata out of sync.", req.TargetPath, meta.TargetPath, req.VolumeId, meta.Name))
		}
		configMap = meta.Name
	}
	mounter := mount.New("")
	if err = mounter.Unmount(req.TargetPath); err != nil {
		// check if volume was already unmounted
		mounts, listErr := mounter.List()
		if listErr != nil {
			logger.Error(listErr, "failed to list mounts when unmounting volume")
			unpublishErr.WithLabelValues(configMap, "failed to unmount volume").Inc()
			// return the original err to the user
			return nil, status.Error(codes.Internal, err.Error())
		}
		for _, mountpoint := range mounts {
			if strings.HasPrefix(mountpoint.Path, req.TargetPath) {
				logger.Error(err, "failed to unmount volume, mount still exists")
				unpublishErr.WithLabelValues(configMap, "failed to unmount volume").Inc()
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
		// volume was already unmounted, continue unpublishing
		logger.V(2).Info(fmt.Sprintf("failed to unmount volume %q, err was %q, did not detect the path in the system mounts, assuming it was already unmounted successfully", req.VolumeId, err.Error()))
		unpublishErr.WithLabelValues(configMap, "volume was already unmounted").Inc()
	}
	logger.V(2).Info(fmt.Sprintf("node unpublish volume succeeded for volume id %q target path %q", req.VolumeId, req.TargetPath))
	doCleanup(configMap, req.VolumeId, req.TargetPath)
	unpublish.WithLabelValues(configMap).Inc()
	unpublishTime.WithLabelValues(configMap).Observe(time.Since(start).Seconds())
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (d *driver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	panic("implement me")
}

func (d *driver) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	panic("implement me")
}

func (d *driver) NodeGetVolumeStats(ctx context.Context, in *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	panic("implement me")
}

func (d *driver) NodeExpandVolume(ctx context.Context, in *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	panic("implement me")
}

func (d *driver) NodeGetCapabilities(ctx context.Context, in *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	nscaps := []*csi.NodeServiceCapability{
		{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: csi.NodeServiceCapability_RPC_UNKNOWN,
				},
			},
		},
	}
	logger.V(2).Info("node get capabilities called")
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: nscaps,
	}, nil
}

func (d *driver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	logger.V(2).Info("node get info called")
	return &csi.NodeGetInfoResponse{
		NodeId: d.nodeID,
	}, nil
}
