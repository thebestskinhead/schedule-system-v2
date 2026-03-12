package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"schedule-system-v2/backend/internal/model"
)

// EduCrawler 教务系统爬虫
type EduCrawler struct {
	client  *http.Client
	cookies string
}

// NewEduCrawler 创建爬虫实例
func NewEduCrawler(cookies string) *EduCrawler {
	return &EduCrawler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cookies: cookies,
	}
}

// CrawlSchedule 爬取指定学期的课表
// 返回: week -> [weekday][period]bool (true表示有课)
func (c *EduCrawler) CrawlSchedule(semester string) (map[int][5][4]bool, error) {
	result := make(map[int][5][4]bool)

	// 遍历1-30周
	for week := 1; week <= 30; week++ {
		html, err := c.fetchSchedulePage(semester, week)
		if err != nil {
			return nil, fmt.Errorf("获取第%d周课表失败: %w", week, err)
		}

		// 解析课表
		schedule := c.parseSchedule(html)
		result[week] = schedule

		// 防止请求过快
		time.Sleep(200 * time.Millisecond)
	}

	return result, nil
}

// fetchSchedulePage 获取某周的课表页面
func (c *EduCrawler) fetchSchedulePage(semester string, week int) (string, error) {
	formData := url.Values{
		"cj0701id":  {""},
		"zc":        {strconv.Itoa(week)},
		"demo":      {""},
		"xnxq01id":  {semester},
		"sfFD":      {"1"},
		"wkbkc":     {"1"},
		"kbjcmsid":  {"0B841F8A531A4C05BE8DB7DB4B40AEF1"},
	}

	req, err := http.NewRequest(
		"POST",
		"https://kdjw.hnust.edu.cn/jsxsd/xskb/xskb_list.do",
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		return "", err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", c.cookies)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://kdjw.hnust.edu.cn/jsxsd/xskb/xskb_list.do")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// parseSchedule 解析HTML课表
// 返回: [weekday][period]bool (true表示有课)
func (c *EduCrawler) parseSchedule(html string) [5][4]bool {
	var schedule [5][4]bool

	// 查找课表表格 - 使用非贪婪匹配
	tableRegex := regexp.MustCompile(`(?s)<table[^>]*id="timetable"[^>]*>(.*?)</table>`)
	tableMatch := tableRegex.FindStringSubmatch(html)
	if len(tableMatch) < 2 {
		fmt.Printf("未找到课表表格，HTML长度: %d\n", len(html))
		return schedule
	}
	table := tableMatch[1]

	// 查找所有行
	rowRegex := regexp.MustCompile(`(?s)<tr[^>]*>(.*?)</tr>`)
	rows := rowRegex.FindAllStringSubmatch(table, -1)

	fmt.Printf("找到 %d 行\n", len(rows))

	// 节次映射 (第一二节->0, 第三四节->1, 第五六节->2, 第七八节->3)
	periodMap := map[string]int{
		"第一二节": 0,
		"第三四节": 1,
		"第五六节": 2,
		"第七八节": 3,
	}

	// 遍历行（跳过表头）
	for i, row := range rows {
		if i == 0 {
			continue // 跳过表头
		}
		rowContent := row[1]

		// 判断是哪一节
		periodIdx := -1
		for periodName, idx := range periodMap {
			if strings.Contains(rowContent, periodName) {
				periodIdx = idx
				break
			}
		}

		if periodIdx == -1 {
			continue
		}

		// 查找所有单元格 (周一到周日，但只取前5天)
		cellRegex := regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)
		cells := cellRegex.FindAllStringSubmatch(rowContent, -1)

		// 遍历周一到周五 (前5个单元格)
		for dayIdx := 0; dayIdx < 5 && dayIdx < len(cells); dayIdx++ {
			cell := cells[dayIdx][1]
			// 检查是否有课程
			hasClass := c.checkHasClass(cell)
			schedule[dayIdx][periodIdx] = hasClass
		}
	}

	return schedule
}

// checkHasClass 检查单元格是否有课
func (c *EduCrawler) checkHasClass(cell string) bool {
	// 只查找没有 display:none 的 kbcontent 或 kbcontent1 div
	// class 后面可能有空格，如 class="kbcontent1" 或 class="kbcontent1" 
	visibleDivRegex := regexp.MustCompile(`class\s*=\s*"kbcontent1?"\s*[^>]*>(.*?)</div>`)
	matches := visibleDivRegex.FindAllStringSubmatch(cell, -1)
	
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		// 检查这个div是否有 display:none
		divStart := match[0]
		if strings.Contains(divStart, `display:none`) || strings.Contains(divStart, `display: none`) {
			continue
		}
		
		content := match[1]
		// 清理HTML标签
		cleanRegex := regexp.MustCompile(`<[^>]+>`)
		text := cleanRegex.ReplaceAllString(content, "")
		
		// 解码HTML实体
		text = strings.ReplaceAll(text, "&nbsp;", " ")
		text = strings.TrimSpace(text)
		
		// 如果有中文课程名，说明有课（检查是否包含中文字符）
		if hasChinese(text) && text != "" {
			return true
		}
	}
	
	return false
}

// hasChinese 检查字符串是否包含中文字符
func hasChinese(s string) bool {
	for _, r := range s {
		if r >= '\u4e00' && r <= '\u9fa5' {
			return true
		}
	}
	return false
}

// ConvertToAvailability 将有课表转换为无课表（取反）
// 输入: week -> [weekday][period]bool (true表示有课)
// 输出: 无课时间段列表
func ConvertToAvailability(userID int, hasClassMap map[int][5][4]bool) []model.Availability {
	var availabilities []model.Availability

	for week, schedule := range hasClassMap {
		for weekday := 0; weekday < 5; weekday++ {
			for period := 0; period < 4; period++ {
				// 取反: 有课=false, 无课=true
				if !schedule[weekday][period] {
					availabilities = append(availabilities, model.Availability{
						UserID:  userID,
						Week:    week,
						Weekday: weekday + 1, // 1-5
						Period:  period + 1,  // 1-4
					})
				}
			}
		}
	}

	return availabilities
}

// CrawlRequest 爬虫请求参数
type CrawlRequest struct {
	Cookies   string `json:"cookies" binding:"required"`
	Semester  string `json:"semester" binding:"required"`
	StartWeek int    `json:"start_week"` // 可选，默认1
	EndWeek   int    `json:"end_week"`   // 可选，默认30
}

// CrawlResult 爬虫结果
type CrawlResult struct {
	WeeksParsed    int                    `json:"weeks_parsed"`
	TotalCells     int                    `json:"total_cells"`
	AvailableCells int                    `json:"available_cells"`
	Schedule       map[int][5][4]bool     `json:"schedule,omitempty"`
	Error          string                 `json:"error,omitempty"`
}
