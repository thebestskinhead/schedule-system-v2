package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/extrame/xls"
	"github.com/xuri/excelize/v2"
	"schedule-system-v2/backend/internal/model"
)

// XLSParser XLS课表解析器
type XLSParser struct {
	totalWeeks   int
	days         int
	blocksPerDay int
	bitsPerWeek  int
}

// NewXLSParser 创建XLS解析器
func NewXLSParser() *XLSParser {
	days := 5
	blocksPerDay := 4
	return &XLSParser{
		totalWeeks:   30,
		days:         days,
		blocksPerDay: blocksPerDay,
		bitsPerWeek:  days * blocksPerDay,
	}
}

// ParseXLS 解析XLS文件并返回无课时间列表
func (p *XLSParser) ParseXLS(filePath string, userID int) ([]model.Availability, error) {
	if strings.HasSuffix(strings.ToLower(filePath), ".xls") {
		return p.parseOldXLS(filePath, userID)
	}
	return p.parseXLSX(filePath, userID)
}

// parseOldXLS 解析旧版 .xls 文件 (CDFV2格式)
func (p *XLSParser) parseOldXLS(filePath string, userID int) ([]model.Availability, error) {
	workbook, err := xls.Open(filePath, "utf-8")
	if err != nil {
		return nil, fmt.Errorf("打开XLS文件失败: %w", err)
	}

	sheet := workbook.GetSheet(0)
	if sheet == nil {
		return nil, fmt.Errorf("XLS文件没有工作表")
	}

	var rows [][]string
	for i := 0; i <= int(sheet.MaxRow); i++ {
		row := sheet.Row(i)
		var rowData []string
		for j := 0; j < row.LastCol(); j++ {
			rowData = append(rowData, row.Col(j))
		}
		rows = append(rows, rowData)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("工作表为空")
	}

	return p.parseRows(rows, userID)
}

// parseXLSX 解析新版 .xlsx 文件
func (p *XLSParser) parseXLSX(filePath string, userID int) ([]model.Availability, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开XLSX文件失败: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("XLSX文件没有工作表")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取工作表失败: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("工作表为空")
	}

	return p.parseRows(rows, userID)
}

// parseRows 解析行数据
func (p *XLSParser) parseRows(rows [][]string, userID int) ([]model.Availability, error) {
	// 初始化位图（30周，每周20位，初始全1表示全部空闲）
	bitmap := make([]int, p.totalWeeks)
	for i := range bitmap {
		bitmap[i] = 0xFFFFF // 20个1
	}

	if p.isTimetableFormat(rows) {
		p.parseTimetableFormat(rows, bitmap)
	} else {
		p.parseListFormat(rows, bitmap)
	}

	availabilities := p.bitmapToAvailability(bitmap, userID)
	return availabilities, nil
}

// isTimetableFormat 检测是否为课表格式（任意行包含星期）
func (p *XLSParser) isTimetableFormat(rows [][]string) bool {
	for i := 0; i < len(rows) && i < 5; i++ {
		row := rows[i]
		if len(row) < 2 {
			continue
		}
		for _, cell := range row {
			cell = strings.TrimSpace(cell)
			if cell == "周一" || cell == "星期一" || cell == "周二" || cell == "星期二" ||
				cell == "周三" || cell == "星期三" || cell == "周四" || cell == "星期四" ||
				cell == "周五" || cell == "星期五" {
				return true
			}
		}
	}
	return false
}

// findHeaderRow 找到包含星期表头的行索引
func (p *XLSParser) findHeaderRow(rows [][]string) int {
	for i := 0; i < len(rows) && i < 10; i++ {
		row := rows[i]
		if len(row) < 2 {
			continue
		}
		for _, cell := range row {
			cell = strings.TrimSpace(cell)
			if cell == "周一" || cell == "星期一" || cell == "周二" || cell == "星期二" {
				return i
			}
		}
	}
	return 0
}

