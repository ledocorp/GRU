package ui

// DocumentSpecBlockTypes lists supported JSON block `type` values for BuildDocumentSpec.
// Keep in sync with docs/DOCUMENT_SPEC_AUTHORING.md and document_spec_recipe_test.go.
var DocumentSpecBlockTypes = []string{
	"page", "column", "row", "buttonRow", "form", "field",
	"section", "card", "surface", "viewport", "callout", "code", "backdrop", "presetRow",
	"text", "divider", "separator", "list",
	"badge", "progressBar", "progress",
	"button", "input", "dropdown", "checkbox", "toggle", "radioGroup", "radio", "slider",
	"listTile", "listtile", "appBar", "appbar", "chip", "rating",
	"bottomnav", "bottomNavigation", "bottomNav", "fab", "avatar", "breadcrumbs",
	"combobox", "comboBox", "dateRangePicker", "daterangepicker", "pagination",
	"toolbar", "searchBar", "searchbar", "tabView", "tabview",
	"table", "dataTable", "datatable",
}
