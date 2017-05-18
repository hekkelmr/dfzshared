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
	pb "github.com/hyperledger/fabric/protos/peer"
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

//==============================================================
// SHared policyContract code

// Query ...
//========================================================================================================================
func PolicyContract_query(stub shim.ChaincodeStubInterface, policyContract PolicyContract, args []string) pb.Response {
	fmt.Println("########## Query ##########")
	bsncode := args[0]
	year := args[1]

	currentStatus, err := policyContract_getBsnState(stub, policyContract, bsncode, year)
	if err != nil {
		return shim.Error(err.Error())
	}
	bytes, err := json.Marshal(currentStatus)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(bytes)
}

// Query ...
//========================================================================================================================
func PolicyContract_getUZOVI(policyContract PolicyContract) pb.Response {
	fmt.Println("########## getUZOVI ##########")
	bytes := []byte(policyContract.UzoviCode)
	return shim.Success(bytes)
}

// Init values ...
//========================================================================================================
func PolicyContract_initValues(stub shim.ChaincodeStubInterface, policyContract PolicyContract, args []string) pb.Response {
	fmt.Println("########## Init  values ##########")
	bsncode := args[0]
	year := args[1]
	var declaratie Declaratie

	_, err := GetPerson(stub, bsncode)
	if err != nil {
		return shim.Error("BSN invalid")
	}

	err = policyContract_setBsnState(stub, policyContract, bsncode, year, policyContract.MaximumTreatmentsYear, declaratie)
	if err != nil {

		return shim.Error("error saving via policycontactrepository")
	}

	return shim.Success(nil)
}

// Check if the claim is covered ...
//========================================================================================================================
func PolicyContract_validateClaim(stub shim.ChaincodeStubInterface, policyContract PolicyContract, args []string) pb.Response {
	var declaratie Declaratie

	fmt.Println("########## Validate Claim ##########")
	fmt.Println(args[0])

	bytes := []byte(args[0])

	err := json.Unmarshal(bytes, &declaratie)
	if err != nil {
		fmt.Println(err.Error())
		return policyContract_createResponse("FOUT", 0, 0, "", 0, err.Error(), "")
	}

	// Does the person exists? (can this message be tamperd with???)
	fmt.Println("Checking BSN")
	patient, err := GetPerson(stub, declaratie.Verzekerderecord.Bsncode)
	if err != nil {
		fmt.Println("Error checking BSN")
		return policyContract_createResponse("FOUT", 0, 0, "", 0, err.Error(), "")
	}

	// Is there stil funding for this patient?
	fmt.Println("Checking Credits")
	year := strconv.Itoa(declaratie.Prestatierecord.DatumPrestatie.Year())

	//iets, err := stub.GetHistoryForKey(declaratie.VerzekerdeRecord.Bsncode + ":" + year)
	//fmt.Printf("Iets = %s\n", iets)

	currentStatus, err := policyContract_getBsnState(stub, policyContract, patient.Bsncode, year)
	if err != nil {
		fmt.Println("Error checking Credits")
		return policyContract_createResponse("FOUT", 0, 0, "", 0, err.Error(), "")
	}

	// Check AGB
	fmt.Println("Checking AGB")
	_, err = GetCaregiver(stub, declaratie.Voorlooprecord.AGBPraktijk)
	if err != nil {
		return policyContract_createResponse("FOUT", 0, 0, "", 0, err.Error(), "")
	}

	// Check suppier agreements ...
	contractedTreatment, err := policyContract_getContracted(stub, policyContract, declaratie, year)
	if err != nil {
		fmt.Println("Geen contract afgesloten")
	}

	var noclaim int64
	var msg string
	covered := contractedTreatment.TariefPrestatie
	// Not contracted
	if contractedTreatment.Prestatiecode == "" {
		msg = fmt.Sprintf("Geen contractafspraak, maximale vergoeding %d procent\n", int64(policyContract.Factor*100.0))
		covered = int64((float32(declaratie.Prestatierecord.BerekendBedrag) / 100.0) * policyContract.Factor * 100.0)
	} else {
		if covered == 0 {
			msg = fmt.Sprintf("Deze behandeling bij deze zorgverlener wordt niet vergoed\n")
		} else if declaratie.Prestatierecord.BerekendBedrag > covered {
			msg = fmt.Sprintf("Volgens contractafspraak met zorgverlener is het bedrag %.2f", float32(covered)/100.0)
			declaratie.Prestatierecord.BerekendBedrag = covered
		}
	}
	if covered > declaratie.Prestatierecord.BerekendBedrag {
		covered = declaratie.Prestatierecord.BerekendBedrag
	}
	remaining := currentStatus.Remaining

	if policyContract.Unity == "behandelingen" && covered > 0 {
		remaining = remaining - 1
	} else {
		remaining = remaining - covered
	}

	if remaining < 0 {
		msg = msg + "Uw heeft geen tegoed meer\n"
		covered = 0
	}

	if declaratie.Prestatierecord.BerekendBedrag > covered {
		msg = msg + "Niet volledig vergoed, u moet zelf bijbetalen\n"
		noclaim = declaratie.Prestatierecord.BerekendBedrag - covered
	}
	return policyContract_createResponse("OK", remaining, covered, policyContract.Unity, noclaim, msg, declaratie.Voorlooprecord.AGBPraktijk)
}

