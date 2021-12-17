package ccm

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"time"

	"k8s.io/mount-utils"
)

func doCleanup(configMap, volumeID, targetPath string) {
	start := time.Now()
	if err := cleanupDataDir(); err != nil {
		logger.Error(err, fmt.Sprintf("failed to cleanup on node unpublish volume for volume id %q and target path %q", volumeID, targetPath))
		unpublishErr.WithLabelValues(configMap, "failed to clean up volume data").Inc()
	}
	if err := cleanupMetadataDir(); err != nil {
		logger.Error(err, fmt.Sprintf("failed to cleanup metadata on node unpublish volume for volume id %q and target path %q", volumeID, targetPath))
		unpublishErr.WithLabelValues(configMap, "failed to clean up volume metadata").Inc()
	}
	cleanupTime.WithLabelValues().Observe(time.Since(start).Seconds())
}

// cleanupDataDir walks the contents of the data storage directory, and deletes any volume directories that
// do not have any mounts bound to them. This is more reliable than deleting just the volume at the time of the
// node unpublish request, because its possible we missed some requests, the daemon could have been down and the node
// out of sync of the cluster, etc.
func cleanupDataDir() error {
	dataDir := path.Join(storageDir, "data")
	dirEntries, err := os.ReadDir(dataDir)
	if err != nil {
		return fmt.Errorf("failed to list dir entries for %q: %w", dataDir, err)
	}

	mounter := mount.New("")
	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			// we shouldn't have any bare files here
			logger.Info("[cleanup] unexpected file in data dir: " + path.Join(dataDir, dirEntry.Name()))
			cleanupErr.WithLabelValues(dirEntry.Name(), "unexpected file in metadata dir").Inc()
			continue
		}

		volumeID := dirEntry.Name()
		dataPath := path.Join(dataDir, volumeID)
		logger.V(6).Info("[cleanup] checking mount state of " + dataPath)
		refs, err := mounter.GetMountRefs(dataPath)
		if err != nil {
			logger.Error(err, "cleanup failed to lookup refs for "+dataPath+" - skipping...")
			cleanupErr.WithLabelValues(volumeID, "error listing mount refs").Inc()
			continue
		}
		if len(refs) == 0 {
			logger.V(6).Info(fmt.Sprintf("[cleanup] mount refs for %s: %v", dataPath, refs))
			if err := os.RemoveAll(dataPath); err != nil {
				logger.Error(err, "cleanup failed to delete "+dataPath+" - skipping...")
				cleanupErr.WithLabelValues(volumeID, "removing metadata dir failed").Inc()
				continue
			}
			logger.V(6).Info(fmt.Sprintf("[cleanup] deleted %s successfully", dataPath))
		}
	}
	return nil
}

// cleanupMetadataDir walks the contents of the metadata storage directory, and deletes any metadata directories that
// do not have a corresponding volume mount directory.
func cleanupMetadataDir() error {
	metadataDir := path.Join(storageDir, "metadata")
	dirEntries, err := os.ReadDir(metadataDir)
	if err != nil {
		return fmt.Errorf("failed to list dir entries for %q: %w", metadataDir, err)
	}
	dataDir := path.Join(storageDir, "data")

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			// we shouldn't have any bare files here
			logger.Info("[cleanup] unexpected file in metadata dir: " + path.Join(metadataDir, dirEntry.Name()))
			cleanupErr.WithLabelValues(dirEntry.Name(), "unexpected file in metadata dir").Inc()
			continue
		}

		volumeID := dirEntry.Name()
		metadataPath := path.Join(metadataDir, volumeID)
		dataPath := path.Join(dataDir, volumeID)
		_, err := os.Stat(dataPath)
		if err == nil {
			logger.V(6).Info(fmt.Sprintf("[cleanup] metadata for %s appears to be in use, no cleanup necessary", metadataPath))
			continue
		}
		if !errors.Is(err, fs.ErrNotExist) {
			logger.Error(err, fmt.Sprintf("[cleanup] unexpected error stating data path for volume %q", volumeID))
			cleanupErr.WithLabelValues(volumeID, "unexpected error stating metadata dir").Inc()
			continue
		}
		logger.V(6).Info(fmt.Sprintf("[cleanup] removing metadata for %q", metadataPath))
		if err := os.RemoveAll(metadataPath); err != nil {
			logger.Error(err, "cleanup failed to delete "+metadataPath+" - skipping...")
			cleanupErr.WithLabelValues(volumeID, "removing metadata dir failed").Inc()
			continue
		}
		logger.V(6).Info(fmt.Sprintf("[cleanup] deleted %s successfully", metadataPath))
	}
	return nil
}
