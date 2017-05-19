/*
Copyright Rob Hekkelman 2017 All Rights Reserved.
*/

package dfzshared

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

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
	agbcode := declaratie.Voorlooprecord.AGBPraktijk
	_, err = GetCaregiver(stub, agbcode)
	if err != nil {
		return policyContract_createResponse("FOUT", 0, 0, "", 0, err.Error(), "")
	}

	// Check suppier agreements ...
	prslijst := declaratie.Prestatierecord.Prestatiecodelijst
	prscode := declaratie.Prestatierecord.Prestatiecode
	datum := declaratie.Prestatierecord.DatumPrestatie.String()[0:10]
	contractedTreatment, err := policyContract_getContracted(stub, policyContract.UzoviCode, agbcode, prslijst, prscode, datum)
	if err != nil {
		fmt.Println("Geen contract afgesloten")
		contractedTreatment = ContractedTreatment{"", "", 0, ""}
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
			msg = fmt.Sprintf("Volgens contractafspraak met zorgverlener is het bedrag %.2f\n", float32(covered)/100.0)
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
	msg = msg + policyContract.UzoviCode
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

	response := PolicyContract_validateClaim(stub, policyContract, args)
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

	err = policyContract_setContractBalanceState(stub, policyContract, year, antwoord.Vergoed)
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

	key := policyContract.ContractCode + ":" + bsncode + ":" + year

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

	var currentStatus ContractStatus
	key := policyContract.ContractCode + ":" + bsncode + ":" + year

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

	key := "AGBBAL-" + agbcode + ":" + year
	oldAmount, err := PolicyContract_getBalanceState(stub, policyContract.PolicyContractRepository, key)
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
func policyContract_setContractBalanceState(stub shim.ChaincodeStubInterface, policyContract PolicyContract, year string, amount int64) error {
	fmt.Println("########## Set State ##########")

	key := policyContract.ContractCode + ":CTCBAL:" + year
	oldAmount, err := PolicyContract_getBalanceState(stub, policyContract.PolicyContractRepository, key)
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

// Set Company Balance state ...
//========================================================================================================
func policyContract_setCompanyBalanceState(stub shim.ChaincodeStubInterface, policyContract PolicyContract, year string, amount int64) error {

	fmt.Printf("Company %s balance set add amount:%d\n", policyContract.UzoviCode, amount)

	key := "UZOVIBAL-" + policyContract.UzoviCode + ":" + year
	oldAmount, err := PolicyContract_getBalanceState(stub, policyContract.PolicyContractRepository, key)
	if err != nil {
		return errors.New("error retrieving AGB state")
	}

	fmt.Printf("Company balance set old amount:%d\n", oldAmount)
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

// Set BSN balance state...
//========================================================================================================
func policyContract_setBsnBalanceState(stub shim.ChaincodeStubInterface, policyContract PolicyContract, bsncode string, year string, amount int64) error {

	key := "BSNBAL-" + bsncode + ":" + year
	oldAmount, err := PolicyContract_getBalanceState(stub, policyContract.PolicyContractRepository, key)
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
func PolicyContract_getBalanceState(stub shim.ChaincodeStubInterface, policyContractRepository string, key string) (int64, error) {

	myRepo, err := GetDeployedChaincode(stub, policyContractRepository)
	if err != nil {
		fmt.Printf("Error getting policycontractrepository\n")
	}

	invokeArgs := util.ToChaincodeArgs("query", key)
	response := stub.InvokeChaincode(myRepo, invokeArgs, "")
	if response.Status != shim.OK {
		return 0, errors.New("error querying policycontactrepository")
	}

	fmt.Println(string(response.Payload))
	var amount int64
	json.Unmarshal(response.Payload, &amount)

	return amount, nil
}

// Check Coverage...
//========================================================================================================================
func policyContract_getContracted(stub shim.ChaincodeStubInterface, uzovicode string, agbcode string, prestielijst string, prestatiecode string, date string) (ContractedTreatment, error) {
	var treatment ContractedTreatment
	myRepo, err := GetDeployedChaincode(stub, "healthcarecontractrouter")
	if err != nil {
		fmt.Printf("Error getting healthcare contractrepository\n")
	}

	invokeArgs := util.ToChaincodeArgs("queryContractedTreatment", uzovicode, agbcode, date, prestielijst, prestatiecode)
	response := stub.InvokeChaincode(myRepo, invokeArgs, "")
	if response.Status != shim.OK {
		msg := "Error query on healthcare contract"
		fmt.Println(msg + " " + response.Message)
		return treatment, errors.New(msg)
	}
	json.Unmarshal(response.Payload, &treatment)
	return treatment, nil
}
