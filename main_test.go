package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha1"
	"golang.org/x/crypto/pbkdf2"
	"testing"
)

func TestPKCS7Unpad(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		blockSize int
		want      []byte
		wantErr   bool
	}{
		{
			name:      "valid padding",
			data:      []byte{1, 2, 3, 4, 5, 3, 3, 3},
			blockSize: 8,
			want:      []byte{1, 2, 3, 4, 5},
			wantErr:   false,
		},
		{
			name:      "valid full block padding",
			data:      []byte{1, 2, 3, 4, 4, 4, 4, 4},
			blockSize: 4,
			want:      []byte{1, 2, 3, 4},
			wantErr:   false,
		},
		{
			name:      "empty data",
			data:      []byte{},
			blockSize: 8,
			wantErr:   true,
		},
		{
			name:      "invalid block size multiple",
			data:      []byte{1, 2, 3},
			blockSize: 8,
			wantErr:   true,
		},
		{
			name:      "invalid padding length 0",
			data:      []byte{1, 2, 3, 4, 5, 6, 7, 0},
			blockSize: 8,
			wantErr:   true,
		},
		{
			name:      "invalid padding length too large",
			data:      []byte{1, 2, 3, 4, 5, 6, 7, 9},
			blockSize: 8,
			wantErr:   true,
		},
		{
			name:      "invalid padding bytes",
			data:      []byte{1, 2, 3, 4, 5, 3, 2, 3},
			blockSize: 8,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pkcs7Unpad(tt.data, tt.blockSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("pkcs7Unpad() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if string(got) != string(tt.want) {
					t.Errorf("pkcs7Unpad() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDecryptCBC(t *testing.T) {
	password := []byte("secret")
	plaintext := "hello world"
	blockSize := aes.BlockSize

	// Derive key manually to create ciphertext
	hasher := md5.New()
	hasher.Write(password)
	key := hasher.Sum(nil)

	block, _ := aes.NewCipher(key)
	iv := make([]byte, blockSize)
	for i := range iv {
		iv[i] = byte(i)
	}

	// PKCS7 padding
	padLen := blockSize - (len(plaintext) % blockSize)
	padded := append([]byte(plaintext), make([]byte, padLen)...)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}

	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	fullData := append(iv, ciphertext...)

	tests := []struct {
		name     string
		data     []byte
		password []byte
		want     string
		wantErr  bool
	}{
		{
			name:     "valid decryption",
			data:     fullData,
			password: password,
			want:     plaintext,
			wantErr:  false,
		},
		{
			name:     "invalid password",
			data:     fullData,
			password: []byte("wrong"),
			wantErr:  true,
		},
		{
			name:     "data too short",
			data:     make([]byte, blockSize-1),
			password: password,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decryptCBC(tt.data, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("decryptCBC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("decryptCBC() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecryptGCM(t *testing.T) {
	password := []byte("secret")
	plaintext := "hello gcm"
	salt := make([]byte, 16)
	for i := range salt {
		salt[i] = byte(i)
	}
	nonce := make([]byte, 16)
	for i := range nonce {
		nonce[i] = byte(i + 16)
	}

	key := pbkdf2.Key(password, salt, 1000, 32, sha1.New)
	block, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCMWithNonceSize(block, len(nonce))
	ciphertextWithTag := aesgcm.Seal(nil, nonce, []byte(plaintext), salt)

	fullData := append(salt, append(nonce, ciphertextWithTag...)...)

	tests := []struct {
		name     string
		data     []byte
		password []byte
		want     string
		wantErr  bool
	}{
		{
			name:     "valid decryption",
			data:     fullData,
			password: password,
			want:     plaintext,
			wantErr:  false,
		},
		{
			name:     "invalid password",
			data:     fullData,
			password: []byte("wrong"),
			wantErr:  true,
		},
		{
			name:     "data too short",
			data:     make([]byte, 47),
			password: password,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decryptGCM(tt.data, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("decryptGCM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("decryptGCM() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	password := []byte("secret")
	plaintext := "hello world"

	// Create CBC data
	hasher := md5.New()
	hasher.Write(password)
	keyCBC := hasher.Sum(nil)
	blockCBC, _ := aes.NewCipher(keyCBC)
	iv := make([]byte, aes.BlockSize)
	padLen := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	padded := append([]byte(plaintext), make([]byte, padLen)...)
	for i := len(plaintext); i < len(padded); i++ { padded[i] = byte(padLen) }
	ciphertextCBC := make([]byte, len(padded))
	cipher.NewCBCEncrypter(blockCBC, iv).CryptBlocks(ciphertextCBC, padded)
	fullDataCBC := append(iv, ciphertextCBC...)

	tests := []struct {
		name     string
		mode     string
		data     []byte
		password []byte
		want     string
		wantErr  bool
	}{
		{
			name:     "CBC mode",
			mode:     "CBC",
			data:     fullDataCBC,
			password: password,
			want:     plaintext,
			wantErr:  false,
		},
		{
			name:     "empty mode (default CBC)",
			mode:     "",
			data:     fullDataCBC,
			password: password,
			want:     plaintext,
			wantErr:  false,
		},
		{
			name:     "unknown mode",
			mode:     "XTS",
			data:     fullDataCBC,
			password: password,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decrypt(tt.mode, tt.data, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("decrypt() got = %v, want %v", got, tt.want)
			}
		})
	}
}
