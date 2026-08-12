package models

import "gorm.io/gorm"

// type deploymentType string
//
// var (
// 	// DeploymentTypeApp ; Pterodactyl-style servers
// 	DeploymentTypeApp deploymentType
// 	// DeploymentTypeDocker ; Docker gVisor deployments
// 	DeploymentTypeDocker deploymentType
// )

type Service struct {
	gorm.Model
	ID string `gorm:"primaryKey;<-:create"`

	HasGitRepo bool

	// Resources
	// CPU
	CPU       uint
	CPUBurst  uint
	CPUShares uint16
	// Mmeory
	MemoryBytes      uint
	MemoryBurstBytes uint
	// Disk
	DiskBytes uint
	// Misc
	Swap        bool
	OOMKiller   bool
	Allocations uint32

	// // Deployment information
	// Deployment
}

// // TODO: probably extract these below out to some better spot, idk
//
// type Deployment struct {
// 	Type        deploymentType
// 	EnvVars     map[string]string
// 	Allocations []Allocation `gorm:"serializer:json"`
//
// 	// App
// 	AppConfiguration *AppConfiguration `gorm:"serializer:json"`
//
// 	// Docker
// 	DockerDockerfile *string
// }
//
// type Allocation struct {
// 	IP           string
// 	InternalPort uint16
// 	ExternalPort uint16
// 	// Pinned means the entire IP has been assigned to the service
// 	Pinned bool
// }
//
// type AppConfiguration struct{}
