package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type ServerItem struct {
	Addr  string
	start time.Time
}

type MyRegistry struct {
	timeout time.Duration
	mu      sync.Mutex
	servers map[string]*ServerItem
}

const (
	defaultPath    = "/_myrpc_/registry"
	defaultTimeout = time.Minute * 5
)

func New(timeout time.Duration) *MyRegistry {
	return &MyRegistry{
		servers: make(map[string]*ServerItem),
		timeout: timeout,
	}
}

var DefaultMyRegister = New(defaultTimeout)

// 添加服务实例
func (m *MyRegistry) putServer(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.servers[addr]
	if s == nil {
		m.servers[addr] = &ServerItem{
			Addr:  addr,
			start: time.Now(),
		}
	} else {
		s.start = time.Now()
	}
}

// 返回可用的服务列表，如果存在
func (m *MyRegistry) aliveServers() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var alive []string
	for addr, server := range m.servers {
		if m.timeout == 0 || server.start.Add(m.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			delete(m.servers, addr)
		}
	}
	sort.Strings(alive)
	return alive
}

func (m *MyRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Header().Set("X-Myrpc-Servers", strings.Join(m.aliveServers(), ","))
	case "POST":
		addr := req.Header.Get("X-Myrpc-Server")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		m.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (m *MyRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, m)
	log.Println("rpc registry path:", registryPath)
}

func HandleHTTP() {
	DefaultMyRegister.HandleHTTP(defaultPath)
}

func Heartbeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		duration = defaultTimeout - time.Duration(1)*time.Minute
	}
	var err error
	err = sendHeartbeat(registry, addr)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartbeat(registry, addr)
		}
	}()
}

func sendHeartbeat(registry string, addr string) error {
	log.Println(addr, "send heart beat to registry", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", registry, nil)
	req.Header.Set("X-Myrpc-Server", addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("rpc server: heart beat err:", err)
		return err
	}
	return nil
}
