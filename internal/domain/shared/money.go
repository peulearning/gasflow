package shared

import (

	"errors"
	"fmt"
)

type Money struct {
	cents int64
}

var ErrNegativeMoney = errors.New("money : valor não pode ser negativo.")


func NewMoney(cents int64) (Money, error){
	if cents < 0 {
		return Money{}, ErrNegativeMoney
	}

	return Money{cents: cents}, nil
}


func MustMoney(cents int64) Money {
	m, err := NewMoney(cents)
	if err != nil{
		panic (err)
	}
	return m

}


func (m Money) Cents() int64 { return m.cents }

func (m Money) Add(other Money) Money {

	return Money{cents: m.cents + other.cents}
}

func (m Money) Multiply(factor int) Money {
	return Money{cents: m.cents * int64(factor)}
}

func (m Money) IsZero() bool { return m.cents == 0 }

func (m Money) String() string {
	return fmt.Sprintf("R$ %.2f", float64(m.cents)/100)
}