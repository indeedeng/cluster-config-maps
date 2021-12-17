package ccm

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var _ csi.IdentityServer = (*driver)(nil)

func (d *driver) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	return &csi.GetPluginInfoResponse{
		Name:          "clusterconfigmaps.indeed.com",
		VendorVersion: "0.0.1",
	}, nil
}

func (d *driver) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	logger.V(2).Info("get plugin capabilities called")
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_UNKNOWN,
					},
				},
			},
		},
	}, nil
}

func (d *driver) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{}, nil
}
