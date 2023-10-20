package storage

// dummy in-memroy store implementation
type InMemStore interface {
	Store(key string, endpoint string)
	Get() map[string][]string
	GeEndpointstFor(key string) ([]string, bool)
}

type inMemApiDiscovery struct {
	discoveredAPIs map[string][]string // structure to store discovered api paths
}

func NewInMemStore() InMemStore {
	return &inMemApiDiscovery{
		discoveredAPIs: make(map[string][]string),
	}
}

func (s *inMemApiDiscovery) Store(key string, endpoint string) {
	api, ok := s.discoveredAPIs[key]
	if !ok {
		s.discoveredAPIs[key] = []string{endpoint}
	}
	s.discoveredAPIs[key] = append(api, endpoint)
}

func (s *inMemApiDiscovery) Get() map[string][]string {
	return s.discoveredAPIs
}

func (s *inMemApiDiscovery) GeEndpointstFor(key string) ([]string, bool) {
	endpoints, keyExists := s.discoveredAPIs[key]
	return endpoints, keyExists
}
