package model

import (
	"apcs_refactored/config"
	"reflect"
	"testing"
)

func TestSelectSlotList(t *testing.T) {
	config.InitConfig()
	InitDB()
	defer CloseDB()

	tests := []struct {
		name    string
		want    []Slot
		wantErr bool
	}{
		{
			name: "Test case 1: Find all slots",
			//want:
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectSlotList()
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectSlotList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SelectSlotList() got = %v, want %v", got, tt.want)
			}
		})
	}
}
