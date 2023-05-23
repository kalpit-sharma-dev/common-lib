package protocol

import "github.com/google/uuid"

//TransactionID is a function to return transaction id from the protocole request
func TransactionID(req *Request) string {
	transactionID := req.Headers.GetKeyValue(HdrTransactionID)
	if 0 == len(transactionID) {
		transactionID = req.TransactionID
	}

	if 0 == len(transactionID) {
		return uuid.New().String()
	}

	return transactionID
}
