package state

type Patch struct {
	m map[Key]Value
}

func (p *Patch) Put(key Key, value Value) error {
	p.m[key] = value // TODO 在这里只需要保存增量即可
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
func (p *Patch) Encode() []byte {
	return nil
}
func (p *Patch) Decode(bin []byte) error {
	return nil
}
func (p *Patch) HashHash() []byte {
	return nil
}
