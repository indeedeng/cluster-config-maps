package ccm

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	ctrl "sigs.k8s.io/controller-runtime"
)

var logger = ctrl.Log.WithName("driver")

// Driver implements the following CSI interfaces:
//
//	csi.IdentityServer
//	csi.NodeServer
type driver struct {
	nodeID   string
	endpoint string

	publisher VolumePublisher

	volumeLock sync.Mutex
	volumeBusy map[string]bool

	srv *grpc.Server
}

func newDriver(nodeID, endpoint string, publisher VolumePublisher) *driver {
	return &driver{
		nodeID:     nodeID,
		endpoint:   endpoint,
		publisher:  publisher,
		volumeBusy: make(map[string]bool),
	}
}

func NewDriver(endpoint string) (*driver, error) {
	host, _ := os.Hostname()

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return newDriver(host, endpoint, &nodePublisher{client: client}), nil
}

func (d *driver) Run() error {
	logger.Info("grpc server starting...")

	u, err := url.Parse(d.endpoint)
	if err != nil {
		return fmt.Errorf("unable to parse address %q: %w", d.endpoint, err)
	}

	addr := path.Join(u.Host, filepath.FromSlash(u.Path))
	if u.Host == "" {
		addr = filepath.FromSlash(u.Path)
	}

	// CSI plugins talk only over UNIX sockets currently
	if u.Scheme != "unix" {
		return fmt.Errorf("currently only unix domain sockets are supported, have: %s", u.Scheme)
	}
	// remove the socket if it's already there. This can happen if we
	// deploy a new version and the socket was created from the old running
	// plugin.
	logger.Info("cleaning up socket: " + addr)
	if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove unix domain socket file %s: %w", addr, err)
	}

	listener, err := net.Listen(u.Scheme, addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	errHandler := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			ctrl.Log.WithName("grpc").Error(err, "method failed: "+info.FullMethod)
		}
		return resp, err
	}

	doCleanup()

	d.srv = grpc.NewServer(grpc.UnaryInterceptor(errHandler))
	csi.RegisterIdentityServer(d.srv, d)
	csi.RegisterNodeServer(d.srv, d)

	logger.Info("grpc server started")
	return d.srv.Serve(listener)
}

func (d *driver) Stop() {
	logger.Info("server shutting down...")
	d.srv.Stop()
}
