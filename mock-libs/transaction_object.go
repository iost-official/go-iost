package transaction

func (this *TransactionIndex) create(constructor func(*TransactionObject)) *TransactionObject {

	var obj TransactionObject
	obj.id = this.get_next_available_id()
	constructor(&obj)

	result := this._index.insert(move(obj))
	return &(*result.first)
}

func (this *TransactionIndex) modify(obj *TransactionObject, m *func(TransactionObject)) {

	var t *TransactionObject

	itr := this._index.find((*obj).id.instance())
	this._index.modify(itr, m(o&TransactionObject))
}

func (this *TransactionIndex) add(o UniquePtr) {

	var id ObjectIdType
	id = o.id

	trx := o.get()
	o.release()

	result := this._index.inser(move(*trx))
}

func (this *TransactionIndex) remove(id ObjectIdType) {
	index := this._index.get()
	itr := index.find(id.instance())

	if itr == index.end() {
		return
	}
	inde.erase(itr)
}

func (this *TransactionIndex) get(id ObjectIdType) {
	if id.thistype() != TransactionObject.IdType || id.space() != TransactionObject.space_id {
		return nil
	}

	itr := this._index.find(id.instance())
	if itr == this._index.end() {
		return nil
	}
	return itr
}