func policyContract_createResponse(result string, restant int64, vergoed int64, unity string, noclaim int64, bericht string, agbcode string) pb.Response {
	antwoord := Retourbericht{agbcode, result, restant, vergoed, unity, noclaim, bericht}
	bytes, _ := json.Marshal(antwoord)
	return shim.Success(bytes)
}

// Execute the claim ...
//========================================================================================================================
func PolicyContract_doClaim(stub shim.ChaincodeStubInterface, policyContract PolicyContract, args []string) pb.Response {
	fmt.Println("########## Do Claim ##########")
	var declaratie Declaratie
	var antwoord Retourbericht

	bytes := []byte(args[0])

	err := json.Unmarshal(bytes, &declaratie)
	if err != nil {
		fmt.Println(err.Error())
		return policyContract_createResponse("FOUT", 0, 0, "", 0, err.Error(), "")
	}

	response := policyContract_validateClaim(stub, policyContract, args)
	json.Unmarshal(response.Payload, &antwoord)
	if antwoord.Retourcode == "FOUT" {
		return shim.Error(antwoord.Bericht)
	}
	year := strconv.Itoa(declaratie.Prestatierecord.DatumPrestatie.Year())

	err = policyContract_setBsnState(stub, policyContract, declaratie.Verzekerderecord.Bsncode, year, antwoord.Restant, declaratie)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = policyContract_setAgbBalanceState(stub, policyContract, declaratie.Voorlooprecord.AGBPraktijk, year, antwoord.Vergoed)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = policyContract_setBsnBalanceState(stub, policyContract, declaratie.Voorlooprecord.AGBPraktijk, year, antwoord.Vergoed)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = policyContract_setContractBalanceState(stub, policyContract, policyContract.ChainName, year, antwoord.Vergoed)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = policyContract_setCompanyBalanceState(stub, policyContract, year, antwoord.Vergoed)
	if err != nil {
		return shim.Error(err.Error())
	}

	caregiver, err := GetCaregiver(stub, antwoord.AgbCode)
	if err != nil || caregiver.WalletID == "" {
		msg := "Cannot retrieve caregiver Wallet"
		fmt.Println(msg)
		return shim.Error(msg)
	}

	patient, err := GetPerson(stub, declaratie.Verzekerderecord.Bsncode)
	if err != nil || patient.WalletID == "" {
		msg := "Cannot retrieve patient Wallet"
		fmt.Println(msg)
		return shim.Error(msg)
	}

	company, err := GetInsuranceCompany(stub, policyContract.UzoviCode)
	if err != nil || company.WalletID == "" {
		msg := "Cannot retrieve companyWallet"
		fmt.Println(msg)
		return shim.Error(msg)
	}

	err = MakeCombinedPayment(stub, company.WalletID, caregiver.WalletID, antwoord.Vergoed, patient.WalletID, antwoord.Bijbetalen, "Uitbetaling declaratie "+stub.GetTxID())

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(response.Payload)
}

