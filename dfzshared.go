/*
Copyright Rob Hekkelman 2017 All Rights Reserved.
*/

package dfzshared

import (
	"time"
)

// EIVoorloopRecord ... The header record of a healthclaim
type EIVoorloopRecord struct {
	AGBServicebureau        string `json:"AGBServicebureau"`
	AGBZorgverlener         string `json:"AGBZorgverlener"`
	AGBPraktijk             string `json:"AGBPraktijk"`
	AGBInstelling           string `json:"AGBInstelling"`
	IdentificatieBetalenAan string `json:"IdentificatieBetalenAan"`
	ReferentieBehandeling   string `json:"ReferentieBehandeling"`
}

// EIVerzekerdeRecord ... The subject of a healthclaim
type EIVerzekerdeRecord struct {
	Bsncode string `json:"Bsncode"`
	Naam    string `json:"Naam"`
}

// EIPrestatieRecord ... The details of a healthclaim
type EIPrestatieRecord struct {
	Prestatiecodelijst string    `json:"Prestatiecodelijst"`
	Prestatiecode      string    `json:"Prestatiecode"`
	TariefPrestatie    int64     `json:"TariefPrestatie"`
	BerekendBedrag     int64     `json:"BerekendBedrag"`
	DatumPrestatie     time.Time `json:"DatumPrestatie"`
}

// Declaratie ... Combined healtclaim structure
type Declaratie struct {
	Voorlooprecord   EIVoorloopRecord    `json:"Voorlooprecord"`
	Verzekerderecord EIVerzekerdeRecord  `json:"Verzekerderecord"`
	Prestatierecords []EIPrestatieRecord `json:"Prestatierecords"`
}

// ContractStatus ...
// The status of a policycontract
type ContractStatus struct {
	Remaining  int64      `json:"Remaining"`
	Unity      string     `json:"Unity"`
	Declaratie Declaratie `json:"Declaratie"`
}

// PolicyYear ...
// The name of the policycontrcat per year
type PolicyYear struct {
	Year               string `json:"Year"`
	PolicyContractName string `json:"PolicyContractName"`
}

// PatientPolicyStatus ...
// The status of a PatientPolicy combined wit the ContractStatus
type PatientPolicyStatus struct {
	PatientPolicy  PatientPolicy  `json:"PatientPolicy"`
	ContractStatus ContractStatus `json:"ContractStatus"`
}

// PatientPolicy ...
// The policy a patient has in a given year
type PatientPolicy struct {
	Bsncode    string     `json:"Bsncode"`
	PolicyYear PolicyYear `json:"PolicyYear"`
}

// PolicyContract ...
type PolicyContract struct {
	UzoviCode                string
	ContractCode             string
	SupportedHealthArea      string
	MaximumTreatmentsYear    int64
	ChainName                string
	Unity                    string
	Factor                   float32
	PolicyContractRepository string
}

// Chain ...
// Technical contract to resolve the logical chain name to the phsyical one
type Chain struct {
	Name        string `json:"Name"`
	ChaincodeID string `json:"ChaincodeId"`
}

// Person ...
type Person struct {
	Bsncode  string    `json:"Bsncode"`
	Name     string    `json:"Name"`
	Dob      time.Time `json:"Dob"`
	Dod      time.Time `json:"Dod"`
	WalletID string    `json:"WalletID"`
	ID       string    `json:"ID"`
}

type CareGiver struct {
	Agbcode  string `json:"Agbcode"`
	Name     string `json:"Name"`
	Type     string `json:"Type"`
	Soort    string `json:"Soort"`
	WalletID string `json:"WalletID"`
	ID       string `json:"ID"`
}

type InsuranceCompany struct {
	Uzovicode string `json:"Uzovicode"`
	Name      string `json:"Name"`
	Prefix    string `json:"Prefix"`
	WalletID  string `json:"WalletID"`
	ID        string `json:"ID"`
}

type SimpleEHR struct {
	TreatmentID string     `json:"TreatmentID"`
	Patient     Person     `json:"Patient"`
	Submitter   CareGiver  `json:"Submitter"`
	Diagnosis   string     `json:"Diagnosis"`
	Treatment   string     `json:"Treatment"`
	Claim       Declaratie `json:"Claim"`
	PreviousID  string     `json:"PreviousID"`
	Reference   string     `json:"Reference"`
	Uitgevoerd  string     `json:"Uitgevoerd"`
}

type AssignedPolicies struct {
	years []PolicyYear `json:"years"`
}

// Retourbericht ... The outcome message for a healthclaim
type Retourbericht struct {
	AgbCode        string `json:"AgbCode"`
	Retourcode     string `json:"Retourcode"`
	Restant        int64  `json:"Restant"`
	Vergoed        int64  `json:"Vergoed"`
	RestantEenheid string `json:"RestantEenheid"`
	Bijbetalen     int64  `json:"Bijbetalen"`
	Bericht        string `json:"Bericht"`
}

type ContractedTreatment struct {
	Prestatiecodelijst string `json:"Prestatiecodelijst"`
	Prestatiecode      string `json:"Prestatiecode"`
	TariefPrestatie    int64  `json:"TariefPrestatie"`
	Omschrijving       string `json:"Omschrijving"`
}

type HealthCareContractRouter struct {
	UzoviCode string `json:"UzoviCode"`
	RESTUrl   string `json:"RESTUrl"`
}

type WalletTransaction struct {
	From   string `json:"From"`
	Amount int64  `json:"Amount"`
	Data   string `json:"Data"`
}

type CurecoinWallet struct {
	ID                string            `json:"ID"`
	Balance           int64             `json:"Balance"`
	LatestTransaction WalletTransaction `json:"LatestTransaction"`
}

type AuthorizedEPD struct {
	Grantor string    `json:"Grantor"`
	Grantee string    `json:"Grantee"`
	Expires time.Time `json:"Expires"`
}
