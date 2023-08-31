package model

import (
	"apcs_refactored/config"
	"reflect"
	"testing"
)

func TestSelectItemById(t *testing.T) {
	config.InitConfig()
	InitDB()
	defer CloseDB()

	type args struct {
		itemId int64
	}
	tests := []struct {
		name    string
		args    args
		want    Item
		wantErr bool
	}{
		{
			name: "test case 1",
			args: args{itemId: 1},
			want: Item{
				ItemId: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectItemById(tt.args.itemId)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectItemById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.ItemId, tt.want.ItemId) {
				t.Errorf("SelectItemById() got = %v, want %v", got, tt.want)
			}
		})
	}
}
