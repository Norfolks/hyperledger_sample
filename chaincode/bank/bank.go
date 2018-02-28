package main

import (
    "fmt"
    "bytes"
    "strconv"

    "github.com/hyperledger/fabric/core/chaincode/shim"
    sc "github.com/hyperledger/fabric/protos/peer"
)

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct {
}


// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) sc.Response {
    return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *SimpleAsset) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
    // Retrieve the requested Smart Contract function and arguments
    function, args := APIstub.GetFunctionAndParameters()
    // Route to the appropriate handler function to interact with the ledger appropriately
    if function == "queryUser" {
        return t.queryUser(APIstub, args)
    } else if function == "initLedger" {
        return t.initLedger(APIstub)
    } else if function == "createUser" {
        return t.createUser(APIstub, args)
    } else if function == "queryAllUsers" {
        return t.queryAllUsers(APIstub)
    } else if function == "transfer" {
        return t.transfer(APIstub, args)
    }

    return shim.Error("Invalid Smart Contract function name.")
}

func (t *SimpleAsset) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
    users := []string{
        "user1",
        "user2",
        "user3",
    }

    i := 0
    for i < len(users) {
        fmt.Println("i is ", i)
        APIstub.PutState(users[i], []byte("10"))
        fmt.Println("Added", users[i])
        i = i + 1
    }

    return shim.Success(nil)
}

//Transfer amounts from one's account to another
func (t *SimpleAsset) transfer(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
    if len(args) != 3 {
            return shim.Error("Incorrect arguments. Expecting account from, account to, amount")
    }
    tx_amount, _ := strconv.Atoi(args[2])
    if tx_amount < 0 {
        return shim.Error("Incorrect arguments. Amount can't be belove zero")
    }
    value, _ := APIstub.GetState(args[0])
    from_amount, _ := strconv.Atoi(string(value))
    if from_amount < tx_amount {
        return shim.Error("Not enought amounts at from account")
    }

    value, _ = APIstub.GetState(args[1])
    to_amount, _ := strconv.Atoi(string(value))

    //transaction logic
    from_amount -= tx_amount
    to_amount += tx_amount

    s := strconv.Itoa(from_amount)

    err := APIstub.PutState(args[0], []byte(s))
    if err != nil {
            return shim.Error(fmt.Sprintf("Failed save transaction: %s -> %s, %s", args[0], args[1], args[2]))
    }

    err = APIstub.PutState(args[1], []byte(strconv.Itoa(to_amount)))
    if err != nil {
            return shim.Error(fmt.Sprintf("Failed save transaction: %s -> %s, %s", args[0], args[1], args[2]))
    }
    return shim.Success([]byte(fmt.Sprintf("Success transaction: %s -> %s, %s", args[0], args[1], args[2])))
}

func (t *SimpleAsset) queryUser(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

    if len(args) != 1 {
        return shim.Error("Incorrect number of arguments. Expecting 1")
    }

    amount, _ := APIstub.GetState(args[0])
    return shim.Success(amount)
}

func (t *SimpleAsset) queryAllUsers(APIstub shim.ChaincodeStubInterface) sc.Response {

    startKey := "user1"
    endKey := "user999"

    resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
    if err != nil {
        return shim.Error(err.Error())
    }
    defer resultsIterator.Close()

    // buffer is a JSON array containing QueryResults
    var buffer bytes.Buffer
    buffer.WriteString("[")

    bArrayMemberAlreadyWritten := false
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return shim.Error(err.Error())
        }
        // Add a comma before array members, suppress it for the first array member
        if bArrayMemberAlreadyWritten == true {
            buffer.WriteString(",")
        }
        buffer.WriteString("{\"Key\":")
        buffer.WriteString("\"")
        buffer.WriteString(queryResponse.Key)
        buffer.WriteString("\"")

        buffer.WriteString(", \"Record\":")
        // Record is a JSON object, so we write as-is
        buffer.WriteString(string(queryResponse.Value))
        buffer.WriteString("}")
        bArrayMemberAlreadyWritten = true
    }
    buffer.WriteString("]")

    fmt.Printf("- queryAllUsers:\n%s\n", buffer.String())

    return shim.Success(buffer.Bytes())
}


func (t *SimpleAsset) createUser(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

    if len(args) != 2 {
        return shim.Error("Incorrect number of arguments. Expecting 2")
    }
    
    amount, _ := strconv.Atoi(args[1])
    if amount < 0 {
        return shim.Error("Incorrect arguments. Amount can't be belowe zero")
    }

    APIstub.PutState(args[0], []byte(args[1]))

    return shim.Success(nil)
}


// main function starts up the chaincode in the container during instantiate
func main() {
    if err := shim.Start(new(SimpleAsset)); err != nil {
            fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
    }
}
