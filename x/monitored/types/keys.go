package types

const (
	// ModuleName defines the module name
	ModuleName = "monitored"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_monitored"

	// Version defines the current version the IBC module supports
	Version = "1"

	// PortID is the default port id that module binds to
	PortID = "monitored"

	UpdateInterval = 10

	TimeoutInterval = 20
)

var (
	// PortKey defines the key to store the port ID in store
	PortKey = KeyPrefix("monitored-port-")

	// RegistryChainChannelIDKey defines the key to store the channel ID
	// that is used to send healthcheck updates to registry chain
	RegistryChainChannelIDKey = KeyPrefix("RegistryChainChannelID")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
