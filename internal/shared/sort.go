package shared

// Константы для сортировки
type SortOption string

const (
	SortUploadedNew SortOption = "uploaded_new"
	SortUploadedOld SortOption = "uploaded_old"
	SortNameAZ      SortOption = "name_az"
	SortNameZA      SortOption = "name_za"
	SortRandom      SortOption = "random"
	DefaultSort     SortOption = SortUploadedNew
)

// Валидные значения сортировки
var ValidSorts = map[SortOption]struct{}{
	SortUploadedNew: {},
	SortUploadedOld: {},
	SortNameAZ:      {},
	SortNameZA:      {},
	SortRandom:      {},
}
