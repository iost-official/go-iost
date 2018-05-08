package state


// state pool记录状态转移的结构，可以序列化到其他地方存储
type Patch struct {
	m map[Key]Value
}

func (p *Patch) Put(key Key, value Value) {
	p.m[key] = value
}
func (p *Patch) Get(key Key) Value {
	val, ok := p.m[key]
	if !ok {
		return VNil
	}
	return val
}
func (p *Patch) Has(key Key) bool {
	_, ok := p.m[key]
	return ok
}
func (p *Patch) Delete(key Key) {
	delete(p.m, key)
}
func (p *Patch) Length() int {
	return len(p.m)
}

func (p *Patch) Encode() []byte {
	pr := PatchRaw{
		keys: make([]string, 0),
		vals: make([][]byte, 0),
	}
	for k, v := range p.m {
		pr.keys = append(pr.keys, string(k))
		pr.vals = append(pr.vals, []byte(v.String()))
	}
	b, err := pr.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return b
}
func (p *Patch) Decode(bin []byte) error {
	var pr PatchRaw
	_, err := pr.Unmarshal(bin)
	if err != nil {
		return err
	}

	for i, k := range pr.keys {
		var v Value
		v, err = ParseValue(string(pr.vals[i]))
		if err != nil {
			return err
		}
		p.m[Key(k)] = v
	}
	return nil
}
func (p *Patch) Hash() []byte {
	return nil
}
