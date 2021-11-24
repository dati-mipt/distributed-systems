package hw1

import (
    "context"
	"time"
	//"fmt"

	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
)

type FaultTolerantRegister struct {
	rid      int64
	quorum_w int64
	quorum_r int64
	register util.TimestampedValue
	replicas map[int64]network.Link
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	return &FaultTolerantRegister{
		rid:      rid,
		quorum_w: int64(0),
		quorum_r: int64(0),
		register: util.TimestampedValue{0, util.Timestamp{0, rid}},
		replicas: map[int64]network.Link{},
	}
}

//tradeoff----------------------------------------------------------------------
func GetWriteQuorumNumber(links map[int64]network.Link) int64 {
	return int64(len(links) / 2 + 1);
}

func GetReadQuorumNumber(links map[int64]network.Link) int64 {
	return int64(len(links) / 2 + 1);
}
//------------------------------------------------------------------------------

// forwarding input from different links to one channel
func StoreTimeStamps(ctx context.Context, outChan chan util.TimestampedValue,
	                 inChan <-chan interface{}) {
    select {
	case msg := <- inChan:
		if time, ok := (msg).(util.TimestampedValue); ok {
			outChan <- time
		}
	case <-ctx.Done():
		// silent mode)
		//fmt.Println("Died connection")
	}
}

func ProcessTimeStamps(ctx context.Context, inputTime chan util.TimestampedValue,
                       r *FaultTolerantRegister, ret chan int64) {
	i := int64(0)
	for i < r.quorum_r {
		select {
		case stamp := <- inputTime:
			if r.register.Ts.Less(stamp.Ts) {
		        r.register.Ts.Number = stamp.Ts.Number
				r.register.Val = stamp.Val
		    }
		    i++
		case <-ctx.Done():
			//silent mode)
			//fmt.Println("Died connection")
			break
		}
	}

	ret <- i
}

// collect acks and count them
func StoreAcks(ctx context.Context, outChan chan bool, inChan <-chan interface{}) {
	select {
	case msg := <- inChan:
		if ack, ok := (msg).(bool); ok {
			outChan <- ack
		}
	case <-ctx.Done():
		// silent mode)
		//fmt.Println("Died connection")
	}
}

func ProcessAcks(ctx context.Context, inputAcks chan bool,
                 r *FaultTolerantRegister, ret chan int64) {
	i := int64(0)
	for i < r.quorum_w {
		select {
		case <- inputAcks:
			i++
		case <-ctx.Done():
			//silent mode)
			//fmt.Println("Died connection")
			break
		}
	}

	ret <- i
}

func (r *FaultTolerantRegister) Write(value int64) bool {
    ctx, cancelFunc := context.WithTimeout(context.Background(),
	                                   time.Duration(100)*time.Millisecond)
	defer func() {
		cancelFunc()
	} ()

	// prepare Lamport clocks
	inputTime := make(chan util.TimestampedValue)
	for _, replica := range r.replicas {
		recv := replica.Send(context.Background(), struct{}{})
		go StoreTimeStamps(ctx, inputTime, recv)
	}

    retStamps := make(chan int64)
	go ProcessTimeStamps(ctx, inputTime, r, retStamps)
	retVal := <- retStamps
	if retVal < r.quorum_r {
		return false
	}

	r.register.Ts.Number++
    r.register.Val = value

    // send new value
	inputAcks := make(chan bool)
	for _, replica := range r.replicas {
		recv := replica.Send(context.Background(), r.register)
		go StoreAcks(ctx, inputAcks, recv)
	}

	retAcks := make(chan int64)
	go ProcessAcks(ctx, inputAcks, r, retAcks)
	retVal = <- retAcks
	if retVal < r.quorum_w {
		return false
	}

	return true
}

func (r *FaultTolerantRegister) Read() int64 {
	ctx, cancelFunc := context.WithTimeout(context.Background(),
	                                   time.Duration(150)*time.Millisecond)
	defer func() {
		cancelFunc()
	} ()

	// get value from quorum
	inputVal := make(chan util.TimestampedValue)
	for _, replica := range r.replicas {
		recv := replica.Send(context.Background(), struct{}{})
		go StoreTimeStamps(ctx, inputVal, recv)
	}

	retVals := make(chan int64)
	go ProcessTimeStamps(ctx, inputVal, r, retVals)
    <- retVals
	//numVals := <- retVals
	// if numVals < r.quorum_r {
	//   fmt.Println("Dead quorum")
	//}


    // resend value to peers with old clocks and whose values we don't get
	// so we need resend value only to write quorum (write + read > num peers)
	inputAcks := make(chan bool)
	for _, replica := range r.replicas {
		recv := replica.Send(context.Background(), r.register)
		go StoreAcks(ctx, inputAcks, recv)
	}

	retAcks := make(chan int64)
	go ProcessAcks(ctx, inputAcks, r, retAcks)
	<- retAcks
	//retVal = <- retAcks
	//if retVal < r.quorum_w {
	//	fmt.Println("Dead quorum")
	//}

	return r.register.Val
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.replicas[rid] = link
		r.quorum_w = GetWriteQuorumNumber(r.replicas)
		r.quorum_r = GetReadQuorumNumber(r.replicas)
	}
	//fmt.Println(len(r.replicas), r.quorum_w, r.quorum_r)
	//link.Send(context.Background(), r.register)
}

func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	switch value := msg.(type) {
	case util.TimestampedValue:
		if r.register.Ts.Less(value.Ts) {
			r.register.Ts.Number = value.Ts.Number
			r.register.Val = value.Val
		}
		return true
	case struct{}:
		return r.register
	}

	return nil
}
