package utils

import "testing"

func TestGetValueOrDefault(t *testing.T) {
	type args struct {
		v string
	}
	type testCase struct {
		name string
		args args
		want string
	}
	tests := []testCase{
		{
			name: "TestGetValueOrDefault #1",
			args: args{
				v: "",
			},
			want: "N/A",
		},
		{
			name: "TestGetValueOrDefault #2",
			args: args{
				v: "test_arg",
			},
			want: "test_arg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetValueOrDefault(tt.args.v); got != tt.want {
				t.Errorf("GetValueOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}
