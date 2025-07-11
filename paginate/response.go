package agPaginate

type AgPaginateOptionsResponse struct {
	TotalItems uint64 `json:"totalItems"`
	LimitItems uint64 `json:"limitItems"`
}

type AgPaginatePagesResponse struct {
	TotalPages uint64 `json:"totalPages"`
	ActualPage uint64 `json:"actualPage"`
}

type AgPaginateCursorResponse struct {
	Cursor                string `json:"cursor"`
	ResponseLastKeyCursor string `json:"responseLastKeyCursor,omitempty"`
	Order                 string `json:"order"`
	LimitItems            uint64 `json:"limitItems"`
	HasMoreData           bool   `json:"hasMoreData"`
}

func (a AgPaginateCursorResponse) HaveCursor() bool {
	if a.Cursor != "" {
		return true
	}
	return false
}

func CreateAgPaginateCursorResponse(paginateRequest AgPaginateOptionsRequest, responseSize int) AgPaginateCursorResponse {
	hasMoreData := false
	if responseSize > int(paginateRequest.GetLimitItems()) {
		hasMoreData = true
	}
	return AgPaginateCursorResponse{Cursor: paginateRequest.Cursor, LimitItems: paginateRequest.GetLimitItems(), HasMoreData: hasMoreData, Order: paginateRequest.GetOrder()}
}

func CreateAgPaginateCursorDynamoResponse(paginateRequest AgPaginateOptionsRequest, lastKeyCursor string) AgPaginateCursorResponse {
	hasMoreData := false
	if lastKeyCursor != "" {
		hasMoreData = true
	}
	return AgPaginateCursorResponse{
		Cursor:                paginateRequest.Cursor,
		LimitItems:            paginateRequest.GetLimitItems(),
		HasMoreData:           hasMoreData,
		Order:                 paginateRequest.GetOrder(),
		ResponseLastKeyCursor: lastKeyCursor,
	}
}

type CursorItem interface {
	GetCursorID() string
}

func CreateAgPaginateCursorDynamoResponseV2[T CursorItem](paginateRequest AgPaginateOptionsRequest, lastKeyCursor string, responseItems *[]T) AgPaginateCursorResponse {
	hasMoreData := false
	if lastKeyCursor == "" {
		return AgPaginateCursorResponse{
			Cursor:                paginateRequest.Cursor,
			LimitItems:            paginateRequest.GetLimitItems(),
			HasMoreData:           hasMoreData,
			Order:                 paginateRequest.GetOrder(),
			ResponseLastKeyCursor: lastKeyCursor,
		}
	}
	responseSize := len(*responseItems)

	if uint64(responseSize) > paginateRequest.GetLimitItems() {
		hasMoreData = true
		*responseItems = (*responseItems)[:paginateRequest.GetLimitItems()]
		lastKeyCursor = (*responseItems)[len(*responseItems)-1].GetCursorID()
	} else {
		lastKeyCursor = ""
	}

	return AgPaginateCursorResponse{
		Cursor:                paginateRequest.Cursor,
		LimitItems:            paginateRequest.GetLimitItems(),
		HasMoreData:           hasMoreData,
		Order:                 paginateRequest.GetOrder(),
		ResponseLastKeyCursor: lastKeyCursor,
	}
}
