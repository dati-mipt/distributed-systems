package template

type Operation interface {
}

type OperationResult interface {
}

type ReplicatedDataType interface {
	IsReadOnly(op Operation) bool
	IsUpdateOnly(op Operation) bool
	ComputeResult(op Operation, confirmed []Operation) OperationResult
}

type ReplicatedStorageTemplate interface {
	Perform(op Operation) OperationResult
}