// parseTimetableFormat 解析课表格式
func (p *XLSParser) parseTimetableFormat(rows [][]string, bitmap []int) {
	periodMap := map[string]int{
		"第一二节": 0, "第1-2节": 0, "1-2节": 0, "第1、2节": 0,
		"第三四节": 1, "第3-4节": 1, "3-4节": 1, "第3、4节": 1,
		"第五六节": 2, "第5-6节": 2, "5-6节": 2, "第5、6节": 2,
		"第七八节": 3, "第7-8节": 3, "7-8节": 3, "第7、8节": 3,
	}

	dayMap := map[string]int{
		"周一": 0, "星期一": 0,
		"周二": 1, "星期二": 1,
		"周三": 2, "星期三": 2,
		"周四": 3, "星期四": 3,
		"周五": 4, "星期五": 4,
	}

	headerRowIdx := p.findHeaderRow(rows)
	if headerRowIdx >= len(rows) {
		return
	}

	var dayColIndex []int
	headerRow := rows[headerRowIdx]
	for colIdx, cell := range headerRow {
		cell = strings.TrimSpace(cell)
		if day, ok := dayMap[cell]; ok {
			for len(dayColIndex) <= colIdx {
				dayColIndex = append(dayColIndex, -1)
			}
			dayColIndex[colIdx] = day
		}
	}

	for rowIdx := headerRowIdx + 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		if len(row) == 0 {
			continue
		}

		periodCell := strings.TrimSpace(row[0])
		periodIdx := -1
		for periodName, idx := range periodMap {
			if strings.Contains(periodCell, periodName) {
				periodIdx = idx
				break
			}
		}
		if periodIdx == -1 {
			continue
		}

		for colIdx := 1; colIdx < len(row) && colIdx < len(dayColIndex); colIdx++ {
			day := dayColIndex[colIdx]
			if day == -1 {
				continue
			}

			cell := strings.TrimSpace(row[colIdx])
			if cell == "" || cell == "&nbsp;" {
				continue
			}

			p.parseAndMark(bitmap, cell, day, periodIdx)
		}
	}
}

// parseListFormat 解析列表格式
func (p *XLSParser) parseListFormat(rows [][]string, bitmap []int) {
	dayCol := -1
	courseCol := -1

	if len(rows) > 0 {
		header := rows[0]
		for colIdx, cell := range header {
			cell := strings.ToLower(strings.TrimSpace(cell))
			if cell == "星期" || cell == "day" || cell == "星期几" {
				dayCol = colIdx
			}
			if cell == "课程" || cell == "课程名" || cell == "course" || cell == "课程内容" {
				courseCol = colIdx
			}
		}
	}

	if courseCol == -1 {
		courseCol = 0
	}

	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		if len(row) <= courseCol {
			continue
		}

		cell := strings.TrimSpace(row[courseCol])
		if cell == "" {
			continue
		}

		day := 0
		if dayCol >= 0 && dayCol < len(row) {
			day = p.parseDay(strings.TrimSpace(row[dayCol]))
			if day == -1 {
				day = 0
			}
		}

		p.parseAndMarkFlexible(bitmap, cell, day)
	}
}

// parseAndMark 解析课程文本并标记占用（支持一个单元格内多门课程）
func (p *XLSParser) parseAndMark(bitmap []int, courseText string, day int, periodIdx int) {
	// 清理HTML实体
	courseText = strings.ReplaceAll(courseText, "&nbsp;", " ")
	courseText = strings.ReplaceAll(courseText, "&lt;", "<")
	courseText = strings.ReplaceAll(courseText, "&gt;", ">")
	courseText = strings.ReplaceAll(courseText, "&amp;", "&")

	// 按空行分割多个课程块
	courses := p.splitCourses(courseText)
	
	for _, course := range courses {
		p.parseSingleCourse(bitmap, strings.TrimSpace(course), day, periodIdx)
	}
}

// splitCourses 将单元格文本分割成多个课程块（修复版，移除Go不支持的正则语法）
func (p *XLSParser) splitCourses(text string) []string {
	var courses []string
	
	// 标准化换行符
	text = strings.ReplaceAll(text, "\r\n", "\n")
	
	// 策略1：按空行分割（两个及以上换行符）
	blocks := strings.Split(text, "\n\n")
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block != "" {
			courses = append(courses, block)
		}
	}
	
	// 如果策略1得到多个块，直接返回
	if len(courses) > 1 {
		return courses
	}
	
	// 策略2：如果没分割开，但有多个时间模式，按时间行分割
	if len(courses) == 1 && strings.Count(text, "[") >= 2 {
		lines := strings.Split(text, "\n")
		var currentCourse []string
		courses = nil // 清空之前的结果
		
		for i, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			
			// 检测是否是新的课程开始（包含时间模式 [xx-xx节]）
			if strings.Contains(line, "[") && strings.Contains(line, "]") && strings.Contains(line, "节") {
				// 如果已经有累积的内容，保存为一个课程
				if len(currentCourse) > 0 {
					courses = append(courses, strings.Join(currentCourse, "\n"))
					currentCourse = nil
				}
			}
			
			currentCourse = append(currentCourse, line)
			
			// 如果是最后一行，保存当前课程
			if i == len(lines)-1 && len(currentCourse) > 0 {
				courses = append(courses, strings.Join(currentCourse, "\n"))
			}
		}
		
		// 如果策略2成功分割，返回结果
		if len(courses) > 1 {
			return courses
		}
	}
	
	// 如果都没分割成功，返回原始文本（去除首尾空白）
	if len(courses) == 0 && strings.TrimSpace(text) != "" {
		return []string{strings.TrimSpace(text)}
	}
	
	return courses
}

