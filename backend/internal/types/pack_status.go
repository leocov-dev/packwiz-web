package types

type PackStatus string

const (
	PackStatusPublished PackStatus = "published"
	PackStatusDraft     PackStatus = "draft"
	PackStatusHidden    PackStatus = "hidden"
)
