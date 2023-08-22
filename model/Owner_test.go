package model

import (
	"apcs_refactored/config"
	"reflect"
	"testing"
)

func TestSelectOwnerByOwnerInfo(t *testing.T) {
	config.InitConfig()
	InitDB()
	defer CloseDB()

	type args struct {
		info OwnerInfo
	}
	tests := []struct {
		name    string
		args    args
		want    Owner
		wantErr bool
	}{
		{"Test case 1: Find test owner Bob",
			args{
				OwnerInfo{
					OwnerName: "Bob",
					PhoneNum:  "01012345678",
					Address:   "111-222",
				},
			},
			Owner{
				OwnerId:   1,
				OwnerName: "Bob",
				PhoneNum:  "01012345678",
				Address:   "111-222",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectOwnerByOwnerInfo(tt.args.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectOwnerByOwnerInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SelectOwnerByOwnerInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}
