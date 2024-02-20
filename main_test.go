package main

import (
	"fmt"
	"testing"
)

func TestTransfer(t *testing.T) {
	testCases := []struct {
		name            string
		fromIBAN        string
		toIBAN          string
		amount          float64
		expectedError   error
		expectedBalance float64
	}{
		{
			name:            "Перевод с достаточным балансом",
			fromIBAN:        "emition",
			toIBAN:          "BY04CBDC00000000000000000000",
			amount:          100,
			expectedError:   nil,
			expectedBalance: 100,
		},
		{
			name:            "Перевод с малым балансом",
			fromIBAN:        "emition",
			toIBAN:          "BY04CBDC00000000000000000000",
			amount:          2000,
			expectedError:   fmt.Errorf("на балансе %s недостаточно средств", "emition"),
			expectedBalance: 0,
		},

		{
			name:            "Перевод с заблокированного счета",
			fromIBAN:        "BY04CBDC00000000000000000001",
			toIBAN:          "BY04CBDC00000000000000000000",
			amount:          2000,
			expectedError:   fmt.Errorf("один из счетов заблокирован"),
			expectedBalance: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			emitAdress, destroyAdres := IBANGenerator(), IBANGenerator()
			myAdress := "BY04CBDC00000000000000000000"
			myAdress2 := "BY04CBDC00000000000000000001"

			ps := NewPaynemtSystem(emitAdress, destroyAdres)
			ps.CreateAccount(myAdress)
			ps.CreateAccount(myAdress2)
			ps.EmitBalance(1000)

			if tc.name == "Перевод с заблокированного счета" {
				ps.BlockAccount(tc.fromIBAN)
			}

			err := ps.Transfer(tc.fromIBAN, tc.toIBAN, tc.amount)
			if err != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("Ожидалась ошибка '%v', получена '%v'", tc.expectedError, err)
			}

			account := ps.accounts[myAdress]
			if account.Balance != tc.expectedBalance {
				t.Errorf("Ожидалось %v, получено %v", tc.expectedBalance, account.Balance)
			}

		})
	}
}
