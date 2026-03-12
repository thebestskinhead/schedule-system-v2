package main

import (
	"fmt"
	"github.com/extrame/xls"
	"strings"
)

func main() {
	xlsPath := "/workspace/schedule-system-v2/学生个人课表_.xls"

	workbook, err := xls.Open(xlsPath, "utf-8")
	if err != nil {
		fmt.Printf("打开失败: %v\n", err)
		return
	}

	sheet := workbook.GetSheet(0)
	if sheet == nil {
		fmt.Println("没有工作表")
		return
	}

	// 检查第8行（备注行）
	fmt.Println("第8行内容（备注）:")
	row8 := sheet.Row(8)
	for j := 0; j < row8.LastCol(); j++ {
		cell := row8.Col(j)
		if strings.Contains(cell, "实验") {
			fmt.Printf("  列%d: [%s]\n", j, cell)
		}
	}

	// 查找包含"实验"的单元格
	fmt.Println("\n查找包含'实验'的单元格:")
	for i := 0; i <= int(sheet.MaxRow); i++ {
		row := sheet.Row(i)
		for j := 0; j < row.LastCol(); j++ {
			cell := row.Col(j)
			if strings.Contains(cell, "数据结构与算法设计实验") {
				fmt.Printf("  行%d列%d: [%s]\n", i, j, cell)
			}
		}
	}
}
