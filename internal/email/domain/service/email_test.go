package service

import (
	"common/config"
	"email/domain/infrastructure"
	"go.uber.org/mock/gomock"
	"testing"
)

func Test_emailService_SendVerificationCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSender := infrastructure.NewMockEmailSender(ctrl)

	e := &emailService{
		senderEmail: map[string]string{"code": "noreply@test.com"},
		emailSender: mockSender,
		configs: &config.Config{
			ProjectName: "cloud",
			ProjectURL:  "localhost",
		},
	}

	type args struct {
		email string
		code  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal case",
			args: args{
				email: "123@test.com",
				code:  "123456",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := e.SendVerificationCode(tt.args.email, tt.args.code); (err != nil) != tt.wantErr {
				t.Errorf("SendVerificationCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
