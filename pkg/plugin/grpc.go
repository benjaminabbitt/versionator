package plugin

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/benjaminabbitt/versionator/pkg/plugin/proto"
	goplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// rpcTimeout is the default timeout for gRPC calls.
const rpcTimeout = 10 * time.Second

// Handshake is the handshake config for plugins.
var Handshake = goplugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "VERSIONATOR_PLUGIN",
	MagicCookieValue: "v1",
}

// PluginTypes for discovery
const (
	PluginTypeEmit  = "emit"
	PluginTypeBuild = "build"
	PluginTypePatch = "patch"
)

// Plugin binary naming prefixes for external plugin discovery
const (
	PluginPrefixEmit  = "versionator-plugin-emit-"
	PluginPrefixBuild = "versionator-plugin-build-"
	PluginPrefixPatch = "versionator-plugin-patch-"
)

// --- Emit Plugin ---

// EmitPluginInterface is what emit plugins implement
type EmitPluginInterface interface {
	Name() string
	Format() string
	FileExtension() string
	DefaultOutput() string
	Emit(vars map[string]string) (string, error)
}

// EmitGRPCPlugin implements go-plugin's GRPCPlugin for emit plugins
type EmitGRPCPlugin struct {
	goplugin.Plugin
	Impl EmitPluginInterface
}

func (p *EmitGRPCPlugin) GRPCServer(broker *goplugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterEmitPluginServer(s, &EmitGRPCServer{Impl: p.Impl})
	return nil
}

func (p *EmitGRPCPlugin) GRPCClient(ctx context.Context, broker *goplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &EmitGRPCClient{client: proto.NewEmitPluginClient(c)}, nil
}

// EmitGRPCServer is the server-side implementation
type EmitGRPCServer struct {
	proto.UnimplementedEmitPluginServer
	Impl EmitPluginInterface
}

func (s *EmitGRPCServer) GetInfo(ctx context.Context, req *proto.Empty) (*proto.EmitInfo, error) {
	return &proto.EmitInfo{
		Name:          s.Impl.Name(),
		Format:        s.Impl.Format(),
		FileExtension: s.Impl.FileExtension(),
		DefaultOutput: s.Impl.DefaultOutput(),
	}, nil
}

func (s *EmitGRPCServer) Emit(ctx context.Context, req *proto.EmitRequest) (*proto.EmitResponse, error) {
	content, err := s.Impl.Emit(req.Vars)
	if err != nil {
		return &proto.EmitResponse{Error: err.Error()}, nil
	}
	return &proto.EmitResponse{Content: content}, nil
}

// EmitGRPCClient is the client-side implementation with cached info.
type EmitGRPCClient struct {
	client   proto.EmitPluginClient
	info     *proto.EmitInfo
	infoOnce sync.Once
	infoErr  error
}

func (c *EmitGRPCClient) fetchInfo() {
	c.infoOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
		defer cancel()
		c.info, c.infoErr = c.client.GetInfo(ctx, &proto.Empty{})
	})
}

func (c *EmitGRPCClient) Name() string {
	c.fetchInfo()
	if c.infoErr != nil || c.info == nil {
		return ""
	}
	return c.info.Name
}

func (c *EmitGRPCClient) Format() string {
	c.fetchInfo()
	if c.infoErr != nil || c.info == nil {
		return ""
	}
	return c.info.Format
}

func (c *EmitGRPCClient) FileExtension() string {
	c.fetchInfo()
	if c.infoErr != nil || c.info == nil {
		return ""
	}
	return c.info.FileExtension
}

func (c *EmitGRPCClient) DefaultOutput() string {
	c.fetchInfo()
	if c.infoErr != nil || c.info == nil {
		return ""
	}
	return c.info.DefaultOutput
}

func (c *EmitGRPCClient) Emit(vars map[string]string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
	defer cancel()
	resp, err := c.client.Emit(ctx, &proto.EmitRequest{Vars: vars})
	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}
	return resp.Content, nil
}

// --- Build Plugin ---

// BuildPluginInterface is what build plugins implement
type BuildPluginInterface interface {
	Name() string
	Format() string
	GenerateFlags(vars map[string]string) (string, error)
}

// BuildGRPCPlugin implements go-plugin's GRPCPlugin for build plugins
type BuildGRPCPlugin struct {
	goplugin.Plugin
	Impl BuildPluginInterface
}

func (p *BuildGRPCPlugin) GRPCServer(broker *goplugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterBuildPluginServer(s, &BuildGRPCServer{Impl: p.Impl})
	return nil
}

func (p *BuildGRPCPlugin) GRPCClient(ctx context.Context, broker *goplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &BuildGRPCClient{client: proto.NewBuildPluginClient(c)}, nil
}

// BuildGRPCServer is the server-side implementation
type BuildGRPCServer struct {
	proto.UnimplementedBuildPluginServer
	Impl BuildPluginInterface
}

