package service

import (
	"account/domain/account/entity"
	"account/domain/account/repository"
	"common/eventbus"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func Test_verificationService_GenerateAuthCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
		targetPublishErr error
		want             string
		wantErr          bool
	}{
		{
			name:             "normal case",
			args:             args{email: "123@test.com", expiration: 10 * time.Minute},
			targetSetCodeErr: nil,
			targetPublishErr: nil,
			want:             expectedFactory.NewVerificationCode("123@test.com").Get(),
			wantErr:          false,
		},
		{
			name:             "db error",
			args:             args{email: "123@test.com", expiration: 10 * time.Minute},
			targetSetCodeErr: fmt.Errorf("db error"),
			targetPublishErr: nil,
			want:             "",
			wantErr:          true,
		},
		{
			name:             "producer error",
			args:             args{email: "123@test.com", expiration: 10 * time.Minute},
			targetSetCodeErr: nil,
			targetPublishErr: errors.New("producer error"),
			want:             "",
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := repository.NewMockCodeRepository(ctrl)
			mockProducer := eventbus.NewMockProducer(ctrl)
			v := &verificationService{
				codeRepo:      mockRepo,
				factory:       actualFactory,
				eventProducer: mockProducer,
			}
			mockRepo.EXPECT().SetCode(gomock.Any(), gomock.Any(), gomock.Any()).Return(tt.targetSetCodeErr)
			mockProducer.EXPECT().Publish("account", gomock.Any()).Return(tt.targetPublishErr).AnyTimes()
			got, err := v.GenerateAuthCode(tt.args.email, tt.args.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateAuthCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateAuthCode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
