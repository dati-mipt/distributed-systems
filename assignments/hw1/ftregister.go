package ftregister

import (
	"fmt"
	"github.com/dati-mipt/distributed-storage-algorithms/network"
	"github.com/dati-mipt/distributed-storage-algorithms/util"
)

type FaultTolerantRegister struct {
	rid      int64
	current  util.TimestampedValue
	replicas map[int64]network.Link
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	return &FaultTolerantRegister{
		rid:      rid,
		replicas: map[int64]network.Link{},
	}
}

func (r *FaultTolerantRegister) Write(value int64) bool {
	fmt.Println("-----------Write-----------")
	fmt.Printf("write value: %d\n", value)
	
	r.current.Val = value
	r.current.Ts = util.Timestamp{Number: r.current.Ts.Number + 1, Rid: r.rid}


	return true
}

func Callback(rep network.Link, pipe chan util.TimestampedValue) {
	var msg = (rep.BlockingMessage(nil)).(util.TimestampedValue)
	fmt.Println("waiting ...")
	pipe <- msg
}

func BlockingMessageToQuorum(r *FaultTolerantRegister)(msg util.TimestampedValue) {
	// create channel
	message_chan := make(chan util.TimestampedValue, 2)

	// create go routines
	var counter = 0
	for _, rep := range r.replicas {
		go func(pipe chan util.TimestampedValue, rep network.Link) {
			
			var msg = (rep.BlockingMessage(nil)).(util.TimestampedValue)
			
			pipe <- msg
		}(message_chan, rep)
		counter++
	}
	close(message_chan)

	
	fmt.Printf("counter: %d\n", counter)
	for i := 0; i < cap(message_chan); i++ {
		y := <-message_chan
		fmt.Println(y)
	}

	// organize quorum
	var max_replica_resp = (len(r.replicas) / 2) + 1
	var rid_ = 0
	var max_ts util.TimestampedValue
	for elem := range message_chan {
		rid_++
		if elem.Ts.Less(max_ts.Ts) {
			max_ts = elem
		}
	}
	// 
	if(rid_ <= max_replica_resp){
		return util.TimestampedValue{}
	}

	return max_ts
}

// compound read = {read + write}

func (r *FaultTolerantRegister) Read() int64 {
	fmt.Println("------------Read-----------")
	fmt.Printf("rid: %d, ts: %d, read value: %d\n", r.current.Ts.Number, r.rid, r.current.Val)

	BlockingMessageToQuorum(r)

	// read data from replica and set current TimestampedValue due to max replica id

	/*var msg interface{}
	if update, ok := msg.(util.TimestampedValue); ok {
		if r.current.Ts.Less(update.Ts) {
			r.current = update
		}
	}

	// update Timestamp

	r.current.Ts = util.Timestamp{Number: r.current.Ts.Number + 1, Rid: r.rid}

	// write value to replicas

	for _, rep := range r.replicas {
		rep.AsyncMessage(r.current)
	}*/

	return r.current.Val
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.replicas[rid] = link
	}

}

// make a copy of data
func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	if update, ok := msg.(util.TimestampedValue); ok {
		if r.current.Ts.Less(update.Ts) {
			r.current = update
		}
	}
	return r.current
}

func (r *FaultTolerantRegister) Update() {
	for _, rep := range r.replicas {
		rep.AsyncMessage(r.current)
	}
}
