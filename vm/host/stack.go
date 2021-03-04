package host

import (
	"strconv"

	"github.com/iost-official/go-iost/v3/core/contract"
)

// Stack is call stack of this host
type Stack struct {
	h *Host
}

// NewStack new stack
func NewStack(h *Host) Stack {
	return Stack{h: h}
}

// Caller ...
type Caller struct {
	Name      string `json:"name"`
	IsAccount bool   `json:"is_account"`
}

// InitStack init a new call stack
func (s *Stack) InitStack(publisher string) {
	s.h.PushCtx()

	s.h.ctx.Set("stack0", "direct_call")
	s.h.ctx.Set("stack_height", 1) // record stack trace
	s.h.ctx.Set("caller", Caller{
		Name:      publisher,
		IsAccount: true,
	})
}

// PushStack push a new call stack
func (s *Stack) PushStack(cont, api string, withAuth bool) (reenter bool, cost contract.Cost) {
	s.h.PushCtx()

	record := cont + "-" + api
	height := s.h.ctx.Value("stack_height").(int)

	for i := 0; i < height; i++ {
		key := "stack" + strconv.Itoa(i)
		if s.h.ctx.Value(key).(string) == record {
			return true, contract.Cost0()
		}
	}

	key := "stack" + strconv.Itoa(height)
	s.h.ctx.Set(key, record)
	s.h.ctx.Set("stack_height", height+1)

	callerName := ""
	if s.h.ctx.Value("contract_name") != nil {
		callerName = s.h.ctx.Value("contract_name").(string)
	}
	s.h.Context().Set("caller", Caller{
		Name:      callerName,
		IsAccount: false,
	})

	// handle withAuth
	if withAuth {
		authList := s.h.ctx.Value("auth_contract_list").(map[string]int)

		if s.h.IsFork3_0_10 {
			newAuthList := make(map[string]int, len(authList))
			for k, v := range authList {
				newAuthList[k] = v
			}
			newAuthList[s.h.ctx.Value("contract_name").(string)] = 1
			s.h.ctx.Set("auth_contract_list", newAuthList)
		} else {
			authList[s.h.ctx.Value("contract_name").(string)] = 1
			s.h.ctx.Set("auth_contract_list", authList)
		}
	}

	return false, CommonOpCost(height)
}

// PopStack pop a call stack
func (s *Stack) PopStack() {
	s.h.PopCtx()
}

// SetStackInfo set contract info
func (s *Stack) SetStackInfo(cont, api string) {
	s.h.Context().Set("contract_name", cont)
	s.h.Context().Set("abi_name", api)
}

// StackHeight get current stack height
func (s *Stack) StackHeight() int {
	return s.h.ctx.Value("stack_height").(int)
}

// Caller get caller
func (s *Stack) Caller() Caller {
	return s.h.ctx.Value("caller").(Caller)
}
