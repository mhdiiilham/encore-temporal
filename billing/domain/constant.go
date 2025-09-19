package domain

const (
	// SignalAddLineItem is the Temporal signal name used to add a new Item to a Bill.
	SignalAddLineItem string = "ADD_LINE_ITEM"

	// SignalCloseBill is the Temporal signal name used to request closing a Bill.
	SignalCloseBill string = "CLOSE_BILL"

	// QueryTypeGetBilling is the Temporal query type used to fetch the current state of a Bill.
	QueryTypeGetBilling string = "getBill"

	// TemporalQueueName is the Temporal queue task name
	TemporalQueueName string = "billing-task-queue"
)
