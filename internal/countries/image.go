package countries

import (
	"database/sql"
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"github.com/fogleman/gg"
)

// GenerateSummaryImage generates a PNG summary at destPath (e.g., cache/summary.png)
func GenerateSummaryImage(db *sql.DB, destPath string) error {
	total, err := TotalCount(db)
	if err != nil {
		return err
	}

	// get top 5 by estimated_gdp
	q := `SELECT name, estimated_gdp FROM countries WHERE estimated_gdp IS NOT NULL ORDER BY estimated_gdp DESC LIMIT 5`
	rows, err := db.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	type entry struct {
		Name string
		GDP  float64
	}
	var top []entry
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.Name, &e.GDP); err != nil {
			return err
		}
		top = append(top, e)
	}

	// create canvas
	const W = 1000
	const H = 600
	dc := gg.NewContext(W, H)
	dc.SetColor(color.White)
	dc.Clear()

	// header
	dc.SetRGB(0, 0, 0)
	if err := dc.LoadFontFace("/Library/Fonts/Arial.ttf", 28); err != nil {
		// ignore font load error; gg has fallback
	}
	dc.DrawStringAnchored(fmt.Sprintf("Countries Summary (total: %d)", total), W/2, 60, 0.5, 0.5)

	// list top5
	y := 120.0
	for i, e := range top {
		line := fmt.Sprintf("%d. %s — %.2f", i+1, e.Name, e.GDP)
		dc.DrawStringAnchored(line, 60, y, 0, 0.5)
		y += 40
	}

	// ensure directory
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	return dc.SavePNG(destPath)
}