func (s *BuildGRPCServer) GetInfo(ctx context.Context, req *proto.Empty) (*proto.BuildInfo, error) {
	return &proto.BuildInfo{
		Name:   s.Impl.Name(),
		Format: s.Impl.Format(),
	}, nil
}

func (s *BuildGRPCServer) GenerateFlags(ctx context.Context, req *proto.BuildRequest) (*proto.BuildResponse, error) {
	flags, err := s.Impl.GenerateFlags(req.Vars)
	if err != nil {
		return &proto.BuildResponse{Error: err.Error()}, nil
	}
	return &proto.BuildResponse{Flags: flags}, nil
}

// BuildGRPCClient is the client-side implementation with cached info.
type BuildGRPCClient struct {
	client   proto.BuildPluginClient
	info     *proto.BuildInfo
	infoOnce sync.Once
	infoErr  error
}

func (c *BuildGRPCClient) fetchInfo() {
	c.infoOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
		defer cancel()
		c.info, c.infoErr = c.client.GetInfo(ctx, &proto.Empty{})
	})
}

func (c *BuildGRPCClient) Name() string {
	c.fetchInfo()
	if c.infoErr != nil || c.info == nil {
		return ""
	}
	return c.info.Name
}

func (c *BuildGRPCClient) Format() string {
	c.fetchInfo()
	if c.infoErr != nil || c.info == nil {
		return ""
	}
	return c.info.Format
}

func (c *BuildGRPCClient) GenerateFlags(vars map[string]string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
	defer cancel()
	resp, err := c.client.GenerateFlags(ctx, &proto.BuildRequest{Vars: vars})
	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}
	return resp.Flags, nil
}

// --- Patch Plugin ---

// PatchPluginInterface is what patch plugins implement
type PatchPluginInterface interface {
	Name() string
	FilePattern() string
	Description() string
	Patch(content, version string) (string, error)
}

// PatchGRPCPlugin implements go-plugin's GRPCPlugin for patch plugins
type PatchGRPCPlugin struct {
	goplugin.Plugin
	Impl PatchPluginInterface
}

func (p *PatchGRPCPlugin) GRPCServer(broker *goplugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterPatchPluginServer(s, &PatchGRPCServer{Impl: p.Impl})
	return nil
}

func (p *PatchGRPCPlugin) GRPCClient(ctx context.Context, broker *goplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &PatchGRPCClient{client: proto.NewPatchPluginClient(c)}, nil
}

// PatchGRPCServer is the server-side implementation
type PatchGRPCServer struct {
	proto.UnimplementedPatchPluginServer
	Impl PatchPluginInterface
}

func (s *PatchGRPCServer) GetInfo(ctx context.Context, req *proto.Empty) (*proto.PatchInfo, error) {
	return &proto.PatchInfo{
		Name:        s.Impl.Name(),
		FilePattern: s.Impl.FilePattern(),
		Description: s.Impl.Description(),
	}, nil
}

func (s *PatchGRPCServer) Patch(ctx context.Context, req *proto.PatchRequest) (*proto.PatchResponse, error) {
	content, err := s.Impl.Patch(req.Content, req.Version)
	if err != nil {
		return &proto.PatchResponse{Error: err.Error()}, nil
	}
	return &proto.PatchResponse{Content: content}, nil
}

// PatchGRPCClient is the client-side implementation with cached info.
type PatchGRPCClient struct {
	client   proto.PatchPluginClient
	info     *proto.PatchInfo
	infoOnce sync.Once
	infoErr  error
}

func (c *PatchGRPCClient) fetchInfo() {
	c.infoOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
		defer cancel()
		c.info, c.infoErr = c.client.GetInfo(ctx, &proto.Empty{})
	})
}

func (c *PatchGRPCClient) Name() string {
	c.fetchInfo()
	if c.infoErr != nil || c.info == nil {
		return ""
	}
	return c.info.Name
}

func (c *PatchGRPCClient) FilePattern() string {
	c.fetchInfo()
	if c.infoErr != nil || c.info == nil {
		return ""
	}
	return c.info.FilePattern
}

func (c *PatchGRPCClient) Description() string {
	c.fetchInfo()
	if c.infoErr != nil || c.info == nil {
		return ""
	}
	return c.info.Description
}

func (c *PatchGRPCClient) Patch(content, version string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
	defer cancel()
	resp, err := c.client.Patch(ctx, &proto.PatchRequest{
		Content: content,
		Version: version,
	})
	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}
	return resp.Content, nil
}

// --- Plugin Maps for go-plugin ---

// EmitPluginMap for emit plugin discovery
var EmitPluginMap = map[string]goplugin.Plugin{
	PluginTypeEmit: &EmitGRPCPlugin{},
}

// BuildPluginMap for build plugin discovery
var BuildPluginMap = map[string]goplugin.Plugin{
	PluginTypeBuild: &BuildGRPCPlugin{},
}

// PatchPluginMap for patch plugin discovery
var PatchPluginMap = map[string]goplugin.Plugin{
	PluginTypePatch: &PatchGRPCPlugin{},
}
