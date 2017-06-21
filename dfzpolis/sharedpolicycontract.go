/*
Copyright Rob Hekkelman 2017 All Rights Reserved.
*/

package sharedpolis

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	dfzproto "github.com/hekkelmr/dfzshared/dfzprotos"
	dfzutil "github.com/hekkelmr/dfzshared/dfzutils"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Query ...
//========================================================================================================================
func Polisafspraak_query(stub shim.ChaincodeStubInterface, polisafspraak dfzproto.Polisafspraak, args []string) pb.Response {
	fmt.Println("########## Query ##########")
	bsncode := args[0]
	jaar := args[1]

	currentStatus, err := polisafspraak_getBsnState(stub, polisafspraak, bsncode, jaar)
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
func Polisafspraak_getUZOVI(polisafspraak dfzproto.Polisafspraak) pb.Response {
	fmt.Println("########## getUZOVI ##########")
	bytes := []byte(polisafspraak.UzoviCode)
	return shim.Success(bytes)
}

// Init values ...
//========================================================================================================
func Polisafspraak_initValues(stub shim.ChaincodeStubInterface, polisafspraak dfzproto.Polisafspraak, args []string) pb.Response {
	fmt.Println("########## Init  values ##########")
	bsncode := args[0]
	jaar := args[1]
	var declaratie dfzproto.Declaratie

	_, err := dfzutil.GetPersoon(stub, bsncode)
	if err != nil {
		return shim.Error("BSN invalid")
	}

	err = polisafspraak_setBsnState(stub, polisafspraak, bsncode, jaar, polisafspraak.MaxAantalPerJaar, declaratie)
	if err != nil {

		return shim.Error("error saving via policycontactrepository")
	}

	return shim.Success(nil)
}

// Check if the claim is covered ...
//========================================================================================================================
func Polisafspraak_validateClaim(stub shim.ChaincodeStubInterface, polisafspraak dfzproto.Polisafspraak, args []string) pb.Response {
	var declaratie dfzproto.Declaratie

	fmt.Println("########## Validate Claim ##########")
	fmt.Println(args[0])

	bytes := []byte(args[0])

	err := json.Unmarshal(bytes, &declaratie)
	if err != nil {
		fmt.Println(err.Error())
		return polisafspraak_createResponse("FOUT", 0, 0, "", 0, err.Error(), "", nil)
	}

	// is deze declaratie al gedaan?
	fmt.Println("Checking Treatment")
	vorigeDeclaratie, err := polisafspraak_getTreatmentState(stub, polisafspraak, declaratie.Voorlooprecord.ReferentieBehandeling)
	if err != nil {
		fmt.Println("Error checking treatment")
		return polisafspraak_createResponse("FOUT", 0, 0, "", 0, err.Error(), "", nil)
	}
	if vorigeDeclaratie.Voorlooprecord.ReferentieBehandeling == declaratie.Voorlooprecord.ReferentieBehandeling {
		fmt.Println("Behandeling al verwerkt")
		return polisafspraak_createResponse("FOUT", 0, 0, "", 0, "Behandeling al verwerkt", "", nil)

	}

	// Does the person exists? (can this message be tamperd with???)
	fmt.Println("Checking BSN")
	patient, err := dfzutil.GetPersoon(stub, declaratie.Verzekerderecord.Bsncode)
	if err != nil {
		fmt.Println("Error checking BSN")
		return polisafspraak_createResponse("FOUT", 0, 0, "", 0, err.Error(), "", nil)
	}

	// Is there stil funding for this patient?
	fmt.Println("Checking Credits")
	jaar := strconv.Itoa(declaratie.Prestatierecords[0].DatumPrestatie.Year())

	iets, err := stub.GetHistoryForKey(declaratie.Verzekerderecord.Bsncode + ":" + jaar)
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

	currentStatus, err := polisafspraak_getBsnState(stub, polisafspraak, patient.Bsncode, jaar)
	if err != nil {
		fmt.Println("Error checking Credits")
		return polisafspraak_createResponse("FOUT", 0, 0, "", 0, err.Error(), "", nil)
	}

	// Check AGB
	fmt.Println("Checking AGB")
	agbcode := declaratie.Voorlooprecord.AGBPraktijk
	zorgaanbieder, err := dfzutil.GetZorgaanbieder(stub, agbcode)
	if err != nil {
		return polisafspraak_createResponse("FOUT", 0, 0, "", 0, "AGBPraktijk niet bekend", "", nil)
	}

	// Check AGB
	if declaratie.Voorlooprecord.AGBZorgverlener != "" {
		fmt.Println("Checking Zorgverlener")
		agbcode := declaratie.Voorlooprecord.AGBZorgverlener
		_, err = dfzutil.GetZorgverlener(stub, agbcode)
		if err != nil {
			return polisafspraak_createResponse("FOUT", 0, 0, "", 0, "AGB behandelaar niet bekend", "", nil)
		}
		fmt.Println("Checking werkzaam")
		agbcode = declaratie.Voorlooprecord.AGBPraktijk
		agbcode2 := declaratie.Voorlooprecord.AGBZorgverlener
		err = dfzutil.CheckWerkzaam(stub, agbcode, agbcode2)
		if err != nil {
			return polisafspraak_createResponse("FOUT", 0, 0, "", 0, "Behandelaar niet werkzaam in praktijk", "", nil)
		}
	}

	// Check locatie
	if polisafspraak.GebruikLocatieCheck {
		locatieOk := false
		if len(zorgaanbieder.GeoLocaties) > 0 {
			if declaratie.GeoLocatie.Lat != "" && declaratie.GeoLocatie.Lon != "" {
				latLoc, err := strconv.ParseFloat(declaratie.GeoLocatie.Lat, 64)
				if err != nil {
					fmt.Println("Fout bij omzetten latLoc")
				}
				lonLoc, err := strconv.ParseFloat(declaratie.GeoLocatie.Lon, 64)
				if err != nil {
					fmt.Println("Fout bij omzetten lonLoc")
				}
				for _, geoLocatie := range zorgaanbieder.GeoLocaties {
					latVest, err := strconv.ParseFloat(geoLocatie.Lat, 64)
					if err != nil {
						fmt.Println("Fout bij omzetten llatVest")
					}
					lonVest, err := strconv.ParseFloat(geoLocatie.Lon, 64)
					if err != nil {
						fmt.Println("Fout bij omzetten lonVest")
					}
					dist := dfzutil.Distance(latLoc, lonLoc, latVest, lonVest)
					if dist <= 500 {
						locatieOk = true
						break
					}
				}
			}
		}

		if !locatieOk {
			return polisafspraak_createResponse("FOUT", 0, 0, "", 0, "Niet aanwezig op geschikte locatie", "", nil)
		}
	}

	var totaalGedekt int64
	var totalClaimed int64
	var prestaties []dfzproto.PrestatieResultaat
	// Check supplier agreements ...
	for _, prestatieRecord := range declaratie.Prestatierecords {
		bericht := ""
		prslijst := prestatieRecord.Prestatiecodelijst
		prscode := prestatieRecord.Prestatiecode
		datum := prestatieRecord.DatumPrestatie.String()[0:10]
		contractPrestatie, err := polisafspraak_getContracted(stub, polisafspraak.UzoviCode, agbcode, prslijst, prscode, datum)
		if err != nil {
			fmt.Println("Geen contract afgesloten")
			contractPrestatie = dfzproto.Contractprestatie{"", "", 0, "", 0, ""}
		}

		gedekt := contractPrestatie.TariefPrestatie
		percentage := float32(contractPrestatie.Percentage) / 100.0
		// Not contracted
		if contractPrestatie.Herkomst == "" {
			bericht = fmt.Sprintf("Geen contractafspraak, maximale vergoeding %d procent\n", int64(polisafspraak.Factor*100.0))
			gedekt = int64((float32(declaratie.Prestatierecords[0].TariefPrestatie) / 100.0) * polisafspraak.Factor * 100.0)
		} else if contractPrestatie.Herkomst == "Contractafspraak" {
			if gedekt == 0 {
				bericht = fmt.Sprintf("Deze behandeling bij deze zorgverlener wordt niet vergoed\n")
				gedekt = 0.00
			} else if prestatieRecord.TariefPrestatie > gedekt {
				prestatieRecord.TariefPrestatie = gedekt
				bericht = fmt.Sprintf("Volgens contractafspraak met zorgverlener is het bedrag %.2f\n", float32(gedekt)/100.0)
			}
		} else if contractPrestatie.Herkomst == "Polisvoorwaarden" {
			if percentage == 0 {
				bericht = fmt.Sprintf("Volgens polisvoorwaarden is het bedrag %.2f\n", float32(gedekt)/100.0)
			} else {
				gedekt = int64((float32(gedekt) / 100.0) * percentage * 100.0)
				bericht = fmt.Sprintf("Volgens polisvoorwaarden %d perc. vergoedt: %.2f\n", int64(percentage*100.0), float32(gedekt)/100.0)
			}
		}
		if gedekt > prestatieRecord.TariefPrestatie {
			gedekt = prestatieRecord.TariefPrestatie
		}
		totalClaimed = totalClaimed + prestatieRecord.TariefPrestatie
		prestatieRecord.BerekendBedrag = gedekt
		totaalGedekt = totaalGedekt + gedekt
		prestatieResultaat := dfzproto.PrestatieResultaat{prestatieRecord, contractPrestatie.Omschrijving, bericht}
		prestaties = append(prestaties, prestatieResultaat)
	}

	var msg string
	var noclaim int64
	var tegoed int64
	tegoed = 0

	if polisafspraak.MaxAantalPerJaar > 0 {
		tegoed = currentStatus.Tegoed
		if polisafspraak.Eenheid == "behandelingen" && totaalGedekt > 0 {
			tegoed = tegoed - 1
		} else {
			tegoed = tegoed - totaalGedekt
		}

		if tegoed < 0 {
			msg = msg + "Uw heeft geen tegoed meer\n"
			totaalGedekt = 0
		}

		if totalClaimed > totaalGedekt {
			msg = msg + "Niet volledig vergoed, u moet zelf bijbetalen\n"
			noclaim = totalClaimed - totaalGedekt
		}
	}

	msg = msg + polisafspraak.UzoviCode
	return polisafspraak_createResponse("OK", tegoed, totaalGedekt, polisafspraak.Eenheid, noclaim, msg, declaratie.Voorlooprecord.AGBPraktijk, prestaties)
}

func polisafspraak_createResponse(result string, restant int64, vergoed int64, unity string, noclaim int64, bericht string, agbcode string, prestatierecords []dfzproto.PrestatieResultaat) pb.Response {
	antwoord := dfzproto.Retourbericht{agbcode, result, restant, vergoed, unity, noclaim, bericht, prestatierecords}
	bytes, _ := json.Marshal(antwoord)
	return shim.Success(bytes)
}

// Execute the claim ...
//========================================================================================================================
func Polisafspraak_doClaim(stub shim.ChaincodeStubInterface, polisafspraak dfzproto.Polisafspraak, args []string) pb.Response {
	fmt.Println("########## Do Claim ##########")
	var declaratie dfzproto.Declaratie
	var antwoord dfzproto.Retourbericht

	bytes := []byte(args[0])

	err := json.Unmarshal(bytes, &declaratie)
	if err != nil {
		fmt.Println(err.Error())
		return polisafspraak_createResponse("FOUT", 0, 0, "", 0, err.Error(), "", nil)
	}

	response := Polisafspraak_validateClaim(stub, polisafspraak, args)
	json.Unmarshal(response.Payload, &antwoord)
	if antwoord.Retourcode == "FOUT" {
		return shim.Error(antwoord.Bericht)
	}
	jaar := strconv.Itoa(declaratie.Prestatierecords[0].DatumPrestatie.Year())

	err = polisafspraak_setBsnState(stub, polisafspraak, declaratie.Verzekerderecord.Bsncode, jaar, antwoord.Restant, declaratie)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = polisafspraak_setAgbBalanceState(stub, polisafspraak, declaratie.Voorlooprecord.AGBPraktijk, jaar, antwoord.Vergoed)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = polisafspraak_setBsnBalanceState(stub, polisafspraak, declaratie.Voorlooprecord.AGBPraktijk, jaar, antwoord.Vergoed)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = polisafspraak_setContractBalanceState(stub, polisafspraak, jaar, antwoord.Vergoed)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = polisafspraak_setCompanyBalanceState(stub, polisafspraak, jaar, antwoord.Vergoed)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = polisafspraak_setTreatmentState(stub, polisafspraak, declaratie)
	if err != nil {
		return shim.Error(err.Error())
	}

	zorgaanbieder, err := dfzutil.GetZorgaanbieder(stub, antwoord.AgbCode)
	if err != nil || zorgaanbieder.WalletID == "" {
		msg := "Cannot retrieve zorgaanbieder Wallet"
		fmt.Println(msg)
		return shim.Error(msg)
	}

	patient, err := dfzutil.GetPersoon(stub, declaratie.Verzekerderecord.Bsncode)
	if err != nil || patient.WalletID == "" {
		msg := "Cannot retrieve patient Wallet"
		fmt.Println(msg)
		return shim.Error(msg)
	}

	verzekeraar, err := dfzutil.GetVerzekeraar(stub, polisafspraak.UzoviCode)
	if err != nil || verzekeraar.WalletID == "" {
		msg := "Cannot retrieve companyWallet"
		fmt.Println(msg)
		return shim.Error(msg)
	}

	err = dfzutil.DoeGecombneerdeBetaling(stub, verzekeraar.WalletID, zorgaanbieder.WalletID, antwoord.Vergoed, patient.WalletID, antwoord.Bijbetalen, "Uitbetaling declaratie "+stub.GetTxID())

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(response.Payload)
}

// Set State BSN ...
//========================================================================================================
func polisafspraak_setBsnState(stub shim.ChaincodeStubInterface, polisafspraak dfzproto.Polisafspraak, bsncode string, jaar string, treatments int64, declaratie dfzproto.Declaratie) error {

	key := polisafspraak.ContractCode + ":" + bsncode + ":" + jaar

	var currentStatus dfzproto.ContractStatus
	currentStatus.Tegoed = treatments
	currentStatus.Eenheid = polisafspraak.Eenheid
	currentStatus.Declaratie = declaratie
	bytes, err := json.Marshal(currentStatus)
	if err != nil {
		fmt.Printf("Error marshalling currentStatus\n")
		return errors.New("Error marshalling currentStatus")
	}

	invokeArgs := util.ToChaincodeArgs("set", key, string(bytes))
	response := stub.InvokeChaincode(polisafspraak.VerzekeraarRepository, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return errors.New("error saving via policycontactrepository")
	}
	stub.SetEvent("LOG_POLICYCONTRACT_ALTERED", bytes)

	return nil
}

// Get State of BSN...
//========================================================================================================================
func polisafspraak_getBsnState(stub shim.ChaincodeStubInterface, polisafspraak dfzproto.Polisafspraak, bsncode string, jaar string) (dfzproto.ContractStatus, error) {

	var currentStatus dfzproto.ContractStatus
	key := polisafspraak.ContractCode + ":" + bsncode + ":" + jaar

	invokeArgs := util.ToChaincodeArgs("query", key)
	response := stub.InvokeChaincode(polisafspraak.VerzekeraarRepository, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return currentStatus, errors.New("error saving via policycontactrepository")
	}

	json.Unmarshal(response.Payload, &currentStatus)

	return currentStatus, nil
}

// Set State AGB ...
//========================================================================================================
func polisafspraak_setAgbBalanceState(stub shim.ChaincodeStubInterface, polisafspraak dfzproto.Polisafspraak, agbcode string, jaar string, amount int64) error {

	key := "AGBBAL-" + agbcode + ":" + jaar
	oldAmount, err := Polisafspraak_getBalanceState(stub, polisafspraak.VerzekeraarRepository, key)
	if err != nil {
		return errors.New("error retrieving AGB state")
	}

	amount = amount + oldAmount

	val := strconv.FormatInt(amount, 10)
	fmt.Printf("Amount will be %s", val)
	invokeArgs := util.ToChaincodeArgs("set", key, val)
	response := stub.InvokeChaincode(polisafspraak.VerzekeraarRepository, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return errors.New("error saving via policycontactrepository")
	}

	return nil
}

// Get State of Treatment...
//========================================================================================================================
func polisafspraak_getTreatmentState(stub shim.ChaincodeStubInterface, polisafspraak dfzproto.Polisafspraak, treatmentId string) (dfzproto.Declaratie, error) {

	var declaratie dfzproto.Declaratie
	key := "TREATMENT-" + treatmentId

	invokeArgs := util.ToChaincodeArgs("query", key)
	response := stub.InvokeChaincode(polisafspraak.VerzekeraarRepository, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return declaratie, errors.New("error quering via policycontactrepository")
	}

	json.Unmarshal(response.Payload, &declaratie)

	return declaratie, nil
}

// Set State Treatment ...
//========================================================================================================
func polisafspraak_setTreatmentState(stub shim.ChaincodeStubInterface, polisafspraak dfzproto.Polisafspraak, declaratie dfzproto.Declaratie) error {

	key := "TREATMENT-" + declaratie.Voorlooprecord.ReferentieBehandeling

	val, err := json.Marshal(declaratie)
	if err != nil {
		msg := "Error marschalling declaratie"
		fmt.Println(msg)
		return errors.New(msg)
	}

	invokeArgs := util.ToChaincodeArgs("set", key, string(val))
	response := stub.InvokeChaincode(polisafspraak.VerzekeraarRepository, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return errors.New("error saving via policycontactrepository")
	}

	return nil
}

// Set State AGB ...
//========================================================================================================
func polisafspraak_setContractBalanceState(stub shim.ChaincodeStubInterface, polisafspraak dfzproto.Polisafspraak, jaar string, amount int64) error {
	fmt.Println("########## Set State ##########")

	key := polisafspraak.ContractCode + ":CTCBAL:" + jaar
	oldAmount, err := Polisafspraak_getBalanceState(stub, polisafspraak.VerzekeraarRepository, key)
	if err != nil {
		return errors.New("error retrieving AGB state")
	}

	amount = amount + oldAmount

	val := strconv.FormatInt(amount, 10)
	fmt.Printf("Amount will be %s", val)
	invokeArgs := util.ToChaincodeArgs("set", key, val)
	response := stub.InvokeChaincode(polisafspraak.VerzekeraarRepository, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return errors.New("error saving via policycontactrepository")
	}

	return nil
}

// Set Company Balance state ...
//========================================================================================================
func polisafspraak_setCompanyBalanceState(stub shim.ChaincodeStubInterface, polisafspraak dfzproto.Polisafspraak, jaar string, amount int64) error {

	fmt.Printf("Company %s balance set add amount:%d\n", polisafspraak.UzoviCode, amount)

	key := "UZOVIBAL-" + polisafspraak.UzoviCode + ":" + jaar
	oldAmount, err := Polisafspraak_getBalanceState(stub, polisafspraak.VerzekeraarRepository, key)
	if err != nil {
		return errors.New("error retrieving AGB state")
	}

	fmt.Printf("Company balance set old amount:%d\n", oldAmount)
	amount = amount + oldAmount

	val := strconv.FormatInt(amount, 10)
	fmt.Printf("Amount will be %s from %d", val, amount)

	invokeArgs := util.ToChaincodeArgs("set", key, val)
	response := stub.InvokeChaincode(polisafspraak.VerzekeraarRepository, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return errors.New("error saving via policycontactrepository")
	}

	return nil
}

// Set BSN balance state...
//========================================================================================================
func polisafspraak_setBsnBalanceState(stub shim.ChaincodeStubInterface, polisafspraak dfzproto.Polisafspraak, bsncode string, jaar string, amount int64) error {

	key := "BSNBAL-" + bsncode + ":" + jaar
	oldAmount, err := Polisafspraak_getBalanceState(stub, polisafspraak.VerzekeraarRepository, key)
	if err != nil {
		return errors.New("error retrieving BSN state")
	}

	amount = amount + oldAmount

	val := strconv.FormatInt(amount, 10)
	fmt.Printf("Amount will be %s", val)
	invokeArgs := util.ToChaincodeArgs("set", key, val)
	response := stub.InvokeChaincode(polisafspraak.VerzekeraarRepository, invokeArgs, "")
	if response.Status != shim.OK {
		fmt.Printf("Error saving via policycontactrepositor\n")
		return errors.New("error saving via policycontactrepository")
	}

	return nil
}

// Get State of AGB...
//========================================================================================================================
func Polisafspraak_getBalanceState(stub shim.ChaincodeStubInterface, polisafspraakRepository string, key string) (int64, error) {

	invokeArgs := util.ToChaincodeArgs("query", key)
	response := stub.InvokeChaincode(polisafspraakRepository, invokeArgs, "")
	if response.Status != shim.OK {
		return 0, errors.New("error querying policycontactrepository")
	}
	amount, _ := strconv.ParseInt(string(response.Payload), 10, 64)

	return amount, nil
}

// Check Coverage...
//========================================================================================================================
func polisafspraak_getContracted(stub shim.ChaincodeStubInterface, uzovicode string, agbcode string, prestielijst string, prestatiecode string, date string) (dfzproto.Contractprestatie, error) {
	var contractprestatie dfzproto.Contractprestatie

	invokeArgs := util.ToChaincodeArgs("queryContractedTreatment", uzovicode, agbcode, date, prestielijst, prestatiecode)
	response := stub.InvokeChaincode("zorgaanbiederszontractrouter", invokeArgs, "")
	if response.Status != shim.OK {
		msg := "Error query on healthcare contract"
		fmt.Println(msg + " " + response.Message)
		return contractprestatie, errors.New(msg)
	}
	json.Unmarshal(response.Payload, &contractprestatie)
	return contractprestatie, nil
}