// Set State BSN ...
//========================================================================================================
func policyContract_setBsnState(stub shim.ChaincodeStubInterface, policyContract PolicyContract, bsncode string, year string, treatments int64, declaratie Declaratie) error {
	fmt.Println("########## Set State ##########")

	key := bsncode + ":" + year

	myRepo, err := GetDeployedChaincode(stub, policyContract.PolicyContractRepository)
	if err != nil {
		fmt.Printf("Error getting policycontractrepository\n")
		return errors.New("Error getting policycontractrepository")
	}

	var currentStatus ContractStatus
	currentStatus.Remaining = treatments
	currentStatus.Unity = policyContract.Unity
	currentStatus.Declaratie = declaratie
	bytes, err := json.Marshal(currentStatus)
	if err != nil {
		fmt.Printf("Error marshalling currentStatus\n")
		return errors.New("Error marshalling currentStatus")
	}

	invokeArgs := util.ToChaincodeArgs("set", key, string(bytes))
	response := stub.InvokeChaincode(myRepo, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return errors.New("error saving via policycontactrepository")
	}
	stub.SetEvent("LOG_POLICYCONTRACT_ALTERED", bytes)

	return nil
}

// Get State of BSN...
//========================================================================================================================
func policyContract_getBsnState(stub shim.ChaincodeStubInterface, policyContract PolicyContract, bsncode string, year string) (ContractStatus, error) {
	fmt.Println("########## Get State ##########")

	var currentStatus ContractStatus
	key := bsncode + ":" + year

	myRepo, err := GetDeployedChaincode(stub, policyContract.PolicyContractRepository)
	if err != nil {
		fmt.Printf("Error getting policycontractrepository\n")
	}

	invokeArgs := util.ToChaincodeArgs("query", key)
	response := stub.InvokeChaincode(myRepo, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return currentStatus, errors.New("error saving via policycontactrepository")
	}

	json.Unmarshal(response.Payload, &currentStatus)

	return currentStatus, nil
}

// Set State AGB ...
//========================================================================================================
func policyContract_setAgbBalanceState(stub shim.ChaincodeStubInterface, policyContract PolicyContract, agbcode string, year string, amount int64) error {
	fmt.Println("########## Set State ##########")

	key := "AGBBAL-" + agbcode + ":" + year
	oldAmount, err := policyContract_getBalanceState(stub, policyContract, key)
	if err != nil {
		return errors.New("error retrieving AGB state")
	}

	amount = amount + oldAmount

	myRepo, err := GetDeployedChaincode(stub, policyContract.PolicyContractRepository)
	if err != nil {
		fmt.Printf("Error getting policycontractrepository\n")
		return errors.New("Error getting policycontractrepository")
	}

	invokeArgs := util.ToChaincodeArgs("set", key, string(amount))
	response := stub.InvokeChaincode(myRepo, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return errors.New("error saving via policycontactrepository")
	}

	return nil
}

// Set State AGB ...
//========================================================================================================
func policyContract_setContractBalanceState(stub shim.ChaincodeStubInterface, policyContract PolicyContract, contractname string, year string, amount int64) error {
	fmt.Println("########## Set State ##########")

	key := "CTCBAL-" + contractname + ":" + year
	oldAmount, err := policyContract_getBalanceState(stub, policyContract, key)
	if err != nil {
		return errors.New("error retrieving AGB state")
	}

	amount = amount + oldAmount

	myRepo, err := GetDeployedChaincode(stub, policyContract.PolicyContractRepository)
	if err != nil {
		fmt.Printf("Error getting policycontractrepository\n")
		return errors.New("Error getting policycontractrepository")
	}

	invokeArgs := util.ToChaincodeArgs("set", key, string(amount))
	response := stub.InvokeChaincode(myRepo, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return errors.New("error saving via policycontactrepository")
	}

	return nil
}

