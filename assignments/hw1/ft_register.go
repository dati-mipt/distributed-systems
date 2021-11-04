package hw1

import (
	"math/rand"
	"fmt"
	"os"
	"time"
	"context"

	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
)

// Timeout of quorum-read/write operations (in microseconds):
const RequestTimeout = 100

// Message types:
type SetOp struct {
	val  int64
	time util.Timestamp
}

type GetOp struct{}

type SetAck struct{}

type GetAck struct {
	val  int64
	time util.Timestamp
}

// Fault-tolerant register implemetation:
type FaultTolerantRegister struct {
	value        int64
	rand_gen     *rand.Rand
	num_replicas uint16
	quorum_read  uint16
	quorum_write uint16
	lclock       util.Timestamp
	replicas     map[int64]network.Link
	rid_list     []int64
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	return &FaultTolerantRegister{
		value:        0,
		rand_gen:     rand.New(rand.NewSource(time.Now().Unix())),
		num_replicas: 0,
		quorum_read:  0,
		quorum_write: 0,
		lclock:       util.Timestamp{Number: 0, Rid: rid},
		replicas:     map[int64]network.Link{},
		rid_list:     make([]int64, 0),
	}
}

func (r *FaultTolerantRegister) write_quorum() bool {
	// Broadcast write requests to all nodes:
	ack_channels := make(map[int64]<-chan interface{}, len(r.replicas))

	for rid, _ := range r.replicas {
		msg := SetOp{val: r.value, time: r.lclock}
		ack_channels[rid] = r.replicas[rid].Send(context.Background(), msg)
	}

	// Wait for Qw responces in random order:
	var num_set_responces uint16 = 0
	indexes := r.rand_gen.Perm(len(r.rid_list))

	for _, rand_i := range indexes {
		rand_rid := r.rid_list[indexes[rand_i]]

		select {
		case res := <-ack_channels[rand_rid]:
			_, ok := res.(SetAck)
			if !ok {
				fmt.Printf("ERROR: Invalid responce to SetOp message\n")
				os.Exit(1)
			}

			num_set_responces += 1
		// Stop waiting for the acknowledgement after timeout expiration:
		case <-time.After(RequestTimeout * time.Second / 1000000):
			continue
		}

		if num_set_responces == r.quorum_write {
			break
		}
	}

	if num_set_responces < r.quorum_write {
		fmt.Printf("ERROR: Failed to write quorum\n")
		return false
	}

	return true
}

func (r *FaultTolerantRegister) read_quorum() bool {
	// Broadcast read request to all nodes:
	ack_channels := make(map[int64]<-chan interface{}, len(r.replicas))

	for rid, _ := range r.replicas {
		msg := GetOp{}
		ack_channels[rid] = r.replicas[rid].Send(context.Background(), msg)
	}

	// Wait for Qr responces in random order:
	var num_get_responces uint16 = 0
	var latest_timestamp int64 = r.lclock.Number
	var latest_value int64 = r.value

	indexes := r.rand_gen.Perm(len(r.rid_list))

	for _, rand_i := range indexes {
		rand_rid := r.rid_list[indexes[rand_i]]

		select {
		// Update the maximum timestamp:
		case res := <-ack_channels[rand_rid]:
			ack, ok := res.(GetAck)
			if !ok {
				fmt.Printf("ERROR: Invalid responce to GetOp message\n")
				os.Exit(1)
			}

			if latest_timestamp < ack.time.Number {
				latest_timestamp = ack.time.Number
				latest_value = ack.val
			}

			num_get_responces += 1
		// Stop waiting for the acknowledgement after timeout expiration:
		case <-time.After(RequestTimeout * time.Second / 1000000):
			continue
		}

		if num_get_responces == r.quorum_read {
			break
		}
	}

	if num_get_responces < r.quorum_read {
		fmt.Printf("ERROR: Failed to read quorum\n")
		return false
	}

	// Update current timestamp and value:
	r.lclock.Number = latest_timestamp
	r.value = latest_value

	return true
}

func (r *FaultTolerantRegister) Write(value int64) bool {
	// Sync value and time with read-quorum:
	var read_ok bool = r.read_quorum()

	if (!read_ok) {
		return false
	}

	// Update lamport clock:
	r.lclock.Number += 1

	// Write value to write-quorum:
	r.value = value
	var write_ok bool = r.write_quorum()

	return write_ok
}

func (r *FaultTolerantRegister) Read() int64 {
	// Read latest value from read-quorum:
	var read_ok bool = r.read_quorum()
	if (!read_ok) {
		return -1
	}

	// Write read value to write-quorum:
	var write_ok bool = r.write_quorum()
	if (!write_ok) {
		return -1
	}

	return r.value
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	r.replicas[rid] = link
	r.rid_list = append(r.rid_list, rid)

	r.num_replicas += 1
	r.quorum_read = r.num_replicas/2 + 1
	r.quorum_write = r.num_replicas/2 + 1
}

func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	     _, get_ok := msg.(GetOp)
	set_op, set_ok := msg.(SetOp)

	if get_ok {
		return GetAck{val: r.value, time: r.lclock}
	}

	if set_ok {
		if r.lclock.Less(set_op.time) {
			r.lclock.Number = set_op.time.Number
			r.value = set_op.val
		}

		return SetAck{}
	}

	if !set_ok && !get_ok {
		fmt.Printf("ERROR: Received invalid message\n")
		os.Exit(1)
	}

	return nil
}
