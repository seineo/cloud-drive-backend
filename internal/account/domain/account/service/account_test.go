package service

import (
	"account/domain/account/entity"
	"account/domain/account/repository"
	"account/infrastructure/repo"
	"errors"
	"fmt"
	"go.uber.org/mock/gomock"
	"reflect"
	"testing"
)

func Test_accountService_NewAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	var createErr = fmt.Errorf("create error") // 测试用的错误
	var normalAccount = entity.NewAccountWithID(0, "123@456.com", "seineo", "123456")

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
			targetEmailErr:  repo.RecordNotFoundError,
			targetCreateAcc: normalAccount,
			targetCreateErr: nil,
			wantAcc:         normalAccount,
			wantErr:         nil,
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
			targetCreateErr: createErr,
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
			targetEmailErr:  repo.RecordNotFoundError,
			targetCreateAcc: nil,
			targetCreateErr: createErr,
			wantAcc:         nil,
			wantErr:         createErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 构建使用mock仓储的服务
			mockRepo := repository.NewMockAccountRepo(ctrl)
			mockRepo.EXPECT().GetByEmail(tt.args.email).Return(nil, tt.targetEmailErr)
			mockRepo.EXPECT().Create(gomock.Any()).Return(tt.targetCreateAcc, tt.targetCreateErr).AnyTimes()
			svc := NewAccountService(mockRepo, entity.AccountFactoryConfig{
				NicknameRegex: "^[a-zA-Z_][a-zA-Z0-9_-]{0,38}$",
				PasswordRegex: "^[A-Za-z0-9]{6,38}$",
			})

			got, err := svc.NewAccount(tt.args.email, tt.args.nickname, tt.args.password)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantAcc) {
				t.Errorf("NewAccount() got = %v, want %v", got, tt.wantAcc)
			}
		})
	}
}

//TODO 增加其余接口的单测
