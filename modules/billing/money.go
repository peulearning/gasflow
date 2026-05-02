package billing

import "github.com/peulearnig/gasflow/internal/domain/shared"

// MoneyFromCents é um helper para o repositório reconstruir o value object a partir do banco.
func MoneyFromCents(cents int64) (shared.Money, error) {
	return shared.NewMoney(cents)
}