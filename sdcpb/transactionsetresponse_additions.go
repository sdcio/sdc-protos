package sdcpb

import "fmt"

func (t *TransactionSetResponse) GetErrors() []string {
	result := []string{}

	for intentName, intentErrs := range t.GetIntents() {
		for _, e := range intentErrs.GetErrors() {
			result = append(result, fmt.Sprintf("intent: %s, %v", intentName, e))
		}
	}
	return result
}

func (t *TransactionSetResponse) Failed() bool {
	for _, intentErrs := range t.GetIntents() {
		if intentErrs.Failed() {
			return true
		}
	}
	return false
}
