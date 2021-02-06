package apiclient

// AccountResource holds all information about Account
type AccountResource struct {
	Type           string             `json:"type"`                 // value "accounts"
	ID             string             `json:"id"`                   // account ID
	OrganisationID string             `json:"organisation_id"`      // organisation ID of the organisation by which this resource has been created
	Version        int                `json:"version"`              // A counter indicating how many times this resource has been modified. Starting with 0
	Attributes     *AccountAttributes `json:"attributes,omitempty"` // The specific attributes for Accounts
}

// AccountAttributes specific to Account resource
type AccountAttributes struct {
	Country                 string                 `json:"country"` // required
	BaseCurrency            *string                `json:"base_currency,omitempty"`
	AccountNumber           *string                `json:"account_number,omitempty"`
	BankID                  *string                `json:"bank_id,omitempty"`
	BankIDCode              *string                `json:"bank_id_code,omitempty"`
	BIC                     *string                `json:"bic,omitempty"`
	IBAN                    *string                `json:"iban,omitempty"`
	CustomerID              *string                `json:"customer_id,omitempty"`
	Name                    [4]string              `json:"name"` // required
	AlternativeNames        *[3]string             `json:"alternative_names,omitempty"`
	AccountClassification   *string                `json:"account_classification,omitempty"`
	JointAccount            *bool                  `json:"joint_account,omitempty"`
	AccountMatchingOptOut   *bool                  `json:"account_matching_opt_out,omitempty"`
	SecondaryIdentification *string                `json:"secondary_identification,omitempty"`
	Switched                *bool                  `json:"switched,omitempty"`
	PrivateIdentification   *PrivateIdentification `json:"private_identification,omitempty"`
	Status                  *string                `json:"status"`
}

// PrivateIdentification holds information about Account holder
type PrivateIdentification struct {
	BirthDate      *string  `json:"birth_date,omitempty"`
	BirthCountry   *string  `json:"birth_country,omitempty"`
	Identification string   `json:"identification"` // required
	Address        []string `json:"address,omitempty"`
	City           *string  `json:"city,omitempty"`
	Country        *string  `json:"country,omitempty"`
}

// Helper struct for Create action
type CreateAccountResourceRequestData struct {
	Data struct {
		Type           string             `json:"type"`
		ID             string             `json:"id"`
		OrganisationID string             `json:"organisation_id"`
		Attributes     *AccountAttributes `json:"attributes"`
	} `json:"data"`
}
