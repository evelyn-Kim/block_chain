/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"strconv" // 문자열 -> 내부 변수 ( int, bool, float ) 혹은 반대
	"time"
	"log"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi" // hyperledger fabric chaincode GO SDK
)

type SmartContract struct {  // SmartContract 객체중 구조체
	contractapi.Contract // 상속
}

type Car struct { // world state 에 value 에 JSON으로 marshal 되어 저장될 구조체
	Make   string `json:"make"`
	Model  string `json:"model"`
	Colour string `json:"colour"`
	Owner  string `json:"owner"`
}

type QueryResult struct { // queryallcars 에서 검색된 K, V 쌍들을 배열로 만들기 위해 사용되는 검색결과 구조체
	Key    string `json:"Key"`
	Record *Car
}

// History 결과저장을 위한 구조체
type HistoryQueryResult struct{
	Record 		*Car		`json:"record"`
	TxId		string		`json:"txId"`
	Timestamp 	time.Time	`json:"timestamp"`
	IsDelete	bool		`json:"isDelete"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error { // 출력을 받는 주체는? 혹은 InitLedger를 수행해주는 주체는? =>  endorser
	cars := []Car{
		Car{Make: "Toyota", Model: "Prius", Colour: "blue", Owner: "Tomoko"},
		Car{Make: "Ford", Model: "Mustang", Colour: "red", Owner: "Brad"},
		Car{Make: "Hyundai", Model: "Tucson", Colour: "green", Owner: "Jin Soo"},
		Car{Make: "Volkswagen", Model: "Passat", Colour: "yellow", Owner: "Max"},
		Car{Make: "Tesla", Model: "S", Colour: "black", Owner: "Adriana"},
		Car{Make: "Peugeot", Model: "205", Colour: "purple", Owner: "Michel"},
		Car{Make: "Chery", Model: "S22L", Colour: "white", Owner: "Aarav"},
		Car{Make: "Fiat", Model: "Punto", Colour: "violet", Owner: "Pari"},
		Car{Make: "Tata", Model: "Nano", Colour: "indigo", Owner: "Valeria"},
		Car{Make: "Holden", Model: "Barina", Colour: "brown", Owner: "Shotaro"},
	}

	for i, car := range cars {
		carAsBytes, _ := json.Marshal(car) // 구조체를 -> marshal -> JSON포맷으로 저장된 [] byte array 

		err := ctx.GetStub().PutState("CAR"+strconv.Itoa(i), carAsBytes) // Key: CAR0~9

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}

	return nil // error 가 일어나지 않았음을 리턴
}

func (s *SmartContract) CreateCar(ctx contractapi.TransactionContextInterface, carNumber string, make string, model string, colour string, owner string) error { // 매개변수 5개 넣어주는 주체는? endorser <= application(submitTransaction("CreateCar","CAR10","BMW","420D", "white", "bstudent"))
	
	// (TO DO) 오류검증 - 각 매개변수안에 유효값이 들어있는지 검사 
	
	car := Car{
		Make:   make,
		Model:  model,
		Colour: colour,
		Owner:  owner,
	}

	carAsBytes, _ := json.Marshal(car) // 생성한 구조체 Marshal ( 직렬화 )

	return ctx.GetStub().PutState(carNumber, carAsBytes)  
	// value: JSON format
	// endorser peer 반환 -> 서명 -> APP -> orderer -> commiter 동기화
}

func (s *SmartContract) QueryCar(ctx contractapi.TransactionContextInterface, carNumber string) (*Car, error) { // application - evaluateTransaction("QueryCar", "CAR10")
	carAsBytes, err := ctx.GetStub().GetState(carNumber) // state -> JSON format []byte

	if err != nil { // GetState, GetStub, ctx참조가 오류를 만났을때
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if carAsBytes == nil { // key가 저장된적이 없다 or delete된 경우
		return nil, fmt.Errorf("%s does not exist", carNumber)
	}

	// 정상적으로 조회가 된 경우
	car := new(Car) // 객체화 JSON -> 구조체
	_ = json.Unmarshal(carAsBytes, car) // call by REFERENCE

	return car, nil
}

func (s *SmartContract) QueryAllCars(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := ""  // CAR0
	endKey := ""	// CAR9

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey) // state를 범위로 검색하는 함수

	if err != nil { // GetStub, GetStateByRange 오류를 가지느냐?
		return nil, err
	}

	defer resultsIterator.Close() // defer 함수 끝났을때 예약

	results := []QueryResult{} 

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil { // Next가 오류가 있는지
			return nil, err
		}

		car := new(Car) // Car JSON 형식
		_ = json.Unmarshal(queryResponse.Value, car)

		queryResult := QueryResult{Key: queryResponse.Key, Record: car}
		results = append(results, queryResult)
	}

	return results, nil
}

func (s *SmartContract) ChangeCarOwner(ctx contractapi.TransactionContextInterface, carNumber string, newOwner string) error { // application - submitTransaction("ChangeCarOwner", "CAR10", "blockchain")
	
	car, err := s.QueryCar(ctx, carNumber)

	if err != nil {
		return err
	}
	// 검증 생략 -- 이전 소유자에서 전달되었는지 유효한지

	// 자산이동
	car.Owner = newOwner

	carAsBytes, _ := json.Marshal(car) // 직렬화 ( 구조체 -> JSON 포멧의 Byte[] )

	return ctx.GetStub().PutState(carNumber, carAsBytes)
}

// 7. GetHistory upgrade
func (t *SmartContract) GetHistory(ctx contractapi.TransactionContextInterface, carNumber string) ([]HistoryQueryResult, error) {
	log.Printf("GetAssetHistory: ID %v", carNumber) // 체인코드 컨테이너 -> docker logs dev-asset1...

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(carNumber)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var car Car
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &car)
			if err != nil {
				return nil, err
			}
		} else {
			car = Car{}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &car,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}