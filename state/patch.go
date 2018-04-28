package state

type Patch struct {
	m map[Key]Value
}

func (p *Patch) Put(key Key, value Value) error {
	p.m[key] = value
	return nil
}
func (p *Patch) Get(key Key) (Value, error) {
	val, ok := p.m[key]
	if !ok {
		return nil, nil
	}
	return val, nil
}
func (p *Patch) Has(key Key) (bool, error) {
	_, ok := p.m[key]
	return ok, nil
}
func (p *Patch) Delete(key Key) error {
	delete(p.m, key)
	return nil
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
		pr.vals = append(pr.vals, v.Encode())
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
		v.Decode(pr.vals[i])
		p.m[Key(k)] = v
	}
	return nil
}
func (p *Patch) Hash() []byte {
	return nil
}
