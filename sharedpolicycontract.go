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

	err = PolicyContract_setBsnState(stub, policyContract, bsncode, year, policyContract.MaximumTreatmentsYear, declaratie)
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
		return policyContract_createResponse("FOUT", 0, 0, "", 0, err.Error(), "", nil)
	}

	// is deze declaratie al gedaan?
	fmt.Println("Checking Treatment")
	vorigeDeclaratie, err := policyContract_getTreatmentState(stub, policyContract, declaratie.Voorlooprecord.ReferentieBehandeling)
	if err != nil {
		fmt.Println("Error checking treatment")
		return policyContract_createResponse("FOUT", 0, 0, "", 0, err.Error(), "", nil)
	}
	if vorigeDeclaratie.Voorlooprecord.ReferentieBehandeling == declaratie.Voorlooprecord.ReferentieBehandeling {
		fmt.Println("Behandeling al verwerkt")
		return policyContract_createResponse("FOUT", 0, 0, "", 0, "Behandeling al verwerkt", "", nil)

	}

	// Does the person exists? (can this message be tamperd with???)
	fmt.Println("Checking BSN")
	patient, err := GetPerson(stub, declaratie.Verzekerderecord.Bsncode)
	if err != nil {
		fmt.Println("Error checking BSN")
		return policyContract_createResponse("FOUT", 0, 0, "", 0, err.Error(), "", nil)
	}

	// Is there stil funding for this patient?
	fmt.Println("Checking Credits")
	year := strconv.Itoa(declaratie.Prestatierecords[0].DatumPrestatie.Year())

	iets, err := stub.GetHistoryForKey(declaratie.Verzekerderecord.Bsncode + ":" + year)
	if err == nil {
		for iets.HasNext() {
			result, err := iets.Next()
			if err != nil {
				fmt.Println("foutje bij iets")
			}
			fmt.Println(result)
		}
	}

	fmt.Printf("Iets = %s\n", iets)

	currentStatus, err := policyContract_getBsnState(stub, policyContract, patient.Bsncode, year)
	if err != nil {
		fmt.Println("Error checking Credits")
		return policyContract_createResponse("FOUT", 0, 0, "", 0, err.Error(), "", nil)
	}

	// Check AGB
	fmt.Println("Checking AGB")
	agbcode := declaratie.Voorlooprecord.AGBPraktijk
	caregiver, err := GetCaregiver(stub, agbcode)
	if err != nil {
		return policyContract_createResponse("FOUT", 0, 0, "", 0, "AGBPraktijk niet bekend", "", nil)
	}

	// Check AGB
	if declaratie.Voorlooprecord.AGBZorgverlener != "" {
		fmt.Println("Checking Zorgverlener")
		agbcode := declaratie.Voorlooprecord.AGBZorgverlener
		_, err = GetZorgverlener(stub, agbcode)
		if err != nil {
			return policyContract_createResponse("FOUT", 0, 0, "", 0, "AGB behandelaar niet bekend", "", nil)
		}
		fmt.Println("Checking werkzaam")
		agbcode = declaratie.Voorlooprecord.AGBPraktijk
		agbcode2 := declaratie.Voorlooprecord.AGBZorgverlener
		err = GetWerkzaam(stub, agbcode, agbcode2)
		if err != nil {
			return policyContract_createResponse("FOUT", 0, 0, "", 0, "Behandelaar niet werkzaam in praktijk", "", nil)
		}
	}

	// Check locatie
	if policyContract.GebruikLocatieCheck {
		locatieOk := false
		if len(caregiver.GeoLocaties) > 0 {
			if declaratie.GeoLocatie.Lat != "" && declaratie.GeoLocatie.Lon != "" {
				latLoc, err := strconv.ParseFloat(declaratie.GeoLocatie.Lat, 64)
				if err != nil {
					fmt.Println("Fout bij omzetten latLoc")
				}
				lonLoc, err := strconv.ParseFloat(declaratie.GeoLocatie.Lon, 64)
				if err != nil {
					fmt.Println("Fout bij omzetten lonLoc")
				}
				for _, geoLocatie := range caregiver.GeoLocaties {
					latVest, err := strconv.ParseFloat(geoLocatie.Lat, 64)
					if err != nil {
						fmt.Println("Fout bij omzetten llatVest")
					}
					lonVest, err := strconv.ParseFloat(geoLocatie.Lon, 64)
					if err != nil {
						fmt.Println("Fout bij omzetten lonVest")
					}
					dist := Distance(latLoc, lonLoc, latVest, lonVest)
					if dist <= 500 {
						locatieOk = true
						break
					}
				}
			}
		}

		if !locatieOk {
			return policyContract_createResponse("FOUT", 0, 0, "", 0, "Niet aanwezig op geschikte locatie", "", nil)
		}
	}

	var totalCovered int64
	var totalClaimed int64
	var prestaties []PrestatieResultaat
	// Check supplier agreements ...
	for _, prestatieRecord := range declaratie.Prestatierecords {
		bericht := ""
		prslijst := prestatieRecord.Prestatiecodelijst
		prscode := prestatieRecord.Prestatiecode
		datum := prestatieRecord.DatumPrestatie.String()[0:10]
		contractedTreatment, err := policyContract_getContracted(stub, policyContract.UzoviCode, agbcode, prslijst, prscode, datum)
		if err != nil {
			fmt.Println("Geen contract afgesloten")
			contractedTreatment = ContractedTreatment{"", "", 0, "", 0, ""}
		}

		covered := contractedTreatment.TariefPrestatie
		percentage := float32(contractedTreatment.Percentage) / 100.0
		// Not contracted
		if contractedTreatment.Herkomst == "" {
			bericht = fmt.Sprintf("Geen contractafspraak, maximale vergoeding %d procent\n", int64(policyContract.Factor*100.0))
			covered = int64((float32(declaratie.Prestatierecords[0].TariefPrestatie) / 100.0) * policyContract.Factor * 100.0)
		} else if contractedTreatment.Herkomst == "Contractafspraak" {
			if covered == 0 {
				bericht = fmt.Sprintf("Deze behandeling bij deze zorgverlener wordt niet vergoed\n")
				covered = 0.00
			} else if prestatieRecord.TariefPrestatie > covered {
				prestatieRecord.TariefPrestatie = covered
				bericht = fmt.Sprintf("Volgens contractafspraak met zorgverlener is het bedrag %.2f\n", float32(covered)/100.0)
			}
		} else if contractedTreatment.Herkomst == "Polisvoorwaarden" {
			if percentage == 0 {
				bericht = fmt.Sprintf("Volgens polisvoorwaarden is het bedrag %.2f\n", float32(covered)/100.0)
			} else {
				covered = int64((float32(covered) / 100.0) * percentage * 100.0)
				bericht = fmt.Sprintf("Volgens polisvoorwaarden %d perc. vergoedt: %.2f\n", int64(percentage*100.0), float32(covered)/100.0)
			}
		}
		if covered > prestatieRecord.TariefPrestatie {
			covered = prestatieRecord.TariefPrestatie
		}
		totalClaimed = totalClaimed + prestatieRecord.TariefPrestatie
		prestatieRecord.BerekendBedrag = covered
		totalCovered = totalCovered + covered
		prestatieResultaat := PrestatieResultaat{prestatieRecord, contractedTreatment.Omschrijving, bericht}
		prestaties = append(prestaties, prestatieResultaat)
	}

	var msg string
	var noclaim int64
	var remaining int64
	remaining = 0

	if policyContract.MaximumTreatmentsYear > 0 {
		remaining = currentStatus.Remaining
		if policyContract.Unity == "behandelingen" && totalCovered > 0 {
			remaining = remaining - 1
		} else {
			remaining = remaining - totalCovered
		}

		if remaining < 0 {
			msg = msg + "Uw heeft geen tegoed meer\n"
			totalCovered = 0
		}

		if totalClaimed > totalCovered {
			msg = msg + "Niet volledig vergoed, u moet zelf bijbetalen\n"
			noclaim = totalClaimed - totalCovered
		}
	}

	msg = msg + policyContract.UzoviCode
	return policyContract_createResponse("OK", remaining, totalCovered, policyContract.Unity, noclaim, msg, declaratie.Voorlooprecord.AGBPraktijk, prestaties)
}