// Set State AGB ...
//========================================================================================================
func policyContract_setCompanyBalanceState(stub shim.ChaincodeStubInterface, policyContract PolicyContract, year string, amount int64) error {
	fmt.Println("########## Set State ##########")

	key := "UZOVIBAL-" + policyContract.UzoviCode + ":" + year
	oldAmount, err := policyContract_getBalanceState(stub, policyContract, key)
	if err != nil {
		return errors.New("error retrieving AGB state")
	}

	amount = amount + oldAmount

	myRepo, err := GetDeployedChaincode(stub, policyContract.PolicyContractRepository)
	if err != nil {
		fmt.Printf("Error getting policycontractrepository\n")
		return errors.New("Error getting policycontractrepository")
	}

	invokeArgs := util.ToChaincodeArgs("set", key, string(amount))
	response := stub.InvokeChaincode(myRepo, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return errors.New("error saving via policycontactrepository")
	}

	return nil
}

// Set State AGB ...
//========================================================================================================
func policyContract_setBsnBalanceState(stub shim.ChaincodeStubInterface, policyContract PolicyContract, bsncode string, year string, amount int64) error {
	fmt.Println("########## Set State ##########")

	key := "BSNBAL-" + bsncode + ":" + year
	oldAmount, err := policyContract_getBalanceState(stub, policyContract, key)
	if err != nil {
		return errors.New("error retrieving BSN state")
	}

	amount = amount + oldAmount

	myRepo, err := GetDeployedChaincode(stub, policyContract.PolicyContractRepository)
	if err != nil {
		fmt.Printf("Error getting policycontractrepository\n")
		return errors.New("Error getting policycontractrepository")
	}

	invokeArgs := util.ToChaincodeArgs("set", key, string(amount))
	response := stub.InvokeChaincode(myRepo, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return errors.New("error saving via policycontactrepository")
	}

	return nil
}

// Get State of AGB...
//========================================================================================================================
func policyContract_getBalanceState(stub shim.ChaincodeStubInterface, policyContract PolicyContract, key string) (int64, error) {
	fmt.Println("########## Get State ##########")

	myRepo, err := GetDeployedChaincode(stub, policyContract.PolicyContractRepository)
	if err != nil {
		fmt.Printf("Error getting policycontractrepository\n")
	}

	invokeArgs := util.ToChaincodeArgs("query", key)
	response := stub.InvokeChaincode(myRepo, invokeArgs, "")
	if response.Status != shim.OK {
		return 0, errors.New("error querying policycontactrepository")
	}

	var amount int64
	json.Unmarshal(response.Payload, &amount)

	return amount, nil
}

// Check Coverage...
//========================================================================================================================
func policyContract_getContracted(stub shim.ChaincodeStubInterface, policyContract PolicyContract, request Declaratie, year string) (ContractedTreatment, error) {
	var treatment ContractedTreatment
	myRepo, err := GetDeployedChaincode(stub, "healthcarecontract")
	if err != nil {
		fmt.Printf("Error getting healthcare contractrepository\n")
	}
	agbcode := request.Voorlooprecord.AGBPraktijk

	invokeArgs := util.ToChaincodeArgs("queryContractedTreatment", policyContract.UzoviCode, agbcode, year, request.Prestatierecord.Prestatiecodelijst, request.Prestatierecord.Prestatiecode)
	response := stub.InvokeChaincode(myRepo, invokeArgs, "")
	if response.Status != shim.OK {
		msg := "Error query on healthcare contract"
		fmt.Println(msg + " " + response.Message)
		return treatment, errors.New(msg)
	}
	json.Unmarshal(response.Payload, &treatment)
	return treatment, nil
}
