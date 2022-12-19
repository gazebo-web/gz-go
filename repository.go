package gz

// Repository represents a set of methods of a generic repository.
// It is based in the Repository pattern, and allows to a system to change the underlying data structure
// without changing anything else.
// It could be used to provide an abstraction layer between a service and an ORM.
type Repository interface {
	Add(entity interface{}) (interface{}, *ErrMsg)
	Get(id interface{}) (interface{}, *ErrMsg)
	GetAll(offset, limit *int) ([]interface{}, *ErrMsg)
	Update(id interface{}, entity interface{}) (interface{}, *ErrMsg)
	Remove(id interface{}) (interface{}, *ErrMsg)
	Find(criteria func(element interface{}) bool) ([]interface{}, *ErrMsg)
	FindByIDs(ids []interface{}) ([]interface{}, *ErrMsg)
	FindOne(criteria func(element interface{}) bool) (interface{}, *ErrMsg)
	Clear() *ErrMsg
	Count(criteria func(element interface{}) bool) int
}
