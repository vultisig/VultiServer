package policy

type TransactionPolicy struct {
	Limit         float64 `json:"limit"`
	Delay         int     `json:"delay"`
	EmailApproval string  `json:"email_approval"`
	Password      string  `json:"password"`
}
