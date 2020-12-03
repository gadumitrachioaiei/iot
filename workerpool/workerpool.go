package workerpool

type Worker func()

type Pool struct {
	semaphor chan struct{}
	tasks    chan Worker
}

func New(size int) *Pool {
	p := &Pool{tasks: make(chan Worker), semaphor: make(chan struct{}, size)}
	go p.Run()
	return p
}

// DoWork returns after the work has been scheduled to run.
// The work will run in its own goroutine
func (p *Pool) DoWork(w Worker) {
	p.tasks <- w
}

// Run runs as many goroutines concurrently as the capacity of the semaphor
func (p *Pool) Run() {
	for {
		p.semaphor <- struct{}{}
		w := <-p.tasks
		go func() {
			w()
			<-p.semaphor
		}()
	}
}
