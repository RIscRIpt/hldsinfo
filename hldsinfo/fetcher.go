package hldsinfo

import (
	"sync"
	"time"
)

// Fetcher provides interface to async server info fetcher
// Call Close when done to get all fetched servers info.
// Calling Close will block if not all servers info were received.
// Calling Fetch multiple times on same address will provide only one latest result.
// Close must be called only if Get wasn't. Otherwise go-routine leak will occur.
type Fetcher interface {
	Fetch(address string)
	Get() map[string]*Info
	Close()
}

type fetcher struct {
	c          chan string
	cMutex     sync.Mutex
	infos      map[string]*Info
	infosMutex sync.Mutex
	timeout    time.Duration
	wg         sync.WaitGroup
}

// NewFetcher creates and returns a new server info fetcher
func NewFetcher(timeout time.Duration) Fetcher {
	f := &fetcher{
		c:       nil,
		infos:   make(map[string]*Info),
		timeout: timeout,
	}
	return f
}

// Fetch asks Fetcher to asynchronously get server info
func (f *fetcher) Fetch(address string) {
	if f.c == nil {
		f.cMutex.Lock()
		defer f.cMutex.Unlock()
		if f.c == nil {
			f.c = make(chan string)
			go f.run()
		}
	}
	f.c <- address
}

func (f *fetcher) wait() {
	if f.c != nil {
		f.cMutex.Lock()
		defer f.cMutex.Unlock()
		if f.c != nil {
			close(f.c)
			f.c = nil
		}
	}
	f.wg.Wait()
}

// Close terminates fetcher go-routine
func (f *fetcher) Get() map[string]*Info {
	f.wait()
	return f.infos
}

func (f *fetcher) Close() {
	f.wait()
}

func (f *fetcher) getInfo(address string) {
	defer f.wg.Done()
	info, _ := Get(address, time.Now().Add(f.timeout))
	f.infosMutex.Lock()
	defer f.infosMutex.Unlock()
	f.infos[address] = info
}

func (f *fetcher) run() {
	for address := range f.c {
		f.wg.Add(1)
		go f.getInfo(address)
	}
}
