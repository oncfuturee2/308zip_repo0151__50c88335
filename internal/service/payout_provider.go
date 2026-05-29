package service

import (
	"log"

	"distribution-commission/internal/models"
)

type PayoutProvider interface {
	Payout(withdraw *models.WithdrawRequest) (string, error)
	ValidateBankCard(bankCardNo, bankName, realName string) (bool, error)
}

type MockPayoutProvider struct{}

func NewMockPayoutProvider() *MockPayoutProvider {
	return &MockPayoutProvider{}
}

func (p *MockPayoutProvider) Payout(withdraw *models.WithdrawRequest) (string, error) {
	log.Printf("[MockPayoutProvider] Simulating payout: request_no=%s, amount=%d, bank_card=%s",
		withdraw.RequestNo, withdraw.Amount, withdraw.BankCardNo)

	mockTransactionID := "MOCK_TXN_" + withdraw.RequestNo

	log.Printf("[MockPayoutProvider] Payout successful: transaction_id=%s", mockTransactionID)

	return mockTransactionID, nil
}

func (p *MockPayoutProvider) ValidateBankCard(bankCardNo, bankName, realName string) (bool, error) {
	log.Printf("[MockPayoutProvider] Simulating bank card validation: card_no=%s, bank=%s, name=%s",
		maskBankCard(bankCardNo), bankName, realName)

	return true, nil
}

func maskBankCard(cardNo string) string {
	if len(cardNo) <= 8 {
		return "****"
	}
	return cardNo[:4] + "****" + cardNo[len(cardNo)-4:]
}
