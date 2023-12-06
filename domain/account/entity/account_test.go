package entity

import (
	"testing"
)

func TestNewFactory(t *testing.T) {
	type args struct {
		fc AccountFactoryConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal case",
			args: args{fc: AccountFactoryConfig{
				NicknameRegex: "[a-z]",
				PasswordRegex: "[0-9]",
			}},
			wantErr: false,
		},
		{
			name: "empty regex",
			args: args{fc: AccountFactoryConfig{
				NicknameRegex: "",
				PasswordRegex: "",
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAccountFactory(tt.args.fc)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAccountFactory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestFactory_NewAccount(t *testing.T) {
	type args struct {
		email    string
		nickname string
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal case",
			args: args{
				email:    "1@2.com",
				nickname: "a123",
				password: "123456",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			args: args{
				email:    "1.com",
				nickname: "a123",
				password: "123456",
			},
			wantErr: true,
		},
		{
			name: "invalid nickname",
			args: args{
				email:    "1@2.com",
				nickname: "#123a",
				password: "123456",
			},
			wantErr: true,
		},
		{
			name: "invalid password",
			args: args{
				email:    "1@qq.com",
				nickname: "123a",
				password: "123",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &AccountFactory{
				fc: AccountFactoryConfig{
					NicknameRegex: "^[a-zA-Z_][a-zA-Z0-9_-]{0,38}$",
					PasswordRegex: "^[A-Za-z0-9]{6,38}$",
				},
			}
			_, err := f.NewAccount(tt.args.email, tt.args.nickname, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
