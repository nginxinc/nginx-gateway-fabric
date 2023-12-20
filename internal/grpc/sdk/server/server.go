package server

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/nginxinc/nginx-gateway-fabric/internal/agent/file"
	"github.com/nginxinc/nginx-gateway-fabric/internal/grpc/controlplane"
)

type Config struct {
	Generation uint32
	Files      []file.File
}

type ControlPlane struct {
	cfgLock sync.RWMutex
	cfg     Config

	cfgUpdateChannelsLock sync.Mutex
	cfgUpdateChannels     map[uint32]chan struct{}

	successfulConfigApplyCh chan uint32
	latestSuccessfulGen     uint32
	expectedGensCh          chan expectedGeneration

	counter atomic.Uint32

	controlplane.UnimplementedControlPlaneServer
}

type expectedGeneration struct {
	gen uint32
	ch  chan struct{}
}

func NewControlPlane() *ControlPlane {
	return &ControlPlane{
		cfgUpdateChannels:       make(map[uint32]chan struct{}),
		successfulConfigApplyCh: make(chan uint32),
		expectedGensCh:          make(chan expectedGeneration),
	}
}

func (s *ControlPlane) UpdateConfig(ctx context.Context, cfg Config) error {
	if cfg.Generation == 0 {
		panic("generation must be > 0")
	}

	s.cfgLock.Lock()
	s.cfg = cfg
	expectedGen := cfg.Generation
	s.cfgLock.Unlock()

	configAppliedCh := make(chan struct{})

	select {
	case s.expectedGensCh <- expectedGeneration{
		gen: expectedGen,
		ch:  configAppliedCh,
	}:
	case <-ctx.Done():
		return nil
	}

	s.cfgUpdateChannelsLock.Lock()
	log.Printf("Notifying %d channels about config %d update", len(s.cfgUpdateChannels), expectedGen)
	for _, ch := range s.cfgUpdateChannels {
		select {
		case ch <- struct{}{}:
		default:
			// channel is full, skip
		}
	}
	s.cfgUpdateChannelsLock.Unlock()

	select {
	case <-ctx.Done():
		log.Printf("Timeout waiting for config %d apply", expectedGen)
		return fmt.Errorf("timeout waiting for config apply")
	case <-configAppliedCh:
		log.Printf("Config %d applied", expectedGen)
		return nil
	}
}

func (s *ControlPlane) Start(ctx context.Context) error {
	var expectedGen *expectedGeneration

	for {
		select {
		case expGen := <-s.expectedGensCh:
			if expectedGen != nil {
				close(expectedGen.ch)
			}
			expectedGen = &expGen
			if expGen.gen <= s.latestSuccessfulGen {
				close(expGen.ch)
				expectedGen = nil
			}
		case <-ctx.Done():
			if expectedGen != nil {
				close(expectedGen.ch)
			}
			return nil
		case gen := <-s.successfulConfigApplyCh:
			if gen > s.latestSuccessfulGen {
				s.latestSuccessfulGen = gen
			}
			if expectedGen != nil {
				if gen >= expectedGen.gen {
					close(expectedGen.ch)
					expectedGen = nil
				}
			}
		}
	}
}

func (s *ControlPlane) StreamMessages(stream controlplane.ControlPlane_StreamMessagesServer) error {
	// connection request
	msg, err := stream.Recv()
	if err != nil {
		return err
	}

	req := msg.GetConnectionRequest()
	if req == nil {
		log.Printf("Expected connection request, got %v", msg)
		return fmt.Errorf("expected connection request, got %v", msg)
	}
	log.Printf("Received connection request: %v", req)

	id := s.counter.Load()
	s.counter.Add(1)

	configs := make(chan struct{}, 1)

	s.cfgLock.Lock()
	configs <- struct{}{}
	s.cfgLock.Unlock()

	s.cfgUpdateChannelsLock.Lock()
	s.cfgUpdateChannels[id] = configs
	s.cfgUpdateChannelsLock.Unlock()
	defer func() {
		s.cfgUpdateChannelsLock.Lock()
		delete(s.cfgUpdateChannels, id)
		close(configs)
		s.cfgUpdateChannelsLock.Unlock()
	}()

	for {
		select {
		case <-configs:
			var cfg *controlplane.Configuration
			s.cfgLock.RLock()
			files := make([]*controlplane.File, 0, len(s.cfg.Files))

			for _, f := range s.cfg.Files {
				files = append(files, &controlplane.File{
					Path:    f.Path,
					Content: f.Content,
					Type:    controlplane.FileType(f.Type),
				})
			}

			cfg = &controlplane.Configuration{
				Generation: s.cfg.Generation,
				Files:      files,
			}
			s.cfgLock.RUnlock()

			err := stream.Send(&controlplane.ControlPlaneMessage{
				Message: &controlplane.ControlPlaneMessage_Configuration{
					Configuration: cfg,
				},
			})
			if err != nil {
				log.Printf("Failed to send configuration to %d: %v", id, err)
				return err
			}
			log.Printf("Sent configuration to %d: %v", id, cfg)

		case <-stream.Context().Done():
			log.Printf("Stream closed for %d", id)
			return nil
		}

		// expect config apply
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("Failed to receive configuration apply result from %d: %v", id, err)
			return err
		}

		apply := msg.GetConfigurationApplyResult()
		if apply == nil {
			log.Printf("Expected configuration apply result, got %v", msg)
			return fmt.Errorf("expected configuration apply result, got %v", msg)
		}

		log.Printf("Received configuration apply result: %v", apply)
		select {
		case s.successfulConfigApplyCh <- apply.Generation:
		case <-stream.Context().Done():
			return nil
		}
	}
}
