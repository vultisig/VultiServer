package types

type Pricing struct {
	ID        string  `json:"id" validate:"required"`
	Type      string  `json:"type" validate:"required,oneof=FREE SINGLE RECURRING PER_TX"`
	Frequency *string `json:"frequency,omitempty" validate:"omitempty,oneof=ANNUAL MONTHLY WEEKLY"`
	Amount    float64 `json:"amount" validate:"gte=0"`
	Metric    string  `json:"metric" validate:"required,oneof=FIXED PERCENTAGE"`
}

type PricingCreateDto struct {
	Type      string  `json:"type" validate:"required,oneof=FREE SINGLE RECURRING PER_TX"`
	Frequency string  `json:"frequency,omitempty" validate:"omitempty,oneof=ANNUAL MONTHLY WEEKLY"`
	Amount    float64 `json:"amount" validate:"gte=0"`
	Metric    string  `json:"metric" validate:"required,oneof=FIXED PERCENTAGE"`
}
