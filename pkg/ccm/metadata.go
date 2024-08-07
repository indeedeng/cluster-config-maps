package ccm

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"
)

// ClusterConfigMapMeta contains metadata recorded when a cluster config map csi volume request occurs. This is useful
// for debugging, as you can easily inspect the metadata to rebuild the original volume request. Additionally,
// subsequent grpc requests, like the volume unpublish request, lack much of the metadata present in the publish request.
// Recording the original request details allows future requests to have a more complete view of the cluster config map volume.
type ClusterConfigMapMeta struct {
	Name       string        `json:"name"`
	Created    time.Time     `json:"created"`
	Mode       string        `json:"mode"`
	VolumeID   string        `json:"volumeId"`
	TargetPath string        `json:"targetPath"`
	FSType     string        `json:"fsType"`
	BindOpts   []string      `json:"bindOpts"`
	Directory  DirectoryMeta `json:"directory"`
}

// dir is a helper func which ensures the directory name for the volume id exists under the ccm data dir, or creates it if it does not.
func dir(name, volumeID string) (string, error) {
	dir := path.Join(storageDir, name, volumeID)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create data dir %q: %w", dir, err)
	}
	return dir, nil
}

// DataDir ensures the directory for the cluster config map csi data exists and returns the path.
func (c *ClusterConfigMapMeta) DataDir() (string, error) {
	return dir("data", c.VolumeID)
}

// MetadataDir ensures the directory for the cluster config map metadata exists and returns the path.
func (c *ClusterConfigMapMeta) MetadataDir() (string, error) {
	return dir("metadata", c.VolumeID)
}

// FileMode returns the unix permissions for files created for the cluster config map, or the default permissions if unset.
func (c *ClusterConfigMapMeta) FileMode() (os.FileMode, error) {
	if c.Mode == "" {
		return defaultMode, nil
	}
	if parsedMode, err := strconv.ParseUint(c.Mode, 8, 32); err == nil {
		return os.FileMode(parsedMode), nil
	} else {
		return defaultMode, fmt.Errorf("failed to parse mode %q for ccm %q: %w", c.Mode, c.Name, err)
	}
}

// WriteMetadata marshals and persists the json metadata of cluster config map to the filesystem.
func (c *ClusterConfigMapMeta) WriteMetadata() error {
	dir, err := c.MetadataDir()
	if err != nil {
		return fmt.Errorf("failed to get metadata dir: %w", err)
	}

	bytes, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return os.WriteFile(path.Join(dir, "metadata.json"), bytes, defaultMode)
}

// ReadMetadata unmarshalls and parses the json metadata of cluster config map from the filesystem.
func ReadMetadata(volumeID string) (*ClusterConfigMapMeta, error) {
	dir, err := dir("metadata", volumeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata dir: %w", err)
	}

	bytes, err := os.ReadFile(path.Join(dir, "metadata.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata.json for volume %q: %w", volumeID, err)
	}
	meta := &ClusterConfigMapMeta{}
	if err = json.Unmarshal(bytes, meta); err != nil {
		return nil, fmt.Errorf("failed unmarshal to metadata.json %q: %w", string(bytes), err)
	}
	return meta, nil
}

type DirectoryMeta struct {
	Path     string        `json:"path"`
	Contents []ContentMeta `json:"files"`
}

type ContentMeta struct {
	Filename string `json:"filename"`
	SHA512   string `json:"sha512"`
}
