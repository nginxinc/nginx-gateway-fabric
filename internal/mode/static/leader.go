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
	// These values are the defaults used by the core client.
	renewDeadline = 10 * time.Second
	leaseDuration = 15 * time.Second
	retryPeriod   = 2 * time.Second
)

// leaderElectorRunnableConfig holds all the configuration for the leaderElector struct.
type leaderElectorRunnableConfig struct {
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

// leaderElectorRunnable wraps a leaderelection.LeaderElector so that it implements the manager.Runnable interface
// and can be managed by the manager.
type leaderElectorRunnable struct {
	le *leaderelection.LeaderElector
}

// Start runs the leaderelection.LeaderElector and blocks until the context is canceled or Run returns.
func (l *leaderElectorRunnable) Start(ctx context.Context) error {
	l.le.Run(ctx)
	return nil
}

// IsLeader returns if the Pod is the current leader.
func (l *leaderElectorRunnable) IsLeader() bool {
	return l.le.IsLeader()
}

// newLeaderElector returns a new leader elector client.
func newLeaderElectorRunnable(config leaderElectorRunnableConfig) (*leaderElectorRunnable, error) {
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

	return &leaderElectorRunnable{le: leaderElector}, nil
}
