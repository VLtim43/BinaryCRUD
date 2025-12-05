package crypto

import (
	"errors"
	"math/big"
)

type SimpleRSA struct {
	// Public key components
	N *big.Int // modulus (n = p * q)
	E *big.Int // public exponent

	// Private key component
	D *big.Int // private exponent (d = e^-1 mod φ(n))

	// For educational visibility (normally kept secret)
	P   *big.Int // first prime
	Q   *big.Int // second prime
	Phi *big.Int // φ(n) = (p-1)(q-1)
}

// RSA Key Generation Steps:
// 1. Choose two distinct prime numbers p and q
// 2. Compute n = p * q (this is the modulus)
// 3. Compute φ(n) = (p-1)(q-1) (Euler's totient function)
// 4. Choose e such that 1 < e < φ(n) and gcd(e, φ(n)) = 1 (commonly 65537)
// 5. Compute d = e^(-1) mod φ(n) (modular multiplicative inverse)
func NewSimpleRSA(p, q int64) (*SimpleRSA, error) {
	if !isPrime(p) || !isPrime(q) {
		return nil, errors.New("p and q must be prime numbers")
	}
	if p == q {
		return nil, errors.New("p and q must be distinct primes")
	}

	rsa := &SimpleRSA{
		P: big.NewInt(p),
		Q: big.NewInt(q),
	}

	// Step 2: n = p * q
	rsa.N = new(big.Int).Mul(rsa.P, rsa.Q)

	// Step 3: φ(n) = (p-1)(q-1)
	pMinus1 := new(big.Int).Sub(rsa.P, big.NewInt(1))
	qMinus1 := new(big.Int).Sub(rsa.Q, big.NewInt(1))
	rsa.Phi = new(big.Int).Mul(pMinus1, qMinus1)

	// Step 4: Choose e (we use 65537, a common choice)
	// e must satisfy: gcd(e, φ(n)) = 1
	rsa.E = big.NewInt(65537)

	// If e is too large for our small primes, use a smaller e
	if rsa.E.Cmp(rsa.Phi) >= 0 {
		// Find a smaller valid e
		rsa.E = big.NewInt(3)
		for {
			gcd := new(big.Int).GCD(nil, nil, rsa.E, rsa.Phi)
			if gcd.Cmp(big.NewInt(1)) == 0 {
				break
			}
			rsa.E.Add(rsa.E, big.NewInt(2))
		}
	}

	// Verify gcd(e, φ(n)) = 1
	gcd := new(big.Int).GCD(nil, nil, rsa.E, rsa.Phi)
	if gcd.Cmp(big.NewInt(1)) != 0 {
		return nil, errors.New("cannot find suitable public exponent e")
	}

	// Step 5: d = e^(-1) mod φ(n)
	rsa.D = new(big.Int).ModInverse(rsa.E, rsa.Phi)
	if rsa.D == nil {
		return nil, errors.New("cannot compute modular inverse for d")
	}

	return rsa, nil
}

// NewSimpleRSADefault creates a SimpleRSA with default small primes for demonstration.
// Uses p=61 and q=53, giving n=3233 (enough for small messages).
func NewSimpleRSADefault() (*SimpleRSA, error) {
	return NewSimpleRSA(61, 53)
}

// EncryptByte encrypts a single byte.
// Encryption formula: c = m^e mod n
// where m is the plaintext, e is the public exponent, n is the modulus.
func (r *SimpleRSA) EncryptByte(plaintext byte) *big.Int {
	m := big.NewInt(int64(plaintext))
	// c = m^e mod n (modular exponentiation)
	c := new(big.Int).Exp(m, r.E, r.N)
	return c
}

// DecryptByte decrypts a single encrypted value back to a byte.
// Decryption formula: m = c^d mod n
// where c is the ciphertext, d is the private exponent, n is the modulus.
func (r *SimpleRSA) DecryptByte(ciphertext *big.Int) byte {
	// m = c^d mod n (modular exponentiation)
	m := new(big.Int).Exp(ciphertext, r.D, r.N)
	return byte(m.Int64())
}

// Encrypt encrypts a byte slice, returning a slice of big.Int ciphertexts.
// Each byte is encrypted individually due to the small key size.
func (r *SimpleRSA) Encrypt(plaintext []byte) []*big.Int {
	ciphertext := make([]*big.Int, len(plaintext))
	for i, b := range plaintext {
		ciphertext[i] = r.EncryptByte(b)
	}
	return ciphertext
}

// Decrypt decrypts a slice of big.Int ciphertexts back to bytes.
func (r *SimpleRSA) Decrypt(ciphertext []*big.Int) []byte {
	plaintext := make([]byte, len(ciphertext))
	for i, c := range ciphertext {
		plaintext[i] = r.DecryptByte(c)
	}
	return plaintext
}

// EncryptString encrypts a string and returns ciphertext as big.Int slice.
func (r *SimpleRSA) EncryptString(plaintext string) []*big.Int {
	return r.Encrypt([]byte(plaintext))
}

// DecryptString decrypts ciphertext back to a string.
func (r *SimpleRSA) DecryptString(ciphertext []*big.Int) string {
	return string(r.Decrypt(ciphertext))
}

func (r *SimpleRSA) GetKeyInfo() map[string]string {
	return map[string]string{
		"p (first prime)":             r.P.String(),
		"q (second prime)":            r.Q.String(),
		"n (modulus = p*q)":           r.N.String(),
		"φ(n) (totient = (p-1)(q-1))": r.Phi.String(),
		"e (public exponent)":         r.E.String(),
		"d (private exponent)":        r.D.String(),
	}
}

// isPrime checks if a number is prime using trial division.
func isPrime(n int64) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	for i := int64(3); i*i <= n; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// ModularExponentiation demonstrates the square-and-multiply algorithm.
// This is how m^e mod n is efficiently computed.
// Returns base^exp mod mod
func ModularExponentiation(base, exp, mod *big.Int) *big.Int {
	result := big.NewInt(1)
	base = new(big.Int).Mod(base, mod)

	// Convert exponent to binary and process bit by bit
	expCopy := new(big.Int).Set(exp)
	zero := big.NewInt(0)
	one := big.NewInt(1)
	two := big.NewInt(2)

	for expCopy.Cmp(zero) > 0 {
		// If current bit is 1, multiply result by base
		if new(big.Int).And(expCopy, one).Cmp(one) == 0 {
			result = new(big.Int).Mod(new(big.Int).Mul(result, base), mod)
		}
		// Square the base
		base = new(big.Int).Mod(new(big.Int).Mul(base, base), mod)
		// Right shift exponent (divide by 2)
		expCopy = new(big.Int).Div(expCopy, two)
	}

	return result
}

// ExtendedGCD demonstrates the extended Euclidean algorithm.
// Returns gcd, x, y such that: ax + by = gcd(a, b)
func ExtendedGCD(a, b *big.Int) (gcd, x, y *big.Int) {
	if b.Cmp(big.NewInt(0)) == 0 {
		return new(big.Int).Set(a), big.NewInt(1), big.NewInt(0)
	}

	gcd, x1, y1 := ExtendedGCD(b, new(big.Int).Mod(a, b))

	// x = y1
	x = y1
	// y = x1 - (a/b) * y1
	quotient := new(big.Int).Div(a, b)
	y = new(big.Int).Sub(x1, new(big.Int).Mul(quotient, y1))

	return gcd, x, y
}
