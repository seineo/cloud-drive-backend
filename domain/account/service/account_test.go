package service

import (
	"CloudDrive/domain/account/entity"
	"CloudDrive/domain/account/repository"
	"fmt"
	"go.uber.org/mock/gomock"
	"reflect"
	"testing"
)

func Test_accountService_NewAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		email    string
		nickname string
		password string
	}
	tests := []struct {
		name            string
		args            args
		targetEmailErr  error
		targetCreateAcc *entity.Account
		targetCreateErr error
		wantAcc         *entity.Account
		wantErr         error
	}{
		{
			name: "normal case",
			args: args{
				email:    "123@456.com",
				nickname: "seineo",
				password: "123456",
			},
			targetEmailErr: nil,
			targetCreateAcc: &entity.Account{
				ID:       0,
				Email:    "123@456.com",
				Nickname: "seineo",
				Password: "123456",
			},
			targetCreateErr: nil,
			wantAcc: &entity.Account{
				ID:       0,
				Email:    "123@456.com",
				Nickname: "seineo",
				Password: "123456",
			},
			wantErr: nil,
		},
		{
			name: "email used",
			args: args{
				email:    "123@456.com",
				nickname: "seineo",
				password: "123456",
			},
			targetEmailErr:  EmailUsedError,
			targetCreateAcc: nil,
			targetCreateErr: fmt.Errorf("create error"),
			wantAcc:         nil,
			wantErr:         EmailUsedError,
		},
		{
			name: "database create error",
			args: args{
				email:    "123@456.com",
				nickname: "seineo",
				password: "123456",
			},
			targetEmailErr:  nil,
			targetCreateAcc: nil,
			targetCreateErr: fmt.Errorf("create error"),
			wantAcc:         nil,
			wantErr:         fmt.Errorf("create error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 构建使用mock仓储的服务
			mockRepo := repository.NewMockAccountRepo(ctrl)
			mockRepo.EXPECT().GetByEmail(tt.args.email).Return(nil, tt.targetEmailErr)
			mockRepo.EXPECT().Create(gomock.Any()).Return(tt.targetCreateAcc, tt.targetCreateErr).AnyTimes()
			svc := NewAccountService(mockRepo, entity.FactoryConfig{
				NicknameRegex: "^[a-zA-Z_][a-zA-Z0-9_-]{0,38}$",
				PasswordRegex: "^[A-Za-z0-9]{6,38}$",
			})

			got, err := svc.NewAccount(tt.args.email, tt.args.nickname, tt.args.password)
			if err != tt.wantErr {
				t.Errorf("NewAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantAcc) {
				t.Errorf("NewAccount() got = %v, want %v", got, tt.wantAcc)
			}
		})
	}
}
