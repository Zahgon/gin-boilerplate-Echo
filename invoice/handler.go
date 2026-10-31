package invoice

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// InvoiceController ...
type InvoiceController struct{}

// ItemView holds pre-formatted values for the HTML template.
type ItemView struct {
	Description string
	Quantity    int
	UnitPrice   string
	LineTotal   string
}

// Preview Invoice godoc
// @Summary Preview Invoice as HTML
// @Schemes
// @Description Returns an HTML preview of a sample invoice
// @Tags Invoice
// @Produce html
// @Success 200 {string} string "HTML page"
// @Router /invoice [GET]
func (ctrl InvoiceController) Preview(c echo.Context) error {
	inv := SampleInvoice()

	var items []ItemView
	for _, item := range inv.Items {
		items = append(items, ItemView{
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   fmt.Sprintf("%.2f", item.UnitPrice),
			LineTotal:   fmt.Sprintf("%.2f", float64(item.Quantity)*item.UnitPrice),
		})
	}

	// Requires template "invoice.html" loaded by the echo.Renderer set in main.go
	return c.Render(http.StatusOK, "invoice.html", echo.Map{
		"invoice": inv,
		"items":   items,
		"total":   fmt.Sprintf("%.2f", inv.Total()),
	})
}

// Download Invoice godoc
// @Summary Download Invoice as PDF
// @Schemes
// @Description Generates and downloads a sample invoice as PDF
// @Tags Invoice
// @Produce application/pdf
// @Success 200 {file} binary "PDF file"
// @Failure 500 {object} object "Error generating PDF"
// @Router /invoice/download [GET]
func (ctrl InvoiceController) Download(c echo.Context) error {
	inv := SampleInvoice()

	pdfBytes, err := GeneratePDF(inv)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Could not generate PDF"})
	}

	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, inv.Number))
	return c.Blob(http.StatusOK, "application/pdf", pdfBytes)
}
