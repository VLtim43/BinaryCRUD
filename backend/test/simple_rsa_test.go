package test

import (
	"BinaryCRUD/backend/crypto"
	"fmt"
	"math/big"
	"testing"
)

func TestSimpleRSAKeyGeneration(t *testing.T) {
	// Test with primes p=61, q=53
	rsa, err := crypto.NewSimpleRSA(61, 53)
	if err != nil {
		t.Fatalf("Failed to create SimpleRSA: %v", err)
	}

	// Verify n = p * q = 61 * 53 = 3233
	expectedN := big.NewInt(3233)
	if rsa.N.Cmp(expectedN) != 0 {
		t.Errorf("Expected n=%v, got n=%v", expectedN, rsa.N)
	}

	// Verify φ(n) = (p-1)(q-1) = 60 * 52 = 3120
	expectedPhi := big.NewInt(3120)
	if rsa.Phi.Cmp(expectedPhi) != 0 {
		t.Errorf("Expected φ(n)=%v, got φ(n)=%v", expectedPhi, rsa.Phi)
	}

	// Verify e * d ≡ 1 (mod φ(n))
	ed := new(big.Int).Mul(rsa.E, rsa.D)
	remainder := new(big.Int).Mod(ed, rsa.Phi)
	if remainder.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("e*d mod φ(n) should be 1, got %v", remainder)
	}

	t.Logf("Key Info:")
	for key, value := range rsa.GetKeyInfo() {
		t.Logf("  %s: %s", key, value)
	}
}

func TestSimpleRSAEncryptDecrypt(t *testing.T) {
	rsa, err := crypto.NewSimpleRSADefault()
	if err != nil {
		t.Fatalf("Failed to create SimpleRSA: %v", err)
	}

	// Test single byte encryption/decryption
	testBytes := []byte{0, 1, 65, 127, 200, 255}

	for _, original := range testBytes {
		// Encrypt: c = m^e mod n
		encrypted := rsa.EncryptByte(original)

		// Decrypt: m = c^d mod n
		decrypted := rsa.DecryptByte(encrypted)

		if decrypted != original {
			t.Errorf("Byte %d: decrypted to %d", original, decrypted)
		}
	}
}

func TestSimpleRSAStringEncryption(t *testing.T) {
	rsa, err := crypto.NewSimpleRSADefault()
	if err != nil {
		t.Fatalf("Failed to create SimpleRSA: %v", err)
	}

	testCases := []string{
		"Hi",
		"RSA",
		"Test",
		"Hello",
	}

	for _, original := range testCases {
		encrypted := rsa.EncryptString(original)
		decrypted := rsa.DecryptString(encrypted)

		if decrypted != original {
			t.Errorf("String '%s': decrypted to '%s'", original, decrypted)
		}
	}
}

func TestSimpleRSAMathDemonstration(t *testing.T) {
	// This test demonstrates the RSA math step by step
	rsa, err := crypto.NewSimpleRSA(61, 53)
	if err != nil {
		t.Fatalf("Failed to create SimpleRSA: %v", err)
	}

	// Let's encrypt the byte value 65 (ASCII 'A')
	m := big.NewInt(65)

	t.Log("=== RSA Encryption/Decryption Demo ===")
	t.Logf("Plaintext message m = %v (ASCII 'A')", m)
	t.Logf("Public key (n, e) = (%v, %v)", rsa.N, rsa.E)
	t.Logf("Private key d = %v", rsa.D)

	// Encryption: c = m^e mod n
	c := new(big.Int).Exp(m, rsa.E, rsa.N)
	t.Logf("Encryption: c = m^e mod n = %v^%v mod %v = %v", m, rsa.E, rsa.N, c)

	// Decryption: m' = c^d mod n
	mPrime := new(big.Int).Exp(c, rsa.D, rsa.N)
	t.Logf("Decryption: m' = c^d mod n = %v^%v mod %v = %v", c, rsa.D, rsa.N, mPrime)

	if m.Cmp(mPrime) != 0 {
		t.Errorf("Decryption failed: expected %v, got %v", m, mPrime)
	}

	t.Log("Success: Original message recovered!")
}

