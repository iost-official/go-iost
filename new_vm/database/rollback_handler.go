package database

type RollbackHandler struct {
	db IMultiValue
}

func newRollbackHandler(db IMultiValue) RollbackHandler {
	return RollbackHandler{
		db: db,
	}
}

func (m *RollbackHandler) Commit() {
	m.db.Commit()
}

func (m *RollbackHandler) Rollback() {
	m.db.Rollback()
}
