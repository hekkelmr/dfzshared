/*
Copyright Rob Hekkelman 2017 All Rights Reserved.
*/

package dfzshared

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
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
	Voorlooprecord   EIVoorloopRecord   `json:"Voorlooprecord"`
	Verzekerderecord EIVerzekerdeRecord `json:"Verzekerderecord"`
	Prestatierecord  EIPrestatieRecord  `json:"Prestatierecord"`
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
}

type CareGiver struct {
	Agbcode  string `json:"Agbcode"`
	Name     string `json:"Name"`
	Type     string `json:"Type"`
	Soort    string `json:"Soort"`
	WalletID string `json:"WalletID"`
}

type InsuranceCompany struct {
	Uzovicode string `json:"Uzovicode"`
	Name      string `json:"Name"`
	Prefix    string `json:"Prefix"`
	WalletID  string `json:"WalletID"`
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

type HealthCareContract struct {
	UzoviCode            string                `json:"UzoviCode"`
	AgbCode              string                `json:"AgbCode"`
	Year                 string                `json:"Year"`
	ContractedTreatments []ContractedTreatment `json:"ContractedTreatments"`
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

// GetDeployedChaincode ... Get a deployed chain
//========================================================================================================
func GetDeployedChaincode(stub shim.ChaincodeStubInterface, name string) (string, error) {
	var chain Chain
	f := "query"
	invokeArgs := util.ToChaincodeArgs(f, name)
	response := stub.InvokeChaincode("chains01", invokeArgs, "")
	if response.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to invoke chain01.")
		fmt.Printf(errStr)
		return "", errors.New(errStr)
	}
	err := json.Unmarshal(response.Payload, &chain)
	if err != nil {
		errStr := fmt.Sprintf("Failed to determine chaincodeid.")
		fmt.Printf(errStr)
		return "", errors.New(errStr)
	}
	return chain.ChaincodeID, nil
}

// GetCaregiver ... Check Caregiver ...
//========================================================================================================================
func GetCaregiver(stub shim.ChaincodeStubInterface, agbcode string) (CareGiver, error) {
	var caregiver CareGiver
	fmt.Printf("Searching %s\n", agbcode)

	caregiverRepo, err := GetDeployedChaincode(stub, "caregiver")
	if err != nil {
		return caregiver, err
	}

	function := "query"
	invokeArgs := util.ToChaincodeArgs(function, agbcode)
	response := stub.InvokeChaincode(caregiverRepo, invokeArgs, "")

	if response.Status != shim.OK {
		msg := "Failed to get state for caregiver chain"
		fmt.Println(msg)
		return caregiver, errors.New(msg)
	}

	err = json.Unmarshal(response.Payload, &caregiver)
	if err != nil {
		return caregiver, err
	}

	return caregiver, nil
}

// GetPerson ...
//========================================================================================================================
func GetPerson(stub shim.ChaincodeStubInterface, bsncode string) (Person, error) {
	var patient Person
	personRepo, err := GetDeployedChaincode(stub, "person")
	if err != nil {
		return patient, err
	}

	function := "query"
	invokeArgs := util.ToChaincodeArgs(function, bsncode)
	response := stub.InvokeChaincode(personRepo, invokeArgs, "")
	if response.Status != shim.OK {
		msg := "Failed to get state for patient chain"
		fmt.Println(msg)
		return patient, errors.New(msg)
	}

	err = json.Unmarshal(response.Payload, &patient)
	if err != nil {
		return patient, err
	}

	return patient, nil
}

// GetInsuranceCompany ...
//========================================================================================================================
func GetInsuranceCompany(stub shim.ChaincodeStubInterface, uzovicode string) (InsuranceCompany, error) {
	var company InsuranceCompany
	companyRepo, err := GetDeployedChaincode(stub, "insurancecompany")
	if err != nil {
		return company, err
	}

	function := "query"
	invokeArgs := util.ToChaincodeArgs(function, uzovicode)
	response := stub.InvokeChaincode(companyRepo, invokeArgs, "")
	if response.Status != shim.OK {
		msg := "Failed to get state for insurancecompany chain"
		fmt.Println(msg)
		return company, errors.New(msg)
	}

	err = json.Unmarshal(response.Payload, &company)
	if err != nil {
		return company, err
	}

	return company, nil
}

// GetWallet ...
//========================================================================================================================
func GetWallet(stub shim.ChaincodeStubInterface, id string) (CurecoinWallet, error) {
	var wallet CurecoinWallet
	repo, err := GetDeployedChaincode(stub, "curecoinwallet")
	if err != nil {
		return wallet, err
	}

	function := "query"
	invokeArgs := util.ToChaincodeArgs(function, id)
	response := stub.InvokeChaincode(repo, invokeArgs, "")
	if response.Status != shim.OK {
		msg := "Failed to get state for curecoinwallet chain"
		fmt.Println(msg)
		return wallet, errors.New(msg)
	}

	err = json.Unmarshal(response.Payload, &wallet)
	if err != nil {
		return wallet, err
	}

	return wallet, nil
}

// NewWallet ...
//========================================================================================================================
func NewWalletID(stub shim.ChaincodeStubInterface) (string, error) {
	var wallet CurecoinWallet
	repo, err := GetDeployedChaincode(stub, "curecoinwallet")
	if err != nil {
		return "", err
	}

	function := "create"
	invokeArgs := util.ToChaincodeArgs(function, "")
	response := stub.InvokeChaincode(repo, invokeArgs, "")
	if response.Status != shim.OK {
		msg := "Failed to get state for curecoinwallet chain"
		fmt.Println(msg)
		return "", errors.New(msg)
	}

	err = json.Unmarshal(response.Payload, &wallet)
	if err != nil {
		return "", err
	}

	return wallet.ID, nil
}

// MakePayment ...
func MakePayment(stub shim.ChaincodeStubInterface, from string, to string, value int64, data string) error {
	walletChain, err := GetDeployedChaincode(stub, "curecoinwallet")
	if err != nil {
		msg := "Curecoinwallet contract does not exist"
		fmt.Println(msg)
		return errors.New(msg)
	}

	function := "makepayment"
	amount := strconv.FormatInt(value, 10)
	fmt.Printf("From %s to %s amount %s\n", from, to, amount)
	invokeArgs := util.ToChaincodeArgs(function, from, to, amount, data)
	respWallet := stub.InvokeChaincode(walletChain, invokeArgs, "")
	if respWallet.Status != shim.OK {
		return errors.New(respWallet.Message)
	}
	return nil
}

// MakeCombinedPayment ...
func MakeCombinedPayment(stub shim.ChaincodeStubInterface, from1 string, to string, valueFrom1 int64, from2 string, valueFrom2 int64, data string) error {
	walletChain, err := GetDeployedChaincode(stub, "curecoinwallet")
	if err != nil {
		msg := "Curecoinwallet contract does not exist"
		fmt.Println(msg)
		return errors.New(msg)
	}

	function := "makecombinedpayment"
	amount1 := strconv.FormatInt(valueFrom1, 10)
	amount2 := strconv.FormatInt(valueFrom2, 10)

	invokeArgs := util.ToChaincodeArgs(function, from1, to, amount1, from2, amount2, data)
	respWallet := stub.InvokeChaincode(walletChain, invokeArgs, "")
	if respWallet.Status != shim.OK {
		return errors.New(respWallet.Message)
	}
	return nil
}

// CheckYear ...
func CheckYear(year string) (uint64, error) {
	intYear, err := strconv.ParseUint(year, 10, 32)
	if err != nil {
		return 0, err
	}
	if intYear < 2016 || intYear > 2099 {
		return 0, errors.New("Invalid year")
	}
	return intYear, nil
}
