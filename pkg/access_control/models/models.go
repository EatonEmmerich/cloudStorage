package models

type PERM int

const (
	READ PERM = 1
	WRITE PERM = 2
	SHARE PERM = 4
)

func (p PERM) String() string {
	switch p {
	case READ:
		return "Read access to document"
	case WRITE:
		return "Write access to document"
	}
	return "Unknown permission type"
}

