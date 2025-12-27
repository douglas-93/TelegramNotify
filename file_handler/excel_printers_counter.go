package file_handler

import (
	"LapaTelegramBot/monitor"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

// GenerateSheet cria uma planilha Excel formatada com os contadores de impressoras
func GenerateSheet(printers []monitor.Printer) (string, error) {
	fileName := fmt.Sprintf("contadores_%s.xlsx", time.Now().Format("2006-01-02_15-04-05"))
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetName := "Contadores"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return "", err
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// Define estilos
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Color:  "FFFFFF",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"4CAF50"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   16,
			Color:  "333333",
			Family: "Calibri",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
	})

	numberStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "right",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
		NumFmt: 3, // Formato de número com separador de milhares
	})

	alternateRowStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"F2F2F2"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
	})

	alternateNumberStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"F2F2F2"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "right",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
		},
		NumFmt: 3,
	})

	// Título
	f.SetCellValue(sheetName, "A1", "RELATÓRIO DE CONTADORES DE IMPRESSORAS")
	f.SetCellStyle(sheetName, "A1", "D1", titleStyle)
	f.MergeCell(sheetName, "A1", "D1")
	f.SetRowHeight(sheetName, 1, 30)

	// Data de geração
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("Gerado em: %s", time.Now().Format("02/01/2006 às 15:04:05")))
	f.MergeCell(sheetName, "A2", "D2")

	// Cabeçalho
	f.SetCellValue(sheetName, "A4", "Impressora")
	f.SetCellValue(sheetName, "B4", "Preto e Branco")
	f.SetCellValue(sheetName, "C4", "Colorido")
	f.SetCellValue(sheetName, "D4", "Total")
	f.SetCellStyle(sheetName, "A4", "D4", headerStyle)
	f.SetRowHeight(sheetName, 4, 25)

	// Dados
	line := 5

	for i, printer := range printers {
		// Alterna estilo de linha
		var cellStyle, numStyle int
		if i%2 == 0 {
			cellStyle = dataStyle
			numStyle = numberStyle
		} else {
			cellStyle = alternateRowStyle
			numStyle = alternateNumberStyle
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", line), printer.HostData.Host)
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", line), fmt.Sprintf("A%d", line), cellStyle)

		f.SetCellValue(sheetName, fmt.Sprintf("B%d", line), printer.BlackCounter)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", line), fmt.Sprintf("B%d", line), numStyle)

		f.SetCellValue(sheetName, fmt.Sprintf("C%d", line), printer.ColorCounter)
		f.SetCellStyle(sheetName, fmt.Sprintf("C%d", line), fmt.Sprintf("C%d", line), numStyle)

		f.SetCellValue(sheetName, fmt.Sprintf("D%d", line), printer.TotalCounter)
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", line), fmt.Sprintf("D%d", line), numStyle)

		line++
	}

	// Ajusta largura das colunas
	f.SetColWidth(sheetName, "A", "A", 35)
	f.SetColWidth(sheetName, "B", "B", 18)
	f.SetColWidth(sheetName, "C", "C", 18)
	f.SetColWidth(sheetName, "D", "D", 18)

	// Congela primeira linha de cabeçalho
	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      4,
		TopLeftCell: "A5",
		ActivePane:  "bottomLeft",
	})

	// Salva a planilha
	if err := f.SaveAs(fileName); err != nil {
		return "", err
	}

	return fileName, nil
}