// parseSingleCourse 解析单个课程块
func (p *XLSParser) parseSingleCourse(bitmap []int, courseText string, day int, defaultPeriodIdx int) {
	lines := strings.Split(courseText, "\n")
	
	var timeInfo string
	periodIdx := defaultPeriodIdx
	
	// 在当前课程块内查找时间信息
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.Contains(line, "[") && strings.Contains(line, "]") {
			timeInfo = line
			
			// 尝试从当前行提取节次，验证是否与默认一致
			slotRegex := regexp.MustCompile(`\[(\d+)-(\d+)节?\]`)
			slotMatch := slotRegex.FindStringSubmatch(line)
			if len(slotMatch) >= 3 {
				startSlot, _ := strconv.Atoi(slotMatch[1])
				extractedPeriodIdx := p.slotToBlock(startSlot)
				// 如果提取的节次与传入的 defaultPeriodIdx 不一致，使用提取的
				if extractedPeriodIdx != -1 {
					periodIdx = extractedPeriodIdx
				}
			}
			break
		}
	}
	
	if timeInfo == "" {
		return
	}

	// 解析周次
	weeks := p.extractWeeksFromTimeInfo(timeInfo)
	if len(weeks) == 0 {
		return
	}

	// 解析节次
	slotRegex := regexp.MustCompile(`\[(\d+)-(\d+)节?\]`)
	slotMatch := slotRegex.FindStringSubmatch(timeInfo)
	
	var startSlot, endSlot int
	if len(slotMatch) >= 3 {
		startSlot, _ = strconv.Atoi(slotMatch[1])
		endSlot, _ = strconv.Atoi(slotMatch[2])
	} else {
		// 使用默认节次映射
		switch periodIdx {
		case 0:
			startSlot, endSlot = 1, 2
		case 1:
			startSlot, endSlot = 3, 4
		case 2:
			startSlot, endSlot = 5, 6
		case 3:
			startSlot, endSlot = 7, 8
		}
	}

	_ = periodIdx // 使用变量避免编译错误，或者直接用第一种方法删除它
	p.markOccupiedFromSlots(bitmap, weeks, day, startSlot, endSlot)
}

// extractWeeksFromTimeInfo 从单行时间信息中提取周次
func (p *XLSParser) extractWeeksFromTimeInfo(timeInfo string) []int {
	// 匹配格式: 9-14([周]), 1-6([周]), 2-15[周] 等
	weekRegex := regexp.MustCompile(`(\d[\d,\-周单双\(\)\[\]]*)\[`)
	weekMatch := weekRegex.FindStringSubmatch(timeInfo)
	
	if len(weekMatch) >= 2 {
		return p.parseWeekPattern(weekMatch[1])
	}
	
	// 备用方案：直接匹配所有数字范围
	rangeRegex := regexp.MustCompile(`(\d+)-(\d+)`)
	matches := rangeRegex.FindAllStringSubmatch(timeInfo, -1)
	if len(matches) > 0 {
		var weeks []int
		for _, match := range matches {
			if len(match) >= 3 {
				start, _ := strconv.Atoi(match[1])
				end, _ := strconv.Atoi(match[2])
				if start > 0 && end <= p.totalWeeks {
					for w := start; w <= end && w <= p.totalWeeks; w++ {
						weeks = append(weeks, w)
					}
				}
			}
		}
		return weeks
	}
	
	return nil
}

// parseAndMarkFlexible 灵活解析课程文本
func (p *XLSParser) parseAndMarkFlexible(bitmap []int, courseText string, day int) {
	courseText = strings.ReplaceAll(courseText, "&nbsp;", " ")
	lines := strings.Split(courseText, "\n")

	var timeInfo string
	periodIdx := -1

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "节") {
			timeInfo = line

			if strings.Contains(line, "01-02") || strings.Contains(line, "1-2") {
				periodIdx = 0
			} else if strings.Contains(line, "03-04") || strings.Contains(line, "3-4") {
				periodIdx = 1
			} else if strings.Contains(line, "05-06") || strings.Contains(line, "5-6") {
				periodIdx = 2
			} else if strings.Contains(line, "07-08") || strings.Contains(line, "7-8") {
				periodIdx = 3
			}
			break
		}
	}

	if timeInfo == "" || periodIdx == -1 {
		return
	}

	p.parseSingleCourse(bitmap, courseText, day, periodIdx)
}

