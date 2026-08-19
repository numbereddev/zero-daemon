package container

// TODO: i have to figure out the best way to do this, should the json even be given here already?
// i also have to worry about race conditions so i may need to add a mutx... idk i am too tired,
// this is problem for future me.

type Config struct {
	Image      string      `json:"image"`
	Resources  Resources   `json:"resources"`
	Networking Networking  `json:"networking"`
	EnvVars    EnvVars     `json:"env_vars"`
	StopCmd    StopCommand `json:"stop_command"`
}

type Resources struct {
	CPU         uint32 `json:"cpu"`
	MemoryBytes uint64 `json:"memory_bytes"`
	SwapBytes   uint64 `json:"swap_bytes"`
	DiskBytes   uint64 `json:"disk_bytes"`
	OOMKiller   bool   `json:"oom_killer"`
	IOWeight    uint16 `json:"io_weight"`
}

type Networking struct {
	IsDedicatedIP bool `json:"is_dedicated_ip"`
	Default       struct {
		IP   string `json:"ip"`
		Port uint16 `json:"port"`
	} `json:"default"`
	Allocations map[string][]uint16 `json:"allocations"`
}

type EnvVars map[string]string

type StopCommand struct{}
