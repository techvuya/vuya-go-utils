package agPaginate

import (
	"errors"
	"strings"

	mathutils "github.com/techvuya/vuya-go-utils/math"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type AgPaginateOptionsRequest struct {
	Cursor     string `json:"cursor"`
	LimitItems uint64 `json:"limitItems"`
	Page       uint64 `json:"page"`
	Order      string `json:"order"` // asc, desc
}

func (a AgPaginateOptionsRequest) GetPageString() string {
	if a.Page == 0 || a.Page == 1 {
		return mathutils.ConvertUint64ToString(0)
	}
	return mathutils.ConvertUint64ToString(a.Page - 1)
}

func (a AgPaginateOptionsRequest) GetPage() uint64 {
	if a.Page == 0 || a.Page == 1 {
		return 0
	}
	return a.Page - 1
}

func (a AgPaginateOptionsRequest) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.Cursor),
		//validation.Field(&a.LimitItems, validation.By(ValidatePaginateLimitItems)),
		validation.Field(&a.Page),
		validation.Field(&a.Order, validation.By(ValidatePaginateOrder)),
	)
}

func ValidatePaginateOrder(value interface{}) error {
	s, _ := value.(string)
	err := validation.Validate(strings.ToUpper(s),
		validation.In(
			"",
			"ASC",
			"DESC",
		),
	)
	return err
}

func ValidatePaginateLimitItems(value interface{}) error {
	s, _ := value.(uint64)
	if s > 200 {
		return errors.New("Invalid limitItems")
	}
	return nil
}

func (a *AgPaginateOptionsRequest) GetCursor() string {
	return a.Cursor
}

func (a *AgPaginateOptionsRequest) HasCursor() bool {
	return a.Cursor != ""
}

func (a *AgPaginateOptionsRequest) GetOrder() string {
	if a.Order == "" || strings.ToLower(a.Order) == "desc" {
		return "DESC"
	}
	if strings.ToLower(a.Order) == "asc" {
		return "ASC"
	}
	return "DESC"
}
func (a *AgPaginateOptionsRequest) GetLimitItems() uint64 {
	if a.LimitItems == 0 {
		return 20
	}
	if a.LimitItems > 200 {
		return 200
	}
	return a.LimitItems
}
