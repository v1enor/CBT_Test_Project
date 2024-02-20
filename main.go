package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type Account struct {
	Balance float64 `json:"balance"`
	IBAN    string  `json:"iban"`
	Status  string  `json:"status"`
}

type PaymentSystem struct {
	accounts map[string]*Account // словарь для быстрого поиска счета по IBAN
}

func NewPaynemtSystem(emitionIBAN string, destroyIBAN string) *PaymentSystem {
	ps := &PaymentSystem{
		accounts: make(map[string]*Account),
	}
	ps.accounts["emition"] = &Account{IBAN: emitionIBAN, Balance: 0, Status: "active"}
	ps.accounts["destroy"] = &Account{IBAN: destroyIBAN, Balance: 0, Status: "active"}
	log.Printf("Создана платежная система")
	return ps
}

func (ps *PaymentSystem) EmitBalance(amount float64) {
	account := ps.accounts["emition"]
	account.Balance += amount
	ps.accounts["emition"] = account
	log.Printf("Эмиссия %f выполнена успешно", amount)
}

func (ps *PaymentSystem) DestroyBalance(fromIBAN string, amount float64) error {
	ps.Transfer(fromIBAN, "destroy", amount)
	ps.ClearDestroy()
	return nil
}

func (ps *PaymentSystem) Transfer(fromIBAN, toIBAN string, amount float64) error {
	log.Printf("Попытка перевод %f с %s на %s", amount, fromIBAN, toIBAN)

	fromAccount, err := ps.accounts[fromIBAN]
	if !err {
		log.Printf("Ошибка: отправитель %s не найден", fromIBAN)
		return fmt.Errorf("отправитель %s не найден", fromIBAN)
	}

	toAccount, err := ps.accounts[toIBAN]
	if !err {
		log.Printf("Ошибка: получатель %s не найден", toIBAN)
		return fmt.Errorf("получатель %s не найден", fromIBAN)
	}

	if fromAccount.Status != "active" || toAccount.Status != "active" {
		log.Printf("Ошибка: один из счетов заблокирован")
		return fmt.Errorf("один из счетов заблокирован")
	}
	if fromAccount.Balance < amount {
		log.Printf("Ошибка: на балансе %s недостаточно средств", fromIBAN)
		return fmt.Errorf("на балансе %s недостаточно средств", fromIBAN)
	}

	fromAccount.Balance -= amount
	toAccount.Balance += amount

	ps.accounts[fromIBAN] = fromAccount
	ps.accounts[toIBAN] = toAccount

	log.Printf("Перевод %f с %s на %s выполнен успешно", amount, fromIBAN, toIBAN)
	return nil
}

type TransferRequest struct {
	FromIBAN string  `json:"from_iban"`
	ToIBAN   string  `json:"to_iban"`
	Amount   float64 `json:"amount"`
}

func (ps *PaymentSystem) TransferJSON(jsonStr string) error {
	var request TransferRequest
	err := json.Unmarshal([]byte(jsonStr), &request)
	if err != nil {
		log.Printf("Ошибка при  JSON: %v", err)
		return fmt.Errorf("ошибка при десериализации JSON: %v", err)
	}

	return ps.Transfer(request.FromIBAN, request.ToIBAN, request.Amount)
}

// очистка счета уничтожения
func (ps *PaymentSystem) ClearDestroy() {
	account := ps.accounts["destroy"]
	account.Balance = 0
	ps.accounts["destroy"] = account
}

func (ps *PaymentSystem) CreateAccount(IBAN string) error {
	_, ok := ps.accounts[IBAN]
	if ok {
		log.Printf("Ошибка: счет %s уже существует", IBAN)
		return fmt.Errorf("счет %s уже существует", IBAN)
	}
	ps.accounts[IBAN] = &Account{IBAN: IBAN, Balance: 0, Status: "active"}
	return nil
}

func (ps *PaymentSystem) BlockAccount(IBAN string) error {
	account, err := ps.accounts[IBAN]
	if !err {
		log.Printf("Ошибка: счет %s не найден", IBAN)
		return fmt.Errorf("счет %s не найден", IBAN)
	}
	account.Status = "blocked"
	ps.accounts[IBAN] = account
	log.Printf("Счет %s заблокирован", IBAN)
	return nil
}

func (ps *PaymentSystem) UnblockAccount(IBAN string) error {
	account, err := ps.accounts[IBAN]
	if !err {
		log.Printf("Ошибка: счет %s не найден", IBAN)
		return fmt.Errorf("счет %s не найден", IBAN)
	}
	account.Status = "active"
	ps.accounts[IBAN] = account
	log.Printf("Счет %s разблокирован", IBAN)
	return nil
}

func (ps *PaymentSystem) PrintAccounts() {
	accountJson, err := json.Marshal(ps.accounts)
	if err != nil {
		log.Printf("Ошибка при выводе счетов %v", err)
		fmt.Println(err)
	}
	fmt.Println(string(accountJson))
}

func generateRandomDigits(n int) string {
	rand.Seed(time.Now().UnixNano())
	var number string
	for i := 0; i < n; i++ {
		number += fmt.Sprintf("%d", rand.Intn(10))
	}
	return number
}

func IBANGenerator() string {
	const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	randString := make([]byte, 6)
	for i := range randString {
		randString[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	country, bank := string(randString[:2]), string(randString[3:5])

	IBAN := country + generateRandomDigits(2) + bank + generateRandomDigits(20)
	return IBAN
}

func main() {

	// генерация случайных IBAN адресов для счетов эмиссии и уничтожения
	emitAdress, destroyAdres := IBANGenerator(), IBANGenerator()

	// мои IBAN адреса
	myAdress := "BY04CBDC00000000000000000000"
	myAdress2 := "BY04CBDC00000000000000000001"

	//cоздание платежной системы
	ps := NewPaynemtSystem(emitAdress, destroyAdres)

	// cоздание счетов
	ps.CreateAccount(myAdress)
	ps.CreateAccount(IBANGenerator())
	ps.CreateAccount(myAdress2)

	// эмиссия 1000
	ps.EmitBalance(1000)
	fmt.Println("Счета до операций")
	ps.PrintAccounts()

	// перевод 100 c эмиссии на мой счет
	ps.Transfer("emition", myAdress, 100)
	fmt.Println("Счета после перевода 100 с эмиссии на мой счет")
	ps.PrintAccounts()

	// уничтожение 10 с моего счета
	ps.DestroyBalance(myAdress, 10)
	fmt.Println("Счета после уничтожения 10 с моего счета")
	ps.PrintAccounts()

	// перевод 100 на другой счет в формате JSON
	jsonStr := `{
		"from_iban": "emition",
		"to_iban": "BY04CBDC00000000000000000000",
		"amount": 100
	}`

	ps.TransferJSON(jsonStr)
	fmt.Println("Счета после перевода 100 с эмиссии на мой счет в формате JSON")
	ps.PrintAccounts()

	// блокировка и разблокировка счета
	ps.BlockAccount(myAdress)
	fmt.Println("Счета после блокировки моего счета")
	ps.PrintAccounts()
	err := ps.Transfer(myAdress, myAdress2, 33)

	if err != nil {
		fmt.Println("Ошибка при переводе с заблокированного счета: ", err)
	}

	ps.UnblockAccount(myAdress)
	ps.Transfer(myAdress, myAdress2, 100)
	fmt.Println("Счета после перевода 100 с моего счета на другой")
	ps.PrintAccounts()
}