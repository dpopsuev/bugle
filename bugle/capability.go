package bugle

// ProtocolVersion is the current Bugle Protocol version.
const ProtocolVersion = "bugle/v1"

// Capabilities declares which protocol layers a server supports.
// Returned on start response. Clients adapt behavior based on declared capabilities.
type Capabilities struct {
	Protocol string       `json:"protocol"`
	Layers   LayerSupport `json:"layers"`
}

// LayerSupport indicates which optional layers the server implements.
type LayerSupport struct {
	Health bool `json:"health"`
	Budget bool `json:"budget"`
	HITL   bool `json:"hitl"`
	Status bool `json:"status"`
}

// DefaultCapabilities returns Core-only capabilities (Level 0).
func DefaultCapabilities() Capabilities {
	return Capabilities{Protocol: ProtocolVersion}
}

// FullCapabilities returns all layers enabled.
func FullCapabilities() Capabilities {
	return Capabilities{
		Protocol: ProtocolVersion,
		Layers: LayerSupport{
			Health: true,
			Budget: true,
			HITL:   true,
			Status: true,
		},
	}
}
