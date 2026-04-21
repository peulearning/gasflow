packpage shared

import {

	"errors"
	"fmt"
}

type Money struct {
	conts int64
}

var ErrNegativeMoney = errors.New("money : valor não pode ser negativo.")


func NewMoney(conts int64) (Money, error){
	if conts < 0 {
		return Money{}, ErrNegativeMoney
	}

	return Money(conts: conts), nil
}


func MustMoney(conts int64) Money {
	m, err := NewMoney(conts)
	if err != nil{
		panic (err)
	}
	return m

}


func (m Money) Conts() int64 { return m.cents }

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