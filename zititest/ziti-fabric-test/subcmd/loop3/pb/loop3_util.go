package loop3_pb

import "ztna-core/ztna/logtrace"

const (
	BlockTypeRandomHashed = "random-hashed"
	BlockTypeSequential   = "sequential"
)

func (test *Test) IsRxRandomHashed() bool {
	logtrace.LogWithFunctionName()
	return test.RxBlockType == "" || test.RxBlockType == BlockTypeRandomHashed
}

func (test *Test) IsRxSequential() bool {
	logtrace.LogWithFunctionName()
	return test.RxBlockType == BlockTypeSequential
}

func (test *Test) IsTxRandomHashed() bool {
	logtrace.LogWithFunctionName()
	return test.TxBlockType == "" || test.TxBlockType == BlockTypeRandomHashed
}

func (test *Test) IsTxSequential() bool {
	logtrace.LogWithFunctionName()
	return test.TxBlockType == BlockTypeSequential
}
