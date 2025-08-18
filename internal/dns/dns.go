package dns

import "swift-dns/internal/config"

type Manager interface {
	GetActive(providers []config.Provider) (string, error)
	Set(provider config.Provider) error
	Clear() error
	Flush() error
}
