package snapshot

import (
	"fmt"
	"sync"

	"github.com/kubernetes-incubator/external-storage/snapshot/pkg/client"
	snapshotvolume "github.com/kubernetes-incubator/external-storage/snapshot/pkg/volume"
	"github.com/libopenstorage/stork/drivers/volume"
	"github.com/libopenstorage/stork/pkg/snapshot/controllers"
	"github.com/portworx/sched-ops/k8s/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/controller"
)

const (
	snapshotProvisionerName = "stork-snapshot"
	snapshotProvisionerID   = "stork"
)

// Snapshot snapshot
type Snapshot struct {
	lock                       sync.Mutex
	stopChannel                chan struct{}
	started                    bool
	snapshotController         *controllers.Snapshotter
	snapshotScheduleController *controllers.SnapshotScheduleController
	snapshotRestoreController  *controllers.SnapshotRestoreController
	provisioner                *controller.ProvisionController
	Driver                     volume.Driver
	Recorder                   record.EventRecorder
}

// GetProvisionerName Gets the name of the provisioner
func GetProvisionerName() string {
	return snapshotProvisionerName
}

// Start initialize the snapshot components
func (s *Snapshot) Start(mgr manager.Manager) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.started {
		return fmt.Errorf("Snapshot controllers have already been started")
	}
	s.stopChannel = make(chan struct{})

	// Start the snapshot controller first so that the CRD gets registered
	s.snapshotController = &controllers.Snapshotter{
		Driver: s.Driver,
	}
	err := s.snapshotController.Start(s.stopChannel)
	if err != nil {
		return fmt.Errorf("error starting snapshot controller: %v", err)
	}

	if err := performRuleRecovery(); err != nil {
		log.Errorf("Failed to perform recovery for snapshot rules due to: %v", err)
		return err
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	if clientset == nil {
		return errors.ErrK8SApiAccountNotSet
	}

	// Start the provisioner controller
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("error getting server version: %v", err)
	}

	snapshotClient, _, err := client.NewClient(config)
	if err != nil {
		return err
	}

	plugins := make(map[string]snapshotvolume.Plugin)
	plugins[s.Driver.String()] = s.Driver.GetSnapshotPlugin()

	snapProvisioner := controllers.NewSnapshotProvisioner(clientset, snapshotClient, plugins, snapshotProvisionerID)

	s.provisioner = controller.NewProvisionController(
		clientset,
		snapshotProvisionerName,
		snapProvisioner,
		serverVersion.GitVersion,
	)
	// stork already does leader elction, don't need it for each controller
	if err := controller.LeaderElection(false)(s.provisioner); err != nil {
		log.Errorf("failed to disable leader election for snapshot controller: %v", err)
		return err
	}

	go s.provisioner.Run(s.stopChannel)

	// Start the snapshot schedule controller
	s.snapshotScheduleController = controllers.NewSnapshotScheduleController(mgr, s.Recorder)
	err = s.snapshotScheduleController.Init(mgr)
	if err != nil {
		return fmt.Errorf("error initializing snapshot schedule controller: %v", err)
	}

	s.snapshotRestoreController = controllers.NewSnapshotRestoreController(mgr, s.Driver, s.Recorder)
	err = s.snapshotRestoreController.Init(mgr)
	if err != nil {
		return fmt.Errorf("error initializing snapshot restore controller: %v", err)
	}
	s.started = true
	return nil
}

// Stop Stops the snapshotter controllers
func (s *Snapshot) Stop() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.started {
		return fmt.Errorf("Snapshot controllers have not been started")
	}

	close(s.stopChannel)

	s.started = false
	return nil
}