func policyContract_createResponse(result string, restant int64, vergoed int64, unity string, noclaim int64, bericht string, agbcode string, prestatierecords []PrestatieResultaat) pb.Response {
	antwoord := Retourbericht{agbcode, result, restant, vergoed, unity, noclaim, bericht, prestatierecords}
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
		return policyContract_createResponse("FOUT", 0, 0, "", 0, err.Error(), "", nil)
	}

	response := PolicyContract_validateClaim(stub, policyContract, args)
	json.Unmarshal(response.Payload, &antwoord)
	if antwoord.Retourcode == "FOUT" {
		return shim.Error(antwoord.Bericht)
	}
	year := strconv.Itoa(declaratie.Prestatierecords[0].DatumPrestatie.Year())

	err = PolicyContract_setBsnState(stub, policyContract, declaratie.Verzekerderecord.Bsncode, year, antwoord.Restant, declaratie)
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

	err = policyContract_setTreatmentState(stub, policyContract, declaratie)
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
func PolicyContract_setBsnState(stub shim.ChaincodeStubInterface, policyContract PolicyContract, bsncode string, year string, treatments int64, declaratie Declaratie) error {

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

	val := strconv.FormatInt(amount, 10)
	fmt.Printf("Amount will be %s", val)
	invokeArgs := util.ToChaincodeArgs("set", key, val)
	response := stub.InvokeChaincode(myRepo, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return errors.New("error saving via policycontactrepository")
	}

	return nil
}

