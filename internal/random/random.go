package random

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	mathrand "math/rand"
)

// Rand задаёт интерфейс для источника случайных чисел.
type Rand interface {
	Intn(n int) int
	Shuffle(n int, swap func(i, j int))
}

// CryptoRand — реализация Rand с криптографическим сидом.
type CryptoRand struct {
	r *mathrand.Rand
}

// NewCryptoRand создаёт новый CryptoRand с криптографически случайным сидом.
func NewCryptoRand() *CryptoRand {
	seedBytes := make([]byte, 8)

	if _, err := cryptoRand.Read(seedBytes); err != nil {
		return &CryptoRand{r: mathrand.New(mathrand.NewSource(1))}
	}

	seed := int64(binary.LittleEndian.Uint64(seedBytes))
	return &CryptoRand{r: mathrand.New(mathrand.NewSource(seed))}
}

// Intn возвращает случайное число в диапазоне [0, n)
func (c *CryptoRand) Intn(n int) int {
	return c.r.Intn(n)
}

// Shuffle перемешивает элементы с помощью случайного источника.
func (c *CryptoRand) Shuffle(n int, swap func(i, j int)) {
	c.r.Shuffle(n, swap)
}
