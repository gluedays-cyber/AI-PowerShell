package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"pi/pkg/platform/windows"
)

var cryptoKey = []byte("StyledMDSecretKeyForAPIEncryption")[:32]

// --- 설정 데이터 구조체 ---
type encryptedConfig struct {
	GoogleAPIKey    string `json:"google_api_key"`
	Model1          string `json:"model_1"`
	Model2          string `json:"model_2"`
	UserInstruction string `json:"user_instruction"`
	PythonPath      string `json:"python_path"`
}

type DecryptedConfig struct {
	GoogleAPIKey string
	Model1       string
	Model2       string
	PythonPath   string
}

// --- 복호화 로직 ---
func decrypt(cryptoText string) (string, error) {
	if cryptoText == "" {
		return "", nil
	}
	dpapiCipher, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}

	ciphertext, err := windows.CryptUnprotectData(dpapiCipher)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(cryptoKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext length too short")
	}
	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// --- 설정 파일 불러오기 ---
func LoadConfig() (*DecryptedConfig, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		appData = filepath.Join(home, "AppData", "Roaming")
	}
	path := filepath.Join(appData, ".tea", ".apikeys.json")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &DecryptedConfig{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var encCfg encryptedConfig
	if err := json.Unmarshal(data, &encCfg); err != nil {
		return nil, err
	}

	decKey, err := decrypt(encCfg.GoogleAPIKey)
	if err != nil {
		return nil, err
	}
	decM1, err := decrypt(encCfg.Model1)
	if err != nil {
		return nil, err
	}
	decM2, err := decrypt(encCfg.Model2)
	if err != nil {
		Log.Warn("Model2 복호화 실패 (빈 값 처리됨): %v\n", err)
		decM2 = ""
	}

	decPyPath, err := decrypt(encCfg.PythonPath)
	if err != nil {
		Log.Warn("PythonPath 복호화 실패: %v\n", err)
		decPyPath = ""
	}
	if decPyPath == "" {
		decPyPath = encCfg.PythonPath
	}

	return &DecryptedConfig{
		GoogleAPIKey: decKey,
		Model1:       decM1,
		Model2:       decM2,
		PythonPath:   decPyPath,
	}, nil
}