// parseWeekPattern 解析周次模式（支持逗号分割的多段区间，如"1-6,9-14"）
func (p *XLSParser) parseWeekPattern(pattern string) []int {
	var weeks []int

	// 清理字符串（保留逗号和连字符）
	pattern = strings.ReplaceAll(pattern, "[周]", "")
	pattern = strings.ReplaceAll(pattern, "(周)", "")
	pattern = strings.ReplaceAll(pattern, "(", "")
	pattern = strings.ReplaceAll(pattern, ")", "")
	pattern = strings.ReplaceAll(pattern, "周", "")
	pattern = strings.TrimSpace(pattern)

	// 检查是否为单双周标记
	isSingle := strings.Contains(pattern, "单")
	isDouble := strings.Contains(pattern, "双")
	
	// 移除单双标记
	cleanPattern := strings.ReplaceAll(pattern, "单", "")
	cleanPattern = strings.ReplaceAll(cleanPattern, "双", "")
	cleanPattern = strings.TrimSpace(cleanPattern)

	// 按逗号分割处理多段区间（如"1-6,9-14"）
	segments := strings.Split(cleanPattern, ",")
	
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}

		// 判断是范围（含"-"）还是单个数字
		if strings.Contains(segment, "-") {
			rangeRegex := regexp.MustCompile(`(\d+)-(\d+)`)
			matches := rangeRegex.FindStringSubmatch(segment)
			if len(matches) >= 3 {
				start, _ := strconv.Atoi(matches[1])
				end, _ := strconv.Atoi(matches[2])

				if start < 1 {
					start = 1
				}
				if end > p.totalWeeks {
					end = p.totalWeeks
				}

				// 应用单双周过滤
				if isSingle {
					if start%2 == 0 {
						start++
					}
					for w := start; w <= end; w += 2 {
						weeks = append(weeks, w)
					}
				} else if isDouble {
					if start%2 != 0 {
						start++
					}
					for w := start; w <= end; w += 2 {
						weeks = append(weeks, w)
					}
				} else {
					for w := start; w <= end; w++ {
						weeks = append(weeks, w)
					}
				}
			}
		} else {
			// 单个数字
			w, err := strconv.Atoi(segment)
			if err == nil && w >= 1 && w <= p.totalWeeks {
				weeks = append(weeks, w)
			}
		}
	}

	return weeks
}

// parseDay 解析星期字符串
func (p *XLSParser) parseDay(dayStr string) int {
	dayStr = strings.TrimSpace(dayStr)
	dayMap := map[string]int{
		"周一": 0, "星期一": 0,
		"周二": 1, "星期二": 1,
		"周三": 2, "星期三": 2,
		"周四": 3, "星期四": 3,
		"周五": 4, "星期五": 4,
	}

	if day, ok := dayMap[dayStr]; ok {
		return day
	}
	return -1
}

// slotToBlock 将节次映射到时段块
func (p *XLSParser) slotToBlock(slot int) int {
	switch {
	case slot >= 1 && slot <= 2:
		return 0
	case slot >= 3 && slot <= 4:
		return 1
	case slot >= 5 && slot <= 6:
		return 2
	case slot >= 7 && slot <= 10:
		return 3
	default:
		return -1
	}
}

// markOccupiedFromSlots 根据节次标记占用
func (p *XLSParser) markOccupiedFromSlots(bitmap []int, weeks []int, day int, startSlot int, endSlot int) {
	startBlock := p.slotToBlock(startSlot)
	endBlock := p.slotToBlock(endSlot)

	if startBlock == -1 || endBlock == -1 {
		return
	}

	for _, week := range weeks {
		if week < 1 || week > p.totalWeeks {
			continue
		}

		weekIdx := week - 1
		for block := startBlock; block <= endBlock; block++ {
			bitPosition := day*4 + block
			if bitPosition < p.bitsPerWeek {
				bitmap[weekIdx] &= ^(1 << bitPosition)
			}
		}
	}
}

// bitmapToAvailability 将位图转换为无课时间列表
func (p *XLSParser) bitmapToAvailability(bitmap []int, userID int) []model.Availability {
	var availabilities []model.Availability

	for weekIdx, weekBits := range bitmap {
		week := weekIdx + 1
		for day := 0; day < p.days; day++ {
			for block := 0; block < p.blocksPerDay; block++ {
				bitPosition := day*4 + block
				if (weekBits>>bitPosition)&1 == 1 {
					availabilities = append(availabilities, model.Availability{
						UserID:  userID,
						Week:    week,
						Weekday: day + 1,
						Period:  block + 1,
					})
				}
			}
		}
	}

	return availabilities
}