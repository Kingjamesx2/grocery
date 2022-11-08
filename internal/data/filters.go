// Filename: internal/data/filters.go

package data

import (
	"math"
	"strings"

	"grocery.jamesfaber.net/internal/validator"
)

type Filters struct {
	Page     int
	PageSize int
	Sort     string
	SortList []string
}

func ValidateFilters(v *validator.Validator, f Filters) {
	//Check page and page_size parameters
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 1000, "page", "maximum of 1000")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page", "maximum of 100")
	//Check that the sort parameter matches a value in the acceptable sort list
	v.Check(validator.In(f.Sort, f.SortList...), "sort", "invalid sort value")
}

// The sortColumn() method safely extracts the sort field query parameter
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

// The sortOrder() method determins whether we should sort by DESC/ASC
func (f Filters) sortOrder() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

// The limit() method determines the LIMIT
func (f Filters) limit() int {
	return f.PageSize
}

// The offset() method calculates the OFFSET
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// The metadata type contains metadata to help with pagination
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// The calculateMetaData() function computes the values for the Metadata fields
func calculateMetaData(TotalRecords int, page int, pageSize int) Metadata {
	if TotalRecords == 0 {
		return Metadata{}
	}
	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil((float64(TotalRecords) / float64(pageSize)))),
		TotalRecords: TotalRecords,
	}
}
