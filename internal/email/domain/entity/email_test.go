package entity

import (
	"reflect"
	"testing"
)

func TestNewEmail(t *testing.T) {
	type args struct {
		sender     string
		recipients []string
		subject    string
		body       string
	}
	tests := []struct {
		name    string
		args    args
		want    *Email
		wantErr bool
	}{
		{
			name: "normal case",
			args: args{
				sender:     "1@test.com",
				recipients: []string{"2@test.com"},
				subject:    "test",
				body:       "test content",
			},
			want: &Email{
				sender:     "1@test.com",
				recipients: []string{"2@test.com"},
				subject:    "test",
				body:       "test content",
			},
			wantErr: false,
		},
		{
			name: "wrong sender",
			args: args{
				sender:     "",
				recipients: []string{"2@test.com"},
				subject:    "",
				body:       "",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wrong recipients",
			args: args{
				sender:     "1@test.com",
				recipients: nil,
				subject:    "",
				body:       "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEmail(tt.args.sender, tt.args.recipients, tt.args.subject, tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEmail() got = %v, want %v", got, tt.want)
			}
		})
	}
}
