package static

import (
	"context"
	"fmt"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

const (
	renewDeadline = 10 * time.Second
	leaseDuration = 15 * time.Second
	retryPeriod   = 2 * time.Second
)

// leaderElectorConfig holds all the configuration for the leaderElector struct.
type leaderElectorConfig struct {
	// kubeConfig is the kube config for the cluster. Used to create coreV1 and coordinationV1 clients which are needed
	// for leader election.
	kubeConfig *rest.Config
	// recorder is the Kubernetes event recorder. Used to record events on the lease lock.
	recorder record.EventRecorder
	// onStartedLeading is the callback that is invoked asynchronously when the Pod starts leading.
	onStartedLeading func(ctx context.Context)
	// onStoppedLeading is the callback that is invoked asynchronously when the Pod stops leading.
	onStoppedLeading func()
	// lockNs is the namespace where the LeaseLock resource lives.
	lockNs string
	// lockName is the name of the LeaseLock resource.
	lockName string
	// identity is the unique name of this Pod. Used to identify the leader.
	identity string
}

// newLeaderElector returns a new leader elector client.
func newLeaderElector(config leaderElectorConfig) (*leaderelection.LeaderElector, error) {
	lock, err := resourcelock.NewFromKubeconfig(
		resourcelock.LeasesResourceLock,
		config.lockNs,
		config.lockName,
		resourcelock.ResourceLockConfig{
			Identity:      config.identity,
			EventRecorder: config.recorder,
		},
		config.kubeConfig,
		renewDeadline,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating lease lock for leader election: %w", err)
	}

	leaderElector, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: leaseDuration,
		RenewDeadline: renewDeadline,
		RetryPeriod:   retryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: config.onStartedLeading,
			OnStoppedLeading: config.onStoppedLeading,
		},
		Name: lock.Describe(),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating leader elector: %w", err)
	}

	return leaderElector, nil
}
