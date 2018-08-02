package new_vm

import (
	"testing"

	"context"

	. "github.com/golang/mock/gomock"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
)

func Init(t *testing.T) (*Monitor, *MockVM, *database.MockIMultiValue) {
	mc := NewController(t)
	defer mc.Finish()
	vm := NewMockVM(mc)
	db := database.NewMockIMultiValue(mc)
	pm := NewMonitor(db, 100)
	return pm, vm, db
}

func TestMonitor_Call(t *testing.T) {
	monitor, vm, db := Init(t)

	ctx := context.Background()
	vm.EXPECT()
	db.EXPECT().Get(Any(), Any()).DoAndReturn(func(table string, key string) (string, error) {
		return "", nil
	})

	monitor.Call(ctx, "contract", "api", "1")
}
