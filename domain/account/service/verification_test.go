package service

import (
	"CloudDrive/domain/account/entity"
	"CloudDrive/domain/account/repository"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func Test_verificationService_SendAuthCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockCodeRepository(ctrl)

	actualFactory, err := entity.NewCodeFactory(5, 1)
	assert.NoError(t, err)
	expectedFactory, err := entity.NewCodeFactory(5, 1)
	assert.NoError(t, err)

	type args struct {
		email      string
		expiration time.Duration
	}
	tests := []struct {
		name             string
		args             args
		targetSetCodeErr error
		want             string
		wantErr          bool
	}{
		{
			name:             "normal case",
			args:             args{email: "123@test.com", expiration: 10 * time.Minute},
			targetSetCodeErr: nil,
			want:             expectedFactory.NewVerificationCode().Get(),
			wantErr:          false,
		},
		{
			name:             "db error",
			args:             args{email: "123@test.com", expiration: 10 * time.Minute},
			targetSetCodeErr: fmt.Errorf("db error"),
			want:             "",
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &verificationService{
				codeRepo: mockRepo,
				factory:  actualFactory,
			}
			mockRepo.EXPECT().SetCode(gomock.Any(), gomock.Any(), gomock.Any()).Return(tt.targetSetCodeErr)
			got, err := v.SendAuthCode(tt.args.email, tt.args.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendAuthCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SendAuthCode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