func TestSimpleRSAInvalidPrimes(t *testing.T) {
	// Test with non-prime numbers
	_, err := crypto.NewSimpleRSA(10, 53)
	if err == nil {
		t.Error("Should reject non-prime p")
	}

	_, err = crypto.NewSimpleRSA(61, 10)
	if err == nil {
		t.Error("Should reject non-prime q")
	}

	// Test with same primes
	_, err = crypto.NewSimpleRSA(53, 53)
	if err == nil {
		t.Error("Should reject p == q")
	}
}

func TestModularExponentiation(t *testing.T) {
	// Test the square-and-multiply algorithm
	// Verify: 4^13 mod 497 = 445
	base := big.NewInt(4)
	exp := big.NewInt(13)
	mod := big.NewInt(497)

	result := crypto.ModularExponentiation(base, exp, mod)
	expected := big.NewInt(445)

	if result.Cmp(expected) != 0 {
		t.Errorf("ModularExponentiation(4, 13, 497) = %v, expected 445", result)
	}
}

func TestExtendedGCD(t *testing.T) {
	// Test: gcd(240, 46) = 2
	// Should find x, y such that: 240x + 46y = 2
	a := big.NewInt(240)
	b := big.NewInt(46)

	gcd, x, y := crypto.ExtendedGCD(a, b)

	if gcd.Cmp(big.NewInt(2)) != 0 {
		t.Errorf("GCD(240, 46) = %v, expected 2", gcd)
	}

	// Verify: ax + by = gcd
	ax := new(big.Int).Mul(a, x)
	by := new(big.Int).Mul(b, y)
	sum := new(big.Int).Add(ax, by)

	if sum.Cmp(gcd) != 0 {
		t.Errorf("ax + by = %v, expected %v", sum, gcd)
	}

	t.Logf("GCD(%v, %v) = %v", a, b, gcd)
	t.Logf("Bezout coefficients: x=%v, y=%v", x, y)
	t.Logf("Verification: %v*%v + %v*%v = %v", a, x, b, y, sum)
}

func TestEncryptToBytes(t *testing.T) {
	rsa, err := crypto.GetInstance()
	if err != nil {
		t.Fatalf("Failed to get instance: %v", err)
	}

	testCases := []string{
		"Hello",
		"John Doe",
		"Test Customer",
	}

	for _, original := range testCases {
		encrypted, err := rsa.EncryptToBytes(original)
		if err != nil {
			t.Fatalf("Failed to encrypt '%s': %v", original, err)
		}

		decrypted, err := rsa.DecryptFromBytes(encrypted)
		if err != nil {
			t.Fatalf("Failed to decrypt '%s': %v", original, err)
		}

		if decrypted != original {
			t.Errorf("Expected '%s', got '%s'", original, decrypted)
		}
	}
}

func TestDifferentPrimePairs(t *testing.T) {
	// Test with various prime pairs
	primePairs := [][2]int64{
		{11, 13},   // Small primes, n=143
		{17, 19},   // n=323
		{61, 53},   // n=3233
		{101, 103}, // Larger primes, n=10403
	}

	for _, pair := range primePairs {
		t.Run(fmt.Sprintf("p=%d,q=%d", pair[0], pair[1]), func(t *testing.T) {
			rsa, err := crypto.NewSimpleRSA(pair[0], pair[1])
			if err != nil {
				t.Fatalf("Failed to create RSA with p=%d, q=%d: %v", pair[0], pair[1], err)
			}

			// Test encryption/decryption
			original := byte(65)
			encrypted := rsa.EncryptByte(original)
			decrypted := rsa.DecryptByte(encrypted)

			if decrypted != original {
				t.Errorf("Decryption failed with p=%d, q=%d", pair[0], pair[1])
			}

			t.Logf("n=%v, e=%v, d=%v", rsa.N, rsa.E, rsa.D)
		})
	}
}
