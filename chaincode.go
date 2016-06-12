/*
Copyright IBM Corp. and SonokoMizuki 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
  "errors"
  "fmt"
  "strconv"
  "github.com/hyperledger/fabric/core/chaincode/shim"
)

type MizukiChaincode struct {
  rate int
}


func main() {
  chaincode := new(MizukiChaincode)
  chaincode.rate = 100
  err := shim.Start(new(MizukiChaincode))
  if err != nil {
    fmt.Printf("Error starting chaincode: %s", err)
  }
}

func (t *MizukiChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
  return nil, nil
}

func (t *MizukiChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
  switch function {
  case "remit":
    if len(args) != 3 {
      return nil, errors.New("Incorrect number of arguments. Expecting 3. (from string, to string, amount int)")
    }
    from := args[0]
    to := args[1]
    value := args[2]

    fmt.Printf("Remit! from:%s, to:%s, value:%s", from, to, value)

    fromVal, err := t.GetMoney(stub, from)
    if err != nil {
      return nil, err
    }
    toVal, err := t.GetMoney(stub, to)
    if err != nil {
      return nil, err
    }

    // Move value
    v, err := strconv.Atoi(value)
    fromVal = fromVal - v
    toVal = toVal + v

    err = t.putMoney(stub, from, fromVal)
    if err != nil {
      return nil, err
    }
    err = t.putMoney(stub,   to, toVal)
    if err != nil {
      return nil, err
    }

    return nil, nil

  case "exchange":
    if len(args) != 3 {
      return nil, errors.New("Incorrect number of arguments. Expecting 3. (account string, amount int, from {yen|mizuki})")
    }
    account := args[0]
    amount, err := strconv.Atoi(args[1])
    if err != nil {
      return nil, err
    }

    value, err := t.GetMoney(stub, account)
    if err != nil {
      return nil, err
    }

    fmt.Printf("Remit! from:%s, amount:%d, value:%d rate:%d", account, amount, value, t.rate)
    if args[2] == "yen"{
      value = value + amount * 100;
      // TODO: remove yen in the world ?
    }else if args[2] == "mizuki"{
      value = value - amount * t.rate;
      // TODO: supply yen to client
    }

    err = t.putMoney(stub, account, value)
    if err != nil {
      return nil, err
    }
    return nil, nil

  case "entry":
    if len(args) != 1 {
      return nil, errors.New("Incorrect number of arguments. Expecting 1. (account string)")
    }
    fmt.Printf("Entryt! new account:%s\n", args[0])

    err := t.putMoney(stub, args[0], 0)
    if err != nil {
      return nil, fmt.Errorf("entry operation failed. Error updating state: %s", err)
    }
    return nil, nil

  case "leave":
    if len(args) != 1 {
      return nil, errors.New("Incorrect number of arguments. Expecting 1. (account string)")
    }
    err := stub.DelState(args[0])
    if err != nil {
      return nil, fmt.Errorf("leave operation failed. Error updating state: %s", err)
    }
    return nil, nil

  default:
    return nil, errors.New("Unsupported operation")
  }
}

func (t *MizukiChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
  switch function {
    case "balance":
      if len(args) < 1 {
        return nil, errors.New("balance operation must include one argument, (account string)")
      }
      account := args[0]
      value, err := t.GetMoney(stub, account)
      if err != nil {
        return nil, fmt.Errorf("get operation failed. Error accessing state: %s", err)
      }

      return []byte(strconv.Itoa(value)), nil
    default:
      return nil, errors.New("Unsupported operation ID.q1")
    }
}


// Sub functions

func (t *MizukiChaincode) GetMoney(stub *shim.ChaincodeStub, account string) (int, error) {
  valBytes, err := stub.GetState(account)
  if err != nil {
    return 0, errors.New("Failed to get state")
  }
  if valBytes == nil {
    return 0, errors.New("Account not found")
  }
  return strconv.Atoi(string(valBytes))
}

func (t *MizukiChaincode) putMoney(stub *shim.ChaincodeStub, account string, value int) (error) {
  fmt.Printf("Put money [%s] <- %d \n", account, value)
  err := stub.PutState( account, []byte(strconv.Itoa(value)))
  if err != nil {
    fmt.Printf("Error putting state %s", err)
    return fmt.Errorf("operation failed. Error updating state: %s", err)
  }
  return nil
}

