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
	AGBServicebureau         string `json:"agbservicebureau"`
	AGBZorgverlener          string `json:"agbzorgverlener"`
	AGBPraktijk              string `json:"agbpraktijk"`
	AGBInstelling            string `json:"agbinstelling"`
	IndentificatieBetalenAan string `json:"indetificatiebetalenaan"`
}

// EIVerzekerdeRecord ... The subject of a healthclaim
type EIVerzekerdeRecord struct {
	Bsncode string `json:"bsncode"`
	Naam    string `json:"naam"`
}

// EIPrestatieRecord ... The details of a healthclaim
type EIPrestatieRecord struct {
	Prestatiecodelijst string    `json:"prestatiecodelijst"`
	Prestatiecode      string    `json:"prestatiecode"`
	TariefPrestatie    int64     `json:"tariefprestatie"`
	BerekendBedrag     int64     `json:"berekendbedrag"`
	DatumPrestatie     time.Time `json:"datumprestatie"`
}

// Declaratie ... Combined healtclaim structure
type Declaratie struct {
	Voorlooprecord   EIVoorloopRecord   `json:"voorlooprecord"`
	VerzekerdeRecord EIVerzekerdeRecord `json:"verzekerderecord"`
	PrestatieRecord  EIPrestatieRecord  `json:"prestatierecord"`
}

// ContractStatus ...
// The status of a policycontract
type ContractStatus struct {
	Remaining  int64      `json:"remaining"`
	Unity      string     `json:"unity"`
	Declaratie Declaratie `json:"declaratie"`
}

// PolicyYear ...
// The name of the policycontrcat per year
type PolicyYear struct {
	Year               string `json:"year"`
	PolicyContractName string `json:"policycontractname"`
}

// PatientPolicyStatus ...
// The status of a PatientPolicy combined wit the ContractStatus
type PatientPolicyStatus struct {
	PatientPolicy  PatientPolicy  `json:"patientpolicy"`
	ContractStatus ContractStatus `json:"contractstatus"`
}

// PatientPolicy ...
// The policy a patient has in a given year
type PatientPolicy struct {
	Bsncode    string     `json:"bsncode"`
	PolicyYear PolicyYear `json:"policyyear"`
}

// Chain ...
// Technical contract to resolve the logical chain name to the phsyical one
type Chain struct {
	Name        string `json:"name"`
	ChaincodeID string `json:"chaincodeid"`
}

// Person ...
type Person struct {
	Bsncode  string    `json:"bsncode"`
	Name     string    `json:"name"`
	Dob      time.Time `json:"dob"`
	Dod      time.Time `json:"dod"`
	WalletID string    `json:"walletid"`
}

type CareGiver struct {
	AgbCode  string `json:"agbcode"`
	Name     string `json:"name"`
	WalletID string `json:"walletid"`
}

type InsuranceCompany struct {
	UzoviCode string `json:"uzovicode"`
	Name      string `json:"name"`
	Prefix    string `json:"prefix"`
	WalletID  string `json:"walletid"`
}

type AssignedPolicies struct {
	years []PolicyYear `json:"years"`
}

// Retourbericht ... The outcome message for a healthclaim
type Retourbericht struct {
	AgbCode        string `json:"agbcode"`
	Retourcode     string `json:"retourcode"`
	Restant        int64  `json:"restant"`
	Vergoed        int64  `json:"vergoed"`
	RestantEenheid string `json:"restanteenheid"`
	Bijbetalen     int64  `json:"bijbetalen"`
	Bericht        string `json:"bericht"`
}

type ContractedTreatment struct {
	Prestatiecodelijst string `json:"prestatiecodelijst"`
	Prestatiecode      string `json:"prestatiecode"`
	TariefPrestatie    int64  `json:"tariefprestatie"`
	Omschrijving       string `json:"omschrijving"`
}

type HealthCareContract struct {
	UzoviCode            string                `json:"uzovicode"`
	AgbCode              string                `json:"agbcode"`
	Year                 string                `json:"year"`
	ContractedTreatments []ContractedTreatment `json:"contractedtreatments"`
}

type WalletTransaction struct {
	From   string `json:"from"`
	Amount int64  `json:"amount"`
	Data   string `json:"data"`
}

type CurecoinWallet struct {
	ID                string            `json:"id"`
	Balance           int64             `json:"balance"`
	LatestTransaction WalletTransaction `json:"latesttransaction"`
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

// Transfer ...
func Transfer(stub shim.ChaincodeStubInterface, from string, to string, value int64, data string) error {
	walletChain, err := GetDeployedChaincode(stub, "curecoinwallet")
	if err != nil {
		msg := "Curecoinwallet contract does not exist"
		fmt.Println(msg)
		return errors.New(msg)
	}

	function := "transfer"
	amount := strconv.FormatInt(value, 10)
	fmt.Printf("From %s to %s amount %s\n", from, to, amount)
	invokeArgs := util.ToChaincodeArgs(function, from, to, amount, data)
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
