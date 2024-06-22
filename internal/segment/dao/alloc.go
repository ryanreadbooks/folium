package dao

// segment table
const (
	TableName    = "alloc_table"
	allocColumns = "id, biz_key, cur_id, step, created_at, updated_at"

	defaultStep  uint32 = 1000
	defaultCurId uint64 = 1
)

// dao Alloc instance representation
type Alloc struct {
	Id        int64  // id primary key
	Key       string // biz_key unique key
	CurId     uint64 // cur_id
	Step      uint32 // step
	CreatedAt int64  // created_at
	UpdatedAt int64  // updated_at
}
