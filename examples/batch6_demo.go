//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() { Register(func() Scene { return &batch6Scene{} }) }

// ─────────────────────────────────────────────────────────────────────────────
// Row types
// ─────────────────────────────────────────────────────────────────────────────

type employeeRow struct {
	ID         int
	Name       string
	Department string
	Role       string
	Salary     int
	Score      float64 // performance 0..100
	Active     bool
}

type productRow struct {
	SKU      string
	Name     string
	Category string
	Price    float64
	Stock    int
	Rating   float64 // 0..5
}

// ─────────────────────────────────────────────────────────────────────────────
// Demo data generators
// ─────────────────────────────────────────────────────────────────────────────

var firstNames = []string{
	"Alice", "Bob", "Carol", "David", "Eve", "Frank", "Grace", "Hank",
	"Iris", "Jack", "Karen", "Liam", "Maya", "Nick", "Olivia", "Paul",
	"Quinn", "Rachel", "Sam", "Tara", "Uma", "Victor", "Wendy", "Xander",
	"Yuki", "Zara",
}
var lastNames = []string{
	"Smith", "Jones", "Brown", "Wilson", "Evans", "Taylor", "Thomas",
	"Roberts", "Walker", "White", "Hall", "Harris", "Martin", "Thompson",
	"Garcia", "Martinez", "Robinson", "Clark", "Lewis", "Lee",
}
var departments = []string{
	"Engineering", "Design", "Marketing", "Sales", "Finance", "HR",
	"Product", "Support", "Legal", "Operations",
}
var roles = []string{
	"Engineer", "Senior Engineer", "Lead", "Manager", "Director",
	"Analyst", "Specialist", "Coordinator", "Associate", "VP",
}
var categories = []string{
	"Electronics", "Accessories", "Software", "Hardware", "Books",
	"Furniture", "Networking", "Peripherals",
}
var productAdjectives = []string{
	"Pro", "Elite", "Ultra", "Smart", "Portable", "Compact", "Premium",
	"Advanced", "Wireless", "Hybrid",
}
var productNouns = []string{
	"Monitor", "Keyboard", "Mouse", "Hub", "Dock", "Headset", "Webcam",
	"Speaker", "Cable", "Adapter", "Charger", "Stand",
}

func generateEmployees(n int) []employeeRow {
	rng := rand.New(rand.NewSource(42))
	rows := make([]employeeRow, n)
	for i := range rows {
		fn := firstNames[rng.Intn(len(firstNames))]
		ln := lastNames[rng.Intn(len(lastNames))]
		rows[i] = employeeRow{
			ID:         1000 + i + 1,
			Name:       fn + " " + ln,
			Department: departments[rng.Intn(len(departments))],
			Role:       roles[rng.Intn(len(roles))],
			Salary:     40000 + rng.Intn(120000),
			Score:      math.Round(rng.Float64()*1000) / 10,
			Active:     rng.Float64() > 0.15,
		}
	}
	return rows
}

func generateProducts(n int) []productRow {
	rng := rand.New(rand.NewSource(99))
	rows := make([]productRow, n)
	for i := range rows {
		adj := productAdjectives[rng.Intn(len(productAdjectives))]
		noun := productNouns[rng.Intn(len(productNouns))]
		rows[i] = productRow{
			SKU:      fmt.Sprintf("SKU-%04d", 1000+i+1),
			Name:     adj + " " + noun,
			Category: categories[rng.Intn(len(categories))],
			Price:    math.Round((9.99+rng.Float64()*490)*100) / 100,
			Stock:    rng.Intn(500),
			Rating:   math.Round(rng.Float64()*50) / 10,
		}
	}
	return rows
}

// ─────────────────────────────────────────────────────────────────────────────
// Scene
// ─────────────────────────────────────────────────────────────────────────────

// batch6Scene demonstrates the DataTable widget (Batch 6).
//
// Two panels + info panel:
//
//   - Left   "Employees (120 rows)"  — sortable, custom status cell, row click.
//   - Center "Products (80 rows)"   — sortable, rating bar cell, horizontal scroll.
//   - Right  "Selection & Sort"     — reactive selection display + sort controls.
type batch6Scene struct {
	BaseScene
}

func (s *batch6Scene) Title() string { return "Batch 6 · DataTable" }

