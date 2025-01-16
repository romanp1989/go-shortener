package config

import (
	"reflect"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name    string
		want    *ConfigENV
		wantErr bool
	}{
		{
			name:    "ParseFlags_Success",
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseFlags()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			testConfig := ConfigENV{}

			if reflect.TypeOf(config) != reflect.TypeOf(&testConfig) {
				t.Errorf("ParseFlags() got type = %v, want type %v", reflect.TypeOf(config), reflect.TypeOf(&testConfig))
			}
		})
	}
}

func Test_createTLSCertificate(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "createTLSCertificate_Success",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTLSConfig := TLSConfig{}
			tlsConfig, err := createTLSCertificate()
			if (err != nil) != tt.wantErr {
				t.Errorf("createTLSCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(tlsConfig) != reflect.TypeOf(&testTLSConfig) {
				t.Errorf("createTLSCertificate() got type = %v, want type %v", reflect.TypeOf(tlsConfig), reflect.TypeOf(&testTLSConfig))
			}
		})
	}
}

func Test_saveTLSParamsToFile(t *testing.T) {
	type args struct {
		tlsConf *TLSConfig
		cfg     ConfigENV
	}

	tlsConfig, _ := createTLSCertificate()
	config := ConfigENV{
		SecretKey: "secret_key",
		HTTPS: HTTPSConfig{
			Enable: true,
			Key:    "dsfsfddsf",
			Pem:    "dsfsdfsdf",
		},
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "saveTLSParamsToFile_Success",
			args: args{
				tlsConf: tlsConfig,
				cfg:     config,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := saveTLSParamsToFile(tt.args.tlsConf, tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("saveTLSParamsToFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
