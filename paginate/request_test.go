package agPaginate

import (
	"testing"
)

func TestAgPaginateOptionsRequestModel(t *testing.T) {
	testCases := []struct {
		description    string
		request        AgPaginateOptionsRequest
		expectedError  bool
		expectedOrder  string
		expectedCursor string
		expectedLimit  uint64
	}{
		{
			"Valid Empty PaginateRequest",
			AgPaginateOptionsRequest{
				Cursor:     "",
				LimitItems: 0,
				Order:      "",
			},
			false,
			"DESC",
			"",
			20,
		},
		{
			"Valid Order ASC PaginateRequest",
			AgPaginateOptionsRequest{
				Cursor:     "",
				LimitItems: 0,
				Order:      "asc",
			},
			false,
			"ASC",
			"",
			20,
		},
		{
			"Valid Order DESC PaginateRequest",
			AgPaginateOptionsRequest{
				Cursor:     "",
				LimitItems: 0,
				Order:      "desc",
			},
			false,
			"DESC",
			"",
			20,
		},
		{
			"Valid Order and Limit PaginateRequest",
			AgPaginateOptionsRequest{
				Cursor:     "",
				LimitItems: 200,
				Order:      "desc",
			},
			false,
			"DESC",
			"",
			200,
		},
		{
			"Valid Cursor Order and Limit PaginateRequest",
			AgPaginateOptionsRequest{
				Cursor:     "dat-65ab0d000000000000000000",
				LimitItems: 100,
				Order:      "desc",
			},
			false,
			"DESC",
			"dat-65ab0d000000000000000000",
			100,
		},
		{
			"Error Invalid Order PaginateRequest",
			AgPaginateOptionsRequest{
				Cursor:     "dat-65ab0d000000000000000000",
				LimitItems: 100,
				Order:      "descdelete",
			},
			true,
			"DESC",
			"dat-65ab0d000000000000000000",
			100,
		},
		{
			"Error Invalid Cursor PaginateRequest",
			AgPaginateOptionsRequest{
				Cursor: "selectfromtable",
			},
			true,
			"DESC",
			"selectfromtable",
			20,
		},
		{
			"Error Invalid Limit PaginateRequest",
			AgPaginateOptionsRequest{
				Cursor:     "",
				LimitItems: 250,
				Order:      "desc",
			},
			true,
			"DESC",
			"",
			200,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.request.Validate()
			if tc.expectedError && err == nil {
				t.Errorf("AgPaginateOptionsRequestModelValidate was expected to return an error ")
			} else if !tc.expectedError && err != nil {
				t.Errorf("AgPaginateOptionsRequestModelValidate was not expected to return an error, but it did: %v", err)
			}

			cursor := tc.request.GetCursor()
			if tc.expectedCursor != cursor {
				t.Errorf("AgPaginateOptionsRequestModel-GetCursor -> Expected: %s  // Returned: %s", tc.expectedCursor, cursor)
			}
			order := tc.request.GetOrder()
			if tc.expectedOrder != order {
				t.Errorf("AgPaginateOptionsRequestModel-GetOrder -> Expected: %s  // Returned: %s", tc.expectedOrder, order)
			}
			limitItems := tc.request.GetLimitItems()
			if tc.expectedLimit != limitItems {
				t.Errorf("AgPaginateOptionsRequestModel-GetLimitItems -> Expected: %d  // Returned: %d", tc.expectedLimit, limitItems)
			}
		})
	}
}
