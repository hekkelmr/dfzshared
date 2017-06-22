/*
Copyright Rob Hekkelman 2017 All Rights Reserved.
*/

package dfzutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"

	dfzp "github.com/hekkelmr/dfzshared/dfzprotos"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// GetZorgaanbieder ...
//========================================================================================================================
func GetZorgaanbieder(stub shim.ChaincodeStubInterface, agbcode string) (dfzp.Zorgaanbieder, error) {
	var zorgaanbieder dfzp.Zorgaanbieder
	fmt.Printf("Searching %s\n", agbcode)

	function := "query"
	invokeArgs := util.ToChaincodeArgs(function, agbcode)
	response := stub.InvokeChaincode("zorgaanbieder", invokeArgs, "")

	if response.Status != shim.OK {
		msg := "Failed to get state for caregiver chain"
		fmt.Println(msg)
		return zorgaanbieder, errors.New(msg)
	}

	err := json.Unmarshal(response.Payload, &zorgaanbieder)
	if err != nil {
		return zorgaanbieder, err
	}

	return zorgaanbieder, nil
}

// GetZorgverlener ...
//========================================================================================================================
func GetZorgverlener(stub shim.ChaincodeStubInterface, agbcode string) (dfzp.Zorgaanbieder, error) {
	var zorgaanbieder dfzp.Zorgaanbieder
	fmt.Printf("Searching %s\n", agbcode)

	function := "queryZorgverlener"
	invokeArgs := util.ToChaincodeArgs(function, agbcode)
	response := stub.InvokeChaincode("zorgaanbieder", invokeArgs, "")

	if response.Status != shim.OK {
		msg := "Failed to get state for caregiver chain"
		fmt.Println(msg)
		return zorgaanbieder, errors.New(msg)
	}

	err := json.Unmarshal(response.Payload, &zorgaanbieder)
	if err != nil {
		return zorgaanbieder, err
	}

	return zorgaanbieder, nil
}

// CheckWerkzaam ... Check of agbcode2 werkzaam is bij agbcode1
//========================================================================================================================
func CheckWerkzaam(stub shim.ChaincodeStubInterface, agbcode string, agbcode2 string) error {

	function := "queryWerkzaam"
	invokeArgs := util.ToChaincodeArgs(function, agbcode, agbcode2)
	response := stub.InvokeChaincode("zorgaanbieder", invokeArgs, "")

	if response.Status != shim.OK {
		msg := "Failed to get state for caregiver chain"
		fmt.Println(msg)
		return errors.New(msg)
	}

	return nil
}

// GetPersoon ...
//========================================================================================================================
func GetPersoon(stub shim.ChaincodeStubInterface, bsncode string) (dfzp.Persoon, error) {
	var patient dfzp.Persoon

	function := "query"
	invokeArgs := util.ToChaincodeArgs(function, bsncode)
	response := stub.InvokeChaincode("persoon", invokeArgs, "")
	if response.Status != shim.OK {
		msg := "Failed to get state for patient chain"
		fmt.Println(msg)
		return patient, errors.New(msg)
	}

	err := json.Unmarshal(response.Payload, &patient)
	if err != nil {
		return patient, err
	}

	return patient, nil
}

// GetVerzekeraar ...
//========================================================================================================================
func GetVerzekeraar(stub shim.ChaincodeStubInterface, uzovicode string) (dfzp.Verzekeraar, error) {
	var verzekeraar dfzp.Verzekeraar

	function := "query"
	invokeArgs := util.ToChaincodeArgs(function, uzovicode)
	response := stub.InvokeChaincode("verzekeraar", invokeArgs, "")
	if response.Status != shim.OK {
		msg := "Failed to get state for insurancecompany chain"
		fmt.Println(msg)
		return verzekeraar, errors.New(msg)
	}

	err := json.Unmarshal(response.Payload, &verzekeraar)
	if err != nil {
		return verzekeraar, err
	}

	return verzekeraar, nil
}

// GetWallet ...
//========================================================================================================================
func GetWallet(stub shim.ChaincodeStubInterface, id string) (dfzp.CurecoinWallet, error) {
	var wallet dfzp.CurecoinWallet

	function := "query"
	invokeArgs := util.ToChaincodeArgs(function, id)
	response := stub.InvokeChaincode("curecoinwallet", invokeArgs, "")
	if response.Status != shim.OK {
		msg := "Failed to get state for curecoinwallet chain"
		fmt.Println(msg)
		return wallet, errors.New(msg)
	}

	err := json.Unmarshal(response.Payload, &wallet)
	if err != nil {
		return wallet, err
	}

	return wallet, nil
}

// NewWalletID ...
//========================================================================================================================
func NewWalletID(stub shim.ChaincodeStubInterface) (string, error) {
	var wallet dfzp.CurecoinWallet

	function := "create"
	invokeArgs := util.ToChaincodeArgs(function, "")
	response := stub.InvokeChaincode("curecoinwallet", invokeArgs, "")
	if response.Status != shim.OK {
		msg := "Failed to get state for curecoinwallet chain"
		fmt.Println(msg)
		return "", errors.New(msg)
	}

	err := json.Unmarshal(response.Payload, &wallet)
	if err != nil {
		return "", err
	}

	return wallet.ID, nil
}

// DoeBetaling ...
func DoeBetaling(stub shim.ChaincodeStubInterface, from string, to string, value int64, data string) error {

	function := "makepayment"
	amount := strconv.FormatInt(value, 10)
	fmt.Printf("From %s to %s amount %s\n", from, to, amount)
	invokeArgs := util.ToChaincodeArgs(function, from, to, amount, data)
	respWallet := stub.InvokeChaincode("curecoinwallet", invokeArgs, "")
	if respWallet.Status != shim.OK {
		return errors.New(respWallet.Message)
	}
	return nil
}

// DoeGecombineerdeBetaling ...
func DoeGecombneerdeBetaling(stub shim.ChaincodeStubInterface, from1 string, to string, valueFrom1 int64, from2 string, valueFrom2 int64, data string) error {

	function := "makecombinedpayment"
	amount1 := strconv.FormatInt(valueFrom1, 10)
	amount2 := strconv.FormatInt(valueFrom2, 10)

	invokeArgs := util.ToChaincodeArgs(function, from1, to, amount1, from2, amount2, data)
	respWallet := stub.InvokeChaincode("curecoinwallet", invokeArgs, "")
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

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// Distance function returns the distance (in meters) between two points of
//     a given longitude and latitude relatively accurately (using a spherical
//     approximation of the Earth) through the Haversin Distance Formula for
//     great arc distance on a sphere with accuracy for small distances
//
// point coordinates are supplied in degrees and converted into rad. in the func
//
// distance returned is METERS!!!!!!
// http://en.wikipedia.org/wiki/Haversine_formula
func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

// ToChaincodeArgs converts string args to []byte args
func ToChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}
