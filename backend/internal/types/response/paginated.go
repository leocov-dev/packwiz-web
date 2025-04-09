package response

type Pagination struct {
	Page  int   `json:"page"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}

type Paginated[T interface{}] struct {
	Results    []T        `json:"results"`
	Pagination Pagination `json:"pagination"`
}

func NewPaginated[T interface{}](results []T, page, size int, total int64) *Paginated[T] {
	return &Paginated[T]{
		Results: results,
		Pagination: Pagination{
			Page:  page,
			Size:  size,
			Total: total,
		},
	}
}
