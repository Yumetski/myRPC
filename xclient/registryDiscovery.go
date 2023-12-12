package xclient

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type MyRegistryDiscovery struct {
	*MultiServersDiscovery
	registry   string
	timeout    time.Duration
	lastUpdate time.Time
}

const defaultUpdateTimeout = time.Second * 10

func NewMyRegistryDiscovery(registerAddr string, timeout time.Duration) *MyRegistryDiscovery {
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}
	d := &MyRegistryDiscovery{
		MultiServersDiscovery: NewMultiServerDiscovery(make([]string, 0)),
		registry:              registerAddr,
		timeout:               timeout,
	}
	return d
}

func (m *MyRegistryDiscovery) Update(servers []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.servers = servers
	m.lastUpdate = time.Now()
	return nil
}

func (m *MyRegistryDiscovery) Refresh() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.lastUpdate.Add(m.timeout).After(time.Now()) {
		return nil
	}
	log.Println("rpc registry: refresh servers from registry", m.registry)
	resp, err := http.Get(m.registry)
	if err != nil {
		log.Println("rpc registry refresh err:", err)
		return err
	}
	servers := strings.Split(resp.Header.Get("X-Myrpc-Servers"), ",")
	m.servers = make([]string, 0, len(servers))
	for _, server := range servers {
		if strings.TrimSpace(server) != "" {
			m.servers = append(m.servers, strings.TrimSpace(server))
		}
	}
	m.lastUpdate = time.Now()
	return nil
}

func (m *MyRegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := m.Refresh(); err != nil {
		return "", err
	}
	return m.MultiServersDiscovery.Get(mode)
}

func (m *MyRegistryDiscovery) GetAll() ([]string, error) {
	if err := m.Refresh(); err != nil {
		return nil, err
	}
	return m.MultiServersDiscovery.GetAll()
}