func (s *batch6Scene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5DataTablePanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func (s *batch6Scene) Build(doc *ui.Document) {
	// ── Page shell → main viewport (see page_shell.go) ────────────────────────
	page := MountAppPage(doc, "b6",
		"Widget Batch 6 · DataTable",
		"Virtual table with sortable columns, row selection, and horizontal scrolling.")
	page.Body.Gap = 12

	grid := NewBatchPageGrid("b6-grid", 12)

	// ══════════════════════════════════════════════════════════════════════════
	// Panel 1: Employees table
	// ══════════════════════════════════════════════════════════════════════════
	pEmp := ui.NewPanel("p-b6-emp", "Employees (120 rows)", 0, 0, 0, 0)
	setSpans5DataTablePanel(pEmp, 12, 12, 6, 4, 4)
	pEmp.Gap = 6
	pEmp.TitleHeight = 32

	empBinding := ui.NewListBinding(generateEmployees(120))

	empCols := []ui.Column[employeeRow]{
		{
			Title: "#", Width: 50, Align: ui.ColumnAlignRight,
			Sortable: true,
			SortLess: func(a, b employeeRow) bool { return a.ID < b.ID },
			Render:   func(e employeeRow) string { return fmt.Sprintf("%d", e.ID) },
		},
		{
			Title: "Name", Width: 150, Sortable: true,
			SortLess: func(a, b employeeRow) bool { return a.Name < b.Name },
			Render:   func(e employeeRow) string { return e.Name },
		},
		{
			Title: "Dept", Width: 110, Sortable: true,
			SortLess: func(a, b employeeRow) bool { return a.Department < b.Department },
			Render:   func(e employeeRow) string { return e.Department },
		},
		{
			Title: "Role", Width: 130, Sortable: true,
			SortLess: func(a, b employeeRow) bool { return a.Role < b.Role },
			Render:   func(e employeeRow) string { return e.Role },
		},
		{
			Title: "Salary", Width: 90, Align: ui.ColumnAlignRight,
			Sortable: true,
			SortLess: func(a, b employeeRow) bool { return a.Salary < b.Salary },
			Render:   func(e employeeRow) string { return fmt.Sprintf("$%d", e.Salary) },
		},
		{
			Title: "Score", Width: 70, Align: ui.ColumnAlignRight,
			Sortable: true,
			SortLess: func(a, b employeeRow) bool { return a.Score < b.Score },
			Render:   func(e employeeRow) string { return fmt.Sprintf("%.1f", e.Score) },
		},
		{
			// Status column — custom cell draw: coloured pill.
			Title: "Status", Width: 80,
			CellDraw: func(e employeeRow, _ int, isSelected bool, bounds rl.Rectangle) {
				label := "Active"
				var bg, fg rl.Color
				if e.Active {
					bg = rl.NewColor(220, 252, 231, 255)
					fg = rl.NewColor(22, 101, 52, 255)
				} else {
					bg = rl.NewColor(254, 226, 226, 255)
					fg = rl.NewColor(153, 27, 27, 255)
					label = "Inactive"
				}
				if isSelected {
					bg = rl.NewColor(199, 210, 254, 255)
					fg = rl.NewColor(30, 27, 75, 255)
				}
				pillW := float32(68)
				pillH := float32(22)
				pillX := bounds.X + (bounds.Width-pillW)/2
				pillY := bounds.Y + (bounds.Height-pillH)/2
				pill := rl.NewRectangle(pillX, pillY, pillW, pillH)
				rl.DrawRectangleRounded(pill, 1.0, 6, bg)
				st := ui.Style{FontSize: 12, TextColor: fg}
				tw := float32(ui.MeasureText(label, 12))
				ui.DrawText(label, pillX+(pillW-tw)/2, float32(ui.TextPosY(pill, st)), 12, fg)
			},
		},
	}

	empTable := ui.NewDataTable("emp-table", empCols, empBinding, 0, 0, 0, 340)

	// Selection info label (updated reactively in info panel — cross-reference
	// via captured binding).
	empInfoLbl, empInfoText := FlexCopyPair("emp-info", "form-value", "Click a row to select")
	empBinding.SubscribeSelection(func() {
		idx := empBinding.GetSelectedIndex()
		if idx < 0 {
			empInfoText.Set("Click a row to select")
			return
		}
		e := empBinding.GetSelectedItem()
		empInfoText.Set(fmt.Sprintf("Selected: %s  |  %s", e.Name, e.Department))
	})

	pEmp.AddChild(empTable)
	pEmp.AddChild(empInfoLbl)

	// ══════════════════════════════════════════════════════════════════════════
	// Panel 2: Products table — demonstrates horizontal scroll + custom cell
	// ══════════════════════════════════════════════════════════════════════════
	pProd := ui.NewPanel("p-b6-prod", "Products (wide columns)", 0, 0, 0, 0)
	setSpans5DataTablePanel(pProd, 12, 12, 6, 4, 4)
	pProd.Gap = 6
	pProd.TitleHeight = 32

	prodBinding := ui.NewListBinding(generateProducts(80))

	prodCols := []ui.Column[productRow]{
		{
			Title: "SKU", Width: 90,
			Sortable: true,
			SortLess: func(a, b productRow) bool { return a.SKU < b.SKU },
			Render:   func(p productRow) string { return p.SKU },
		},
		{
			Title: "Name", Width: 160, Sortable: true,
			SortLess: func(a, b productRow) bool { return a.Name < b.Name },
			Render:   func(p productRow) string { return p.Name },
		},
		{
			Title: "Category", Width: 110, Sortable: true,
			SortLess: func(a, b productRow) bool { return a.Category < b.Category },
			Render:   func(p productRow) string { return p.Category },
		},
		{
			Title: "Price", Width: 80, Align: ui.ColumnAlignRight,
			Sortable: true,
			SortLess: func(a, b productRow) bool { return a.Price < b.Price },
			Render:   func(p productRow) string { return fmt.Sprintf("$%.2f", p.Price) },
		},
		{
			Title: "Stock", Width: 70, Align: ui.ColumnAlignRight,
			Sortable: true,
			SortLess: func(a, b productRow) bool { return a.Stock < b.Stock },
			Render:   func(p productRow) string { return fmt.Sprintf("%d", p.Stock) },
		},
		{
			// Rating column — mini bar chart inside the cell.
			Title: "Rating", Width: 100,
			CellDraw: func(p productRow, _ int, isSelected bool, bounds rl.Rectangle) {
				// Numeric label.
				label := fmt.Sprintf("%.1f", p.Rating)
				st := ui.Style{FontSize: 12, TextColor: rl.NewColor(60, 65, 80, 255)}
				tw := float32(ui.MeasureText(label, 12))
				ui.DrawText(label, bounds.X+6, float32(ui.TextPosY(bounds, st)), 12, st.TextColor)
				// Filled bar.
				barX := bounds.X + 6 + tw + 4
				barW := bounds.Width - (6 + tw + 4) - 8
				barH := float32(8)
				barY := bounds.Y + (bounds.Height-barH)/2
				// Track.
				rl.DrawRectangleRounded(rl.NewRectangle(barX, barY, barW, barH), 1.0, 4,
					rl.NewColor(226, 228, 238, 255))
				// Fill.
				fillW := barW * float32(p.Rating) / 5.0
				if fillW > 0 {
					fillColor := rl.NewColor(79, 70, 229, 220)
					if isSelected {
						fillColor = rl.NewColor(255, 255, 255, 200)
					}
					rl.DrawRectangleRounded(rl.NewRectangle(barX, barY, fillW, barH), 1.0, 4, fillColor)
				}
			},
		},
		// Extra wide column to force horizontal scrolling.
		{
			Title: "Description / Notes", Width: 220,
			Render: func(p productRow) string {
				return fmt.Sprintf("Category: %s - see catalog page %d", p.Category, 100+p.Stock%50)
			},
		},
	}

	prodTable := ui.NewDataTable("prod-table", prodCols, prodBinding, 0, 0, 0, 340)

	prodInfoLbl, prodInfoText := FlexCopyPair("prod-info", "form-value", "Shift+Wheel or drag the bottom bar to pan columns")
	prodBinding.SubscribeSelection(func() {
		idx := prodBinding.GetSelectedIndex()
		if idx < 0 {
			prodInfoText.Set("Shift+Wheel or drag the bottom bar to pan columns")
			return
		}
		p := prodBinding.GetSelectedItem()
		prodInfoText.Set(fmt.Sprintf("Selected: %s  |  $%.2f", p.Name, p.Price))
	})

	pProd.AddChild(prodTable)
	pProd.AddChild(prodInfoLbl)

	// ══════════════════════════════════════════════════════════════════════════
	// Panel 3: Controls — reactive selection info + programmatic sort
	// ══════════════════════════════════════════════════════════════════════════
	pCtrl := ui.NewPanel("p-b6-ctrl", "Selection", 0, 0, 0, 0)
	setSpans5DataTablePanel(pCtrl, 12, 12, 12, 4, 4)
	pCtrl.Gap = 4
	pCtrl.TitleHeight = 32

	// ── Employee controls ─────────────────────────────────────────────────────
	pCtrl.AddChild(ui.NewSeparator("b6-sep1", "Employees", 0, 0, 0, 22))

	empSelID, empSelIDText := FlexCopyPair("emp-sel-id", "form-label", "ID: -")
	empSelName, empSelNameText := FlexCopyPair("emp-sel-name", "form-value", "Name: -")
	empSelDept, empSelDeptText := FlexCopyPair("emp-sel-dept", "form-value", "Dept: -")
	empSelSal, empSelSalText := FlexCopyPair("emp-sel-sal", "form-value", "Salary: -")
	empSelStat, empSelStatText := FlexCopyPair("emp-sel-stat", "form-value", "Status: -")

	empBinding.SubscribeSelection(func() {
		e := empBinding.GetSelectedItem()
		idx := empBinding.GetSelectedIndex()
		if idx < 0 {
			empSelIDText.Set("ID: -")
			empSelNameText.Set("Name: -")
			empSelDeptText.Set("Dept: -")
			empSelSalText.Set("Salary: -")
			empSelStatText.Set("Status: -")
			return
		}
		status := "Inactive"
		if e.Active {
			status = "Active"
		}
		empSelIDText.Set(fmt.Sprintf("ID: %d", e.ID))
		empSelNameText.Set("Name: " + e.Name)
		empSelDeptText.Set("Dept: " + e.Department)
		empSelSalText.Set(fmt.Sprintf("Salary: $%d", e.Salary))
		empSelStatText.Set("Status: " + status)
	})

	pCtrl.AddChild(empSelID)
	pCtrl.AddChild(empSelName)
	pCtrl.AddChild(empSelDept)
	pCtrl.AddChild(empSelSal)
	pCtrl.AddChild(empSelStat)

	// Clear selection button.
	empClearBtn := ui.NewButton("emp-clear", "Clear Employee", 0, 0, 0, 32)
	empClearBtn.SetStyle("button")
	empClearBtn.OnClick = func() { empBinding.SetSelectedIndex(-1) }
	pCtrl.AddChild(empClearBtn)

	// Select first/last.
	empFirstLastRow := ui.NewContainer("emp-fl-row", 0, 0, 0, 32)
	empFirstLastRow.FlexDirection = ui.FlexRow
	empFirstLastRow.Gap = 8
	empFirstLastRow.SetStyle("transparent")

	empFirstBtn := ui.NewButton("emp-first", "First", 0, 0, 0, 32)
	empFirstBtn.SetFlexGrow(1)
	empFirstBtn.SetStyle("button")
	empFirstBtn.OnClick = func() { empBinding.SetSelectedIndex(0) }

	empLastBtn := ui.NewButton("emp-last", "Last", 0, 0, 0, 32)
	empLastBtn.SetFlexGrow(1)
	empLastBtn.SetStyle("button")
	empLastBtn.OnClick = func() { empBinding.SetSelectedIndex(empBinding.Len() - 1) }

	empFirstLastRow.AddChild(empFirstBtn)
	empFirstLastRow.AddChild(empLastBtn)
	pCtrl.AddChild(empFirstLastRow)

	// ── Product controls ──────────────────────────────────────────────────────
	pCtrl.AddChild(ui.NewSeparator("b6-sep2", "Products", 0, 0, 0, 22))

	prodSelSKU, prodSelSKUText := FlexCopyPair("prod-sel-sku", "form-label", "SKU: -")
	prodSelName, prodSelNameText := FlexCopyPair("prod-sel-name", "form-value", "Name: -")
	prodSelPx, prodSelPxText := FlexCopyPair("prod-sel-px", "form-value", "Price: -")
	prodSelStock, prodSelStockText := FlexCopyPair("prod-sel-stock", "form-value", "Stock: -")

	prodBinding.SubscribeSelection(func() {
		p := prodBinding.GetSelectedItem()
		idx := prodBinding.GetSelectedIndex()
		if idx < 0 {
			prodSelSKUText.Set("SKU: -")
			prodSelNameText.Set("Name: -")
			prodSelPxText.Set("Price: -")
			prodSelStockText.Set("Stock: -")
			return
		}
		prodSelSKUText.Set("SKU: " + p.SKU)
		prodSelNameText.Set("Name: " + p.Name)
		prodSelPxText.Set(fmt.Sprintf("Price: $%.2f", p.Price))
		prodSelStockText.Set(fmt.Sprintf("Stock: %d units", p.Stock))
	})

	pCtrl.AddChild(prodSelSKU)
	pCtrl.AddChild(prodSelName)
	pCtrl.AddChild(prodSelPx)
	pCtrl.AddChild(prodSelStock)

	prodClearBtn := ui.NewButton("prod-clear", "Clear Product", 0, 0, 0, 32)
	prodClearBtn.SetStyle("button")
	prodClearBtn.OnClick = func() { prodBinding.SetSelectedIndex(-1) }
	pCtrl.AddChild(prodClearBtn)

	// ── Tips ──────────────────────────────────────────────────────────────────
	pCtrl.AddChild(ui.NewSeparator("b6-sep3", "Tips", 0, 0, 0, 22))

	tips := []string{
		"Click a header to sort",
		"Click again to reverse",
		"Click a row to select",
		"Shift+Wheel pans columns",
		"Drag bottom bar to pan",
	}
	for i, tip := range tips {
		pCtrl.AddChild(FlexCopy(fmt.Sprintf("b6-tip-%d", i), "form-value", "- "+tip))
	}

	// ── Assemble ──────────────────────────────────────────────────────────────
	grid.AddChild(pEmp)
	grid.AddChild(pProd)
	grid.AddChild(pCtrl)

	page.Body.AddChild(grid)
}
