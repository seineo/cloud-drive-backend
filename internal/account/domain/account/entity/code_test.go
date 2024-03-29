package entity

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestCodeFactory_NewVerificationCode(t *testing.T) {
	tests := []struct {
		name      string
		digits    uint
		r1        *rand.Rand
		r2        *rand.Rand
		wantEqual bool
	}{
		{
			name:      "normal case",
			digits:    5,
			r1:        rand.New(rand.NewSource(1)),
			r2:        rand.New(rand.NewSource(1)),
			wantEqual: true,
		},
		{
			name:      "longer digits",
			digits:    10,
			r1:        rand.New(rand.NewSource(1)),
			r2:        rand.New(rand.NewSource(1)),
			wantEqual: true,
		},
		{
			name:      "different seed",
			digits:    5,
			r1:        rand.New(rand.NewSource(time.Now().UnixNano())),
			r2:        rand.New(rand.NewSource(time.Now().UnixNano())),
			wantEqual: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf1 := &CodeFactory{
				digits: tt.digits,
				r:      tt.r1,
			}
			cf2 := &CodeFactory{
				digits: tt.digits,
				r:      tt.r2,
			}
			codeObj1 := cf1.NewVerificationCode("1@test.com")
			codeObj2 := cf2.NewVerificationCode("1@test.com")
			if tt.wantEqual && codeObj1.Get() != codeObj2.Get() {
				t.Errorf("code1 and code2 not equal")
			}
			if !tt.wantEqual && codeObj1.Get() == codeObj2.Get() {
				t.Errorf("code1 and code2 should not be equal")
			}
			if uint(len(codeObj1.Get())) != tt.digits || uint(len(codeObj2.Get())) != tt.digits {
				t.Errorf("code lenth expected: %v, actual: {code1: %v, code2: %v}",
					tt.digits, len(codeObj1.Get()), len(codeObj2.Get()))
			}
			event1 := codeObj1.GetEvent()
			assert.Equal(t, "codeGenerated", event1.GetName())
			event2 := codeObj2.GetEvent()
			assert.Equal(t, "codeGenerated", event2.GetName())
			// 判断id不同
			assert.NotEqual(t, event1.GetID(), event2.GetID())
		})
	}
}

func TestNewCodeFactory(t *testing.T) {
	type args struct {
		digits uint
		seed   int64
	}
	tests := []struct {
		name    string
		args    args
		want    *CodeFactory
		wantErr bool
	}{
		{
			name: "normal case",
			args: args{
				digits: 5,
				seed:   1,
			},
			want: &CodeFactory{
				digits: 5,
				r:      rand.New(rand.NewSource(1)),
			},
			wantErr: false,
		},
		{
			name: "smaller digits",
			args: args{
				digits: 3,
				seed:   1,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "bigger digits",
			args: args{
				digits: 10,
				seed:   1,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCodeFactory(tt.args.digits, tt.args.seed)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCodeFactory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCodeFactory() got = %v, want %v", got, tt.want)
			}
		})
	}
}
