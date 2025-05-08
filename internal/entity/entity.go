package entity

type StringError string

func (e StringError) Error() string {
	return string(e)
}

const (
	ErrPhaseFail = StringError("phase is fail")
)
