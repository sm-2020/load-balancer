// A very basic load balancer implementation.
// It provides round-robin load balancing and sends hearbeats messages
// to backend endpoints in order to detect unreacheable hosts

package main

import (
    "context"
    "flag"
    "fmt"
    "sync"
    "sync/atomic"
    "log"
    "strings"
    "time"
    "net"
    "net/http"
    "net/http/httputil"
    "net/url"
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
    return int(atomic.AddUint64(&s.current,uint64(1)) % uint64(len(s.backends)))
}

//Set whether this backend endpoint is alive or not
func (b  *Backend) SetAlive(isAlive bool) {
    b.mux.Lock()
    b.Alive = isAlive
    b.mux.Unlock()
}
//ISAlive returns true when any backend is alive
func (b *Backend) IsAlive() (alive bool) {
    b.mux.RLock()
    alive = b.Alive
    b.mux.RUnlock()
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
    l := len(s.backends) + next
    for i := next; i < l; i++ {
        idx := i % len(s.backends) // use modding to keep index within range
        if s.backends[idx].IsAlive() {
            if i != next {
                atomic.StoreUint64(&s.current,uint64(idx))
            }
            return s.backends[idx]
        }
    }
    return  nil
}

// Checks to see if a particular backend is alive by pining it
func isBackendAlive(url *url.URL) bool {
    timeout := 2 * time.Second
    conn, err := net.DialTimeout("tcp",url.Host,timeout)
    if err != nil {
        log.Println("Site unreachable, error: ", err)
        return false
    }
    _ = conn.Close()
    return true
}
//Pings every backend endpoints int eh slice to check their status
func (s *ServerPool) HealthCheck() {
    for  _, b := range s.backends {
        status := "up"
        alive := isBackendAlive(b.Url)
        b.SetAlive(alive)
        if !alive {
            status = "down"
        }
        log.Printf("%s [%s]\n", b.Url, status)
    }
}
// Get the number of attepts from the request header
// context package allows you to store useful data in an Http request.
// therefore heavily utilize this to track request specific data such as Attempt count and Retry count.
func GetAttemptsfromRequest(req  *http.Request) int {
    if attempts, ok := req.Context().Value(Attempts).(int); ok {
        return attempts
    }
    return 1
}

//Get the number of failures from the request
func GetRetriesfromRequest(req *http.Request) int {
    if retry,ok := req.Context().Value(Retry).(int); ok {
        return retry
    }
    return 0
}

func runHealthCheck() {
    t := time.NewTicker(time.Minute * 2)
    for {
        select {

         case <- t.C:
            log.Println("Starting health check...")
            serverpool.HealthCheck()
            log.Println("Health check completed")
        }
    }
}

//load balance incoming requests in a round robin manner
func loadBalance(w http.ResponseWriter,req *http.Request) {
    attempts := GetAttemptsfromRequest(req)
    if attempts > 3 {
        log.Printf("%s(%s) Max attempts reached, terminating\n", req.RemoteAddr, req.URL.Path)
        http.Error(w, "Service not available.", http.StatusServiceUnavailable)
        return
    }
    nextService := serverpool.GetNextActivePeer ()
    if nextService  != nil {
        nextService.ReverseProxy.ServeHTTP(w,req)
        return
    }
    http.Error(w,"Service not available.",http.StatusServiceUnavailable)

}
var serverpool ServerPool

func main() {
    //parse args and create ServerPool
    var serverList string
    var port int
    flag.StringVar(&serverList,"backends","", "Load balanced backends, use commas to separate")
    flag.IntVar(&port,"port",3030,"Port to server")
    flag.Parse()

    if len(serverList) == 0 {
        log.Fatal("Please provide one or more server backends to load balance")
    }
    //Now parse the backends
    tokens := strings.Split(serverList,",")

    for _, tok := range tokens {
        serverUrl, err := url.Parse(tok)
        if err != nil {
            log.Fatal(err)
        }
        proxy := httputil.NewSingleHostReverseProxy(serverUrl)
        proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, e error) {
           log.Printf("[%s] %s\n", serverUrl.Host, e.Error())
           retries := GetRetriesfromRequest(req)
           if retries < 3 {
               select {
               case <- time.After(10 * time.Millisecond):
                //increment retries and add it to context
                ctx := context.WithValue(req.Context(),Retry, retries+1)
                proxy.ServeHTTP(w, req.WithContext(ctx))
              }
              return
           }
           //Consider the endpoint to be down after 3 retries
          serverpool.MarkBackendStatus(serverUrl,false)
          attempts :=  GetAttemptsfromRequest(req)
          log.Printf("%s(%s) Attempting another retry. %d\n", req.RemoteAddr, req.URL.Path,attempts)
          ctx := context.WithValue(req.Context(),Attempts,attempts+1)
          loadBalance(w,req.WithContext(ctx))
        }
               serverpool.AddBackend(&Backend {
                   Url: serverUrl,
                   Alive: true,
                   ReverseProxy: proxy,
                })
                log.Printf("Configured endpoint: %s\n", serverUrl)
    }
    //create http server
    server := http.Server {
       Addr:   fmt.Sprintf(":%d", port),
       Handler: http.HandlerFunc(loadBalance),
    }
    go runHealthCheck()

    //Print start message
    log.Printf("Simple load balancer started at :%d\n", port)
    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
