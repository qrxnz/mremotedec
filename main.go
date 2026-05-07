package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/csv"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/crypto/pbkdf2"
)

type Connections struct {
	XMLName            xml.Name `xml:"Connections"`
	BlockCipherMode    string   `xml:"BlockCipherMode,attr"`
	FullFileEncryption string   `xml:"FullFileEncryption,attr"`
	Nodes              []Node   `xml:"Node"`
}

type Node struct {
	Name     string `xml:"Name,attr"`
	Username string `xml:"Username,attr"`
	Hostname string `xml:"Hostname,attr"`
	Password string `xml:"Password,attr"`
	Type     string `xml:"Type,attr"`
	Nodes    []Node `xml:"Node"`
}

var (
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1).
			MarginBottom(1)

	nameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62")).
			Width(10)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	passwordStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229"))
)

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	if len(data)%blockSize != 0 {
		return nil, fmt.Errorf("data is not a multiple of block size")
	}
	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > blockSize {
		return nil, fmt.Errorf("invalid padding length")
	}
	for i := 0; i < padLen; i++ {
		if data[len(data)-1-i] != byte(padLen) {
			return nil, fmt.Errorf("invalid padding byte")
		}
	}
	return data[:len(data)-padLen], nil
}

func decryptCBC(data []byte, password []byte) (string, error) {
	if len(data) < aes.BlockSize {
		return "", fmt.Errorf("data too short")
	}
	iv := data[:aes.BlockSize]
	ciphertext := data[aes.BlockSize:]

	hasher := md5.New()
	hasher.Write(password)
	key := hasher.Sum(nil)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	unpadded, err := pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return "", err
	}

	return string(unpadded), nil
}

func decryptGCM(data []byte, password []byte) (string, error) {
	if len(data) < 48 { // 16 salt + 16 nonce + 16 tag
		return "", fmt.Errorf("data too short for GCM")
	}
	salt := data[:16]
	nonce := data[16:32]
	ciphertextWithTag := data[32:]

	key := pbkdf2.Key(password, salt, 1000, 32, sha1.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCMWithNonceSize(block, len(nonce))
	if err != nil {
		return "", err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertextWithTag, salt)
	if err != nil {
		return "", fmt.Errorf("MAC tag not valid: %v", err)
	}

	return string(plaintext), nil
}

func decrypt(mode string, data []byte, password []byte) (string, error) {
	switch mode {
	case "CBC", "":
		return decryptCBC(data, password)
	case "GCM":
		return decryptGCM(data, password)
	default:
		return "", fmt.Errorf("unknown mode %s", mode)
	}
}

func renderConnection(node Node, decryptedPassword string) string {
	lines := []string{
		nameStyle.Render(node.Name),
		labelStyle.Render("Hostname:") + valueStyle.Render(node.Hostname),
		labelStyle.Render("Username:") + valueStyle.Render(node.Username),
		labelStyle.Render("Password:") + passwordStyle.Render(decryptedPassword),
	}

	return cardStyle.Render(strings.Join(lines, "\n"))
}

func processNode(node Node, mode string, password []byte, writer *csv.Writer, csvFlag bool) {
	if node.Type == "Connection" {
		decryptedPassword := ""
		if node.Password != "" {
			passData, err := base64.StdEncoding.DecodeString(node.Password)
			if err == nil {
				decryptedPassword, _ = decrypt(mode, passData, password)
			}
		}

		if csvFlag {
			writer.Write([]string{node.Name, node.Hostname, node.Username, decryptedPassword})
		} else {
			fmt.Println(renderConnection(node, decryptedPassword))
		}
	}

	for _, child := range node.Nodes {
		processNode(child, mode, password, writer, csvFlag)
	}
}

func main() {
	passwordFlag := flag.String("password", "mR3m", "Optional decryption password")
	flag.StringVar(passwordFlag, "p", "mR3m", "Optional decryption password (shorthand)")
	csvFlag := flag.Bool("csv", false, "Output CSV format")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: mremotedec [options] <config_file>")
		return
	}

	configFile := flag.Arg(0)
	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	var conns Connections
	err = xml.Unmarshal(data, &conns)
	if err != nil {
		fmt.Printf("Error parsing XML: %v\n", err)
		os.Exit(1)
	}

	mode := conns.BlockCipherMode
	if mode == "" {
		mode = "CBC"
	}

	password := []byte(*passwordFlag)

	if conns.FullFileEncryption == "true" {
		s := string(data)
		start := strings.Index(s, ">") + 1
		end := strings.LastIndex(s, "</")
		if start > 0 && end > start {
			encryptedB64 := strings.TrimSpace(s[start:end])
			encryptedData, err := base64.StdEncoding.DecodeString(encryptedB64)
			if err != nil {
				fmt.Printf("Error decoding base64: %v\n", err)
				os.Exit(1)
			}
			decryptedContent, err := decrypt(mode, encryptedData, password)
			if err != nil {
				fmt.Printf("Error decrypting file: %v\n", err)
				os.Exit(1)
			}
			// Wrap it back in a dummy tag to parse nodes
			decryptedContent = "<Connections>" + decryptedContent + "</Connections>"
			err = xml.Unmarshal([]byte(decryptedContent), &conns)
			if err != nil {
				fmt.Printf("Error parsing decrypted XML: %v\n", err)
				os.Exit(1)
			}
		}
	}

	var writer *csv.Writer
	if *csvFlag {
		writer = csv.NewWriter(os.Stdout)
		writer.Write([]string{"Name", "Hostname", "Username", "Password"})
	}

	for _, node := range conns.Nodes {
		processNode(node, mode, password, writer, *csvFlag)
	}

	if *csvFlag {
		writer.Flush()
	}
}