// Get State of Treatment...
//========================================================================================================================
func policyContract_getTreatmentState(stub shim.ChaincodeStubInterface, policyContract PolicyContract, treatmentId string) (Declaratie, error) {

	var declaratie Declaratie
	key := "TREATMENT-" + treatmentId

	myRepo, err := GetDeployedChaincode(stub, policyContract.PolicyContractRepository)
	if err != nil {
		fmt.Printf("Error getting policycontractrepository\n")
	}

	invokeArgs := util.ToChaincodeArgs("query", key)
	response := stub.InvokeChaincode(myRepo, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return declaratie, errors.New("error quering via policycontactrepository")
	}

	json.Unmarshal(response.Payload, &declaratie)

	return declaratie, nil
}

// Set State Treatment ...
//========================================================================================================
func policyContract_setTreatmentState(stub shim.ChaincodeStubInterface, policyContract PolicyContract, declaratie Declaratie) error {

	key := "TREATMENT-" + declaratie.Voorlooprecord.ReferentieBehandeling

	myRepo, err := GetDeployedChaincode(stub, policyContract.PolicyContractRepository)
	if err != nil {
		fmt.Printf("Error getting policycontractrepository\n")
		return errors.New("Error getting policycontractrepository")
	}

	val, err := json.Marshal(declaratie)
	if err != nil {
		msg := "Error marschalling declaratie"
		fmt.Println(msg)
		return errors.New(msg)
	}

	invokeArgs := util.ToChaincodeArgs("set", key, string(val))
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

	val := strconv.FormatInt(amount, 10)
	fmt.Printf("Amount will be %s", val)
	invokeArgs := util.ToChaincodeArgs("set", key, val)
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

	val := strconv.FormatInt(amount, 10)
	fmt.Printf("Amount will be %s from %d", val, amount)

	invokeArgs := util.ToChaincodeArgs("set", key, val)
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

	val := strconv.FormatInt(amount, 10)
	fmt.Printf("Amount will be %s", val)
	invokeArgs := util.ToChaincodeArgs("set", key, val)
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
	amount, _ := strconv.ParseInt(string(response.Payload), 10, 64)

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
