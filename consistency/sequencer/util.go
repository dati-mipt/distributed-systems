package sequencer

type Operation interface {
}

type OperationResult interface {
}

type ReplicatedDataType interface {
	IsReadOnly(op Operation) bool
	IsUpdateOnly(op Operation) bool
	ComputeResult(op Operation, confirmed []Operation) OperationResult
}
