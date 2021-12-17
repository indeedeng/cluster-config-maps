package ccm

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"indeed.com/compute-platform/cluster-config-map/apis/clusterconfigmap/v1alpha1"

	"k8s.io/client-go/kubernetes"

	"k8s.io/mount-utils"
)

//go:generate go run github.com/vektra/mockery/v2 --name=VolumePublisher --inpackage --structname=mockVolumePublisher

// VolumePublisher handles the mounting of volume and populating the content for the volume mount.
type VolumePublisher interface {
	// Mount will mount the target path of the cluster config map to the storage directory for the volume id, unless it's already mounted.
	Mount(ctx context.Context, meta *ClusterConfigMapMeta) error
	// Populate fetches the contents of the cluster config map from kubernetes and synthesizes the files and records their metadata.
	Populate(ctx context.Context, meta *ClusterConfigMapMeta) error
}

type nodePublisher struct {
	client *kubernetes.Clientset
}

var _ VolumePublisher = (*nodePublisher)(nil)

func (n *nodePublisher) Mount(ctx context.Context, meta *ClusterConfigMapMeta) error {
	mounter := mount.New("")

	// clean up the mount if it already exists but is not valid
	notMnt, err := mounter.IsLikelyNotMountPoint(meta.TargetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(meta.TargetPath, 0750); err != nil {
				return fmt.Errorf("failed to create mount target %q: %w", meta.TargetPath, err)
			}
			notMnt = true
		} else {
			return fmt.Errorf("mount target exists, is not mountable %q: %w", meta.TargetPath, err)
		}
	}
	if !notMnt {
		logger.V(2).Info("volume " + meta.VolumeID + " is already mounted")
		return nil
	}

	logger.V(2).Info("mounting the volume")
	if err := mounter.Mount(meta.Directory.Path, meta.TargetPath, meta.FSType, meta.BindOpts); err != nil {
		return fmt.Errorf("failed to bind mount %q to %q with fs %q", meta.Directory.Path, meta.TargetPath, meta.FSType)
	}

	logger.V(2).Info("bind mounting the volume is finished")
	return nil
}

func (n *nodePublisher) Populate(ctx context.Context, ccmm *ClusterConfigMapMeta) error {
	dir, err := ccmm.DataDir()
	if err != nil {
		return err
	}
	absPath := fmt.Sprintf("/apis/%s/%s/clusterconfigmaps/%s", v1alpha1.Group, v1alpha1.Version, ccmm.Name)
	logger.V(3).Info("querying for ccm: " + absPath)
	ccmBytes, err := n.client.RESTClient().Get().AbsPath(absPath).DoRaw(ctx)
	if err != nil {
		return fmt.Errorf("failed to read cluster configmap: %w", err)
	}
	var ccm v1alpha1.ClusterConfigMap
	if err := json.NewDecoder(bytes.NewReader(ccmBytes)).Decode(&ccm); err != nil {
		logger.V(5).Info("failed to decide cluster config map: %s", string(ccmBytes))
		return fmt.Errorf("failed to decode cluster configmap: %w", err)
	}

	meta := DirectoryMeta{
		Path:     dir,
		Contents: make([]ContentMeta, 0, len(ccm.Data)),
	}
	mode, _ := ccmm.FileMode()

	sha := sha512.New()
	for filename, contents := range ccm.Data {
		target := path.Join(dir, filename)
		logger.V(5).Info("writing data to target " + target)

		err = ioutil.WriteFile(target, []byte(contents), mode)
		if err != nil {
			return fmt.Errorf("failed to write configmap to target %q: %w", target, err)
		}
		sha.Reset()
		_, _ = sha.Write([]byte(contents))
		checksumStr := hex.EncodeToString(sha.Sum(nil))
		meta.Contents = append(meta.Contents, ContentMeta{
			Filename: filename,
			SHA512:   checksumStr,
		})
	}

	ccmm.Directory = meta
	if err = ccmm.WriteMetadata(); err != nil {
		return fmt.Errorf("failed to persist metadata for volume %q: %w", ccmm.VolumeID, err)
	}

	return nil
}
