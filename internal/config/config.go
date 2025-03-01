package config

import (
	"bytes"
	"cmp"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"log"
	"math/big"
	"net"
	"os"
	"path"
	"time"
)

// Options console flag configuration struct
//var Options struct {
//	FlagRunPort     string
//	FlagShortURL    string
//	FlagLogLevel    string
//	FlagFileStorage string
//	FlagDatabaseDsn string
//	FlagSecretKey   string
//	FlagEnableHTTPS bool
//}

// ConfigENV env configuration params
type ConfigENV struct {
	ServerAddress     string `env:"SERVER_ADDRESS" json:"server_address,omitempty"`
	GRPCServerAddress string `env:"GRPC_SERVER_ADDRESS" json:"grpc_server_address,omitempty"`
	BaseURL           string `env:"BASE_URL" json:"base_url,omitempty"`
	LogLevel          string `env:"LOG_LEVEL"`
	FileStorage       string `env:"FILE_STORAGE_PATH" json:"file_storage_path,omitempty"`
	DatabaseDsn       string `env:"DATABASE_DSN" json:"database_dsn,omitempty"`
	SecretKey         string `env:"SECRET_KEY"`
	TrustedSubnet     string `env:"TRUSTED_SUBNET"`
	HTTPS             HTTPSConfig
}

// HTTPSConfig https config struct with key, pem
type HTTPSConfig struct {
	Enable bool `env:"ENABLE_HTTPS" json:"enable_https,omitempty"`
	Key    string
	Pem    string
}

// TLSConfig config for tls key, pem
type TLSConfig struct {
	Key *rsa.PrivateKey
	Pem []byte
}

// ParseFlags function for parse application run flags
func ParseFlags() (*ConfigENV, error) {
	var cfg ConfigENV
	var configPath string

	pwd, _ := os.Getwd()
	cfg.HTTPS.Key = path.Join(pwd, "/server.key")
	cfg.HTTPS.Pem = path.Join(pwd, "/server.pem")

	flag.StringVar(&cfg.ServerAddress, "a", ":8080", "port to run server")
	flag.StringVar(&cfg.GRPCServerAddress, "g", ":3200", "port to run grpc server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "address to run server")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.FileStorage, "f", "/tmp/shortener.txt", "file storage")
	flag.StringVar(&cfg.DatabaseDsn, "d", "", "Database DSN")
	flag.StringVar(&cfg.SecretKey, "sk", "sdfsdfsadfsdafasfsaf", "Secret key")
	flag.BoolVar(&cfg.HTTPS.Enable, "s", false, "Enable HTTPS")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "Trusted subnet")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		log.Printf("Ошибка при парсинге переменных окружения %s", err.Error())
		return nil, err
	}

	if os.Getenv("CONFIG") != "" {
		configPath = os.Getenv("CONFIG")
	}

	if configPath != "" {
		fCfg := ConfigENV{
			HTTPS: HTTPSConfig{},
		}
		confFromFile, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("can't read config from file: %w", err)
		}

		if err := json.Unmarshal(confFromFile, &cfg); err != nil {
			return nil, fmt.Errorf("can't parse config from file: %w", err)
		}

		cfg.ServerAddress = cmp.Or(cfg.ServerAddress, fCfg.ServerAddress)
		cfg.GRPCServerAddress = cmp.Or(cfg.GRPCServerAddress, fCfg.GRPCServerAddress)
		cfg.BaseURL = cmp.Or(cfg.BaseURL, fCfg.BaseURL)
		cfg.FileStorage = cmp.Or(cfg.FileStorage, fCfg.FileStorage)
		cfg.DatabaseDsn = cmp.Or(cfg.DatabaseDsn, fCfg.DatabaseDsn)
		cfg.HTTPS.Enable = cmp.Or(cfg.HTTPS.Enable, fCfg.HTTPS.Enable)
	}

	if cfg.HTTPS.Enable {
		tlsConf, err := createTLSCertificate()
		if err != nil {
			return nil, err
		}

		err = saveTLSParamsToFile(tlsConf, cfg)
		if err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}

func createTLSCertificate() (*TLSConfig, error) {
	cert := &x509.Certificate{
		// указываем уникальный номер сертификата
		SerialNumber: big.NewInt(1),
		// заполняем базовую информацию о владельце сертификата
		Subject: pkix.Name{
			Organization: []string{"Shortener Company"},
			Country:      []string{"RU"},
			CommonName:   "localhost",
		},
		// разрешаем использование сертификата для 127.0.0.1 и ::1
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		// сертификат верен, начиная со времени создания
		NotBefore: time.Now(),
		// время жизни сертификата — 10 лет
		NotAfter:     time.Now().AddDate(0, 3, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		// устанавливаем использование ключа для цифровой подписи,
		// а также клиентской и серверной авторизации
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	return &TLSConfig{
		Key: privateKey,
		Pem: certBytes,
	}, nil
}

func saveTLSParamsToFile(tlsConf *TLSConfig, cfg ConfigENV) error {
	var (
		certPEM       bytes.Buffer
		privateKeyPEM bytes.Buffer
		file          *os.File
	)
	err := pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: tlsConf.Pem,
	})

	if err != nil {
		return fmt.Errorf("pem encode error: %w", err)
	}

	file, _ = os.Create(cfg.HTTPS.Pem)
	defer file.Close()

	_, err = certPEM.WriteTo(file)
	if err != nil {
		return fmt.Errorf("write to file error: %w", err)
	}

	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(tlsConf.Key),
	})
	if err != nil {
		return fmt.Errorf("private key encode error: %w", err)
	}

	file, _ = os.Create(cfg.HTTPS.Key)
	defer file.Close()

	_, err = privateKeyPEM.WriteTo(file)
	if err != nil {
		return fmt.Errorf("write to file error: %w", err)
	}

	return nil
}
