// A very basic load balancer implementation.
// It provides round-robin load balancing and sends hearbeats messages
// to backend endpoints in order to detect unreacheable hosts

package main

import (
    "context"
    "flag"
    "fmt"
    "sync"
)

// Store information about the backend endpoints
type Backend struct {
    Url             *url.URL
    Alive            bool
    mux             sync.RWMutex
    ReverseProxy    *httputil.ReverseProxy
}

const (
    Attempts  int = iota
    Retry
)

// Tracks all the backend endpoints in a slice and has
// a counter variable
type ServerPool struct {
    backends    []*Backend
    current     uint64
}

// Add an backend to the server pool
func (s *ServerPool) AddBackend(backend *Backend) {
    s.backends = append(s.backends,backend)
}
//Increase the counter and returns the next available index in the ServerPool slice
func (s *ServerPool) NextIndex() int {
    return int(atomic.addUint64(&s.current,uint64(1)) % uint64(len(s.backends)))
}

//Set whether this backend endpoint is alive or not
func (b  *Backend) SetAlive(isAlive bool) {
    b.mux.Lock()
    b.Alive = isAlive
    b.mux.Unlock()
}
//ISAlive returns true when any backend is alive
func (b *Backend) isAlive() (alive bool) {
    b.mux.RLock()
    alive = b.Alive
    n.mux.RUnlock()
    return
}
//Mark backend status change of a a particular server
func (s *ServerPool) MarkBackendStatus(backendURL *url.URL, alive bool) {
    for _, b := range s.backends {
        if b.Url.String() == backendURL.String() {
            b.SetAlive(alive)
            break
        }
    }
}
//Returns the next active/isAlive endpoint to accept the next request
func (s *ServerPool) GetNextActivePeer() *Backend {
    //Look over the ServerPool to find the next active backend endpoint
    // and if isAlive then return itsi value

    next := s.NextIndex()
    //start from the next and move a full cycle
    l = len(s.backends) + next
    for i := next; i < l; i++ {
        idx := i % len(s.backends) // use modding to keep index within range
        if s.backends[idx].IsAlive() {
            if i != next {
                atomic.StoreUint64(&s.current,uint64(idx))
            }
            return backends[idx]
        }
    }
    return  nil
}

// Checks to see if a particular backend is alive by pining it
func isBackendAlive(url *url.URL) bool {
    timeout := 2 * time.Second
    conn, err := net.DialTimeout("tcp",u.Host,timeout)
    if err != nil {
        log.Println("Site unreachable, error: ", err)
        return false
    }
    _ = conn.close()
    return true
}
//Pings every backend endpoints int eh slice to check their status
func (s *ServerPool) HealthCheck() {
    for  _, b := range s.backends {
        status := "up"
        alive := isBackendAlive(b.url)
        b.SetAlive(alive)
        if !alive {
            status := "down"
        }
        log.Printf("%s [%s]\n", b.URL, status)
    }
}
// Get the number of attepts from the request header
func GetAttemptsfromRequest(req  *http.Request) int {
    if attempts, ok := req.Context().value(Attempts).i(int); ok {
        return attempts
    }
    return 1
}

//Get the number of failures from the request
func GetRetriesfromRequest(req *http.Request) int {
    if retry,ok := req.Context().value(Retry).(int); ok {
        return retry
    }
    return 0
}

func runHealthCheck() {
    t := time.NewTicker(time.Minute * 2)
    for {
        select {
        case <- t.C:
            Println("Starting health check...")
            serverPool.HealthCheck()
             log.Println("Health check completed")
        }
    }
}

//load balance incoming requests in a round robin manner
func loadBalance(w http.ResponseWriter,req *http.Request) {

}
