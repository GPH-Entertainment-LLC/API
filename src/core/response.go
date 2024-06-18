package core

type AddPackSuccessResp struct {
	Message string `json:"message"`
}

type ImageIdResp struct {
	ID string `json:"id"`
}

type SvcError struct {
	Message string `json:"message"`
}

func (e *SvcError) Error() string {
	return e.Message
}

type ErrorResp struct {
	Message string `json:"message"`
}

type DBErrorResp struct {
	Message string `json:"message"`
}

func (e *DBErrorResp) Error() string {
	return e.Message
}

func (e *ErrorResp) Error() string {
	return e.Message
}

type AmountResp struct {
	Amount uint64
}
