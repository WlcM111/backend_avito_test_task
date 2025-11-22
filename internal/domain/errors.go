package domain

import "errors"

// ErrorCodeTeamExists указывает, что команда с таким именем уже существует.
// Остальные коды описывают различные доменные ошибки.
const (
	ErrorCodeTeamExists  = "TEAM_EXISTS"
	ErrorCodePRExists    = "PR_EXISTS"
	ErrorCodePRMerged    = "PR_MERGED"
	ErrorCodeNotAssigned = "NOT_ASSIGNED"
	ErrorCodeNoCandidate = "NO_CANDIDATE"
	ErrorCodeNotFound    = "NOT_FOUND"
	ErrorCodeInternal    = "INTERNAL"
)

// ErrTeamExists возвращается, когда пытаются создать уже существующую команду.
// Остальные ошибки описывают типовые доменные ситуации без привязки к коду.
var (
	ErrTeamExists          = errors.New("team already exists")
	ErrPRExists            = errors.New("pull request already exists")
	ErrPRMerged            = errors.New("pull request already merged")
	ErrReviewerNotAssigned = errors.New("reviewer not assigned")
	ErrNoCandidate         = errors.New("no replacement candidate")
	ErrNotFound            = errors.New("not found")
)

// DomainError оборачивает доменную ошибку с кодом для HTTP-слоя.
//
//revive:disable-next-line:exported
type DomainError struct {
	Code string
	Err  error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	return e.Code
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError создаёт новую DomainError с указанным кодом и исходной ошибкой.
func NewDomainError(code string, err error) *DomainError {
	return &DomainError{
		Code: code,
		Err:  err,
	}
}
