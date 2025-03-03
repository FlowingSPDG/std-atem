package stdatem

// ATEMHost represents a single ATEM switcher configuration
type ATEMHost struct {
	IP            string // ATEM IP address
	Name          string // Friendly name for the ATEM
	AutoReconnect bool   // Whether to automatically reconnect
}

// Config Config for setup
type Config struct {
	Params    []string   // Set os.Args
	ATEMHosts []ATEMHost // List of ATEM hosts
	Debug     bool       // debug
}
