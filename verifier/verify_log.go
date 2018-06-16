package verifier

import "github.com/iost-official/prototype/common"

type VerifyLog struct {
	ThreadCount int
	List        [][]int
	current     []int
}

func NewLog(tc int) VerifyLog {
	return VerifyLog{
		ThreadCount: tc,
		List:        make([][]int, tc),
		current:     make([]int, tc),
	}
}

func (v *VerifyLog) Verify(thread, contractid int) {
	v.List[thread] = append(v.List[thread], contractid)
}

func (v *VerifyLog) Next(thread int) int {
	return v.List[thread][v.current[thread]]
}
func (v *VerifyLog) Encode() []byte {
	list := make([][]int32, v.ThreadCount)
	for i, queue := range v.List {
		list[i] = make([]int32, len(queue))
		for _, id := range queue {
			list[i] = append(list[i], int32(id))
		}
	}
	vlr := VerifyLogRaw{
		List: list,
	}
	b, err := vlr.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return b
}
func (v *VerifyLog) Decode(buf []byte) error {
	vlr := VerifyLogRaw{}
	_, err := vlr.Unmarshal(buf)
	if err != nil {
		return err
	}

	v.ThreadCount = len(vlr.List)
	for i, queue := range vlr.List {
		v.List[i] = make([]int, len(queue))
		for _, id := range queue {
			v.List[i] = append(v.List[i], int(id))
		}
	}
	v.current = make([]int, v.ThreadCount)
	return nil
}
func (v *VerifyLog) Hash() []byte {
	return common.Sha256(v.Encode())
}
