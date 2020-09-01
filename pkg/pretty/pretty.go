package pretty

import (
	"github.com/jedib0t/go-pretty/v6/table"
)

func Header(headers ...string) table.Row {
	newHeaders := []interface{}{"second", 3}
	return table.Row{newHeaders[0]}
}

func Body(body []interface{}) []table.Row {
	var tableBody []table.Row
	for _, b := range body {
		tableBody = append(tableBody, table.Row{b})
	}
	return tableBody
}

func Format(header table.Row, body []table.Row) string {
	t := table.NewWriter()
	tTemp := table.Table{}
	tTemp.Render()

	t.AppendHeader(header)
	t.AppendRows(body)

	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateFooter = false
	t.Style().Options.SeparateHeader = false
	t.Style().Options.SeparateRows = false

	return t.Render()
}
