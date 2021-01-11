package ftlregister

import (
	"context"
	"math"

	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
)

// FaultTolerantRegister хранит значение счетчиков на разных репликах
type FaultTolerantRegister struct {
	rid      int64
	current  util.TimestampedValue
	replicas map[int64]network.Link
}

// NewFaultTolerantRegister — конструктор регистра по rid и replicas
func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	return &FaultTolerantRegister{
		rid:      rid,
		replicas: map[int64]network.Link{},
	}
}

// "составные" READ  и WRITE
// READ = {read + write}
// WRITE = {read + write}

func (ftr *FaultTolerantRegister) Write(value int64) bool {
	currVal := BlockingMessageToQuorum(ftr, nil)
	currVal.Val = value
	currVal.Ts.Number = ftr.current.Ts.Number + 1
	currVal.Ts.Rid = ftr.rid

	ftr.current = currVal
	BlockingMessageToQuorum(ftr, value)

	return true
}

func (ftr *FaultTolerantRegister) Read() int64 {
	currVal := BlockingMessageToQuorum(ftr, nil)
	BlockingMessageToQuorum(ftr, currVal)

	return ftr.current.Val
}

// BlockingMessageToQuorum делает BlockingMessage запрос на все реплики
// и ждёт ответа от кворума реплик
func BlockingMessageToQuorum(ftr *FaultTolerantRegister, msg interface{}) util.TimestampedValue {
	messages := make(chan util.TimestampedValue, len(ftr.replicas))

	for _, rep := range ftr.replicas {
		go func(rep *network.Link) {
			// Send возвращает канал, передающий интерфейс
			messages <- (<-((*rep).Send(context.Background(), msg))).(util.TimestampedValue)
		}(&rep)
	}
	// close(messages)

	majority := int(math.Ceil(float64(len(ftr.replicas) / 2)))
	var counter = 0
	var newTsV util.TimestampedValue
	for elem := range messages {
		if counter < majority {
			if !elem.Ts.Less(newTsV.Ts) {
				newTsV = elem
			}
			counter++
		} else {
			break
		}

	}
	return newTsV
}

// Introduce устанавливает links между репликами
func (ftr *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		ftr.replicas[rid] = link
	}
}

// Receive принимает сообщения
func (ftr *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	if update, ok := msg.(util.TimestampedValue); ok {
		if ftr.current.Ts.Less(update.Ts) {
			ftr.current = update
		}
	}
	return ftr.current
}
