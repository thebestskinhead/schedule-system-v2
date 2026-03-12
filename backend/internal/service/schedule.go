package service

import (
	"errors"
	"math/rand"
	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/model"
	"sort"
	"time"
)

type ScheduleService struct {
	availabilityDAO *dao.AvailabilityDAO
	dutyDAO         *dao.DutyDAO
}

func NewScheduleService() *ScheduleService {
	return &ScheduleService{
		availabilityDAO: dao.NewAvailabilityDAO(),
		dutyDAO:         dao.NewDutyDAO(),
	}
}

func (s *ScheduleService) PreviewSchedule(req *model.ScheduleRequest) (*model.SchedulePreview, error) {
	rand.Seed(time.Now().UnixNano())

	// 设置默认值
	if req.MinPerCell < 0 {
		req.MinPerCell = 0
	}
	if req.MaxPerDay <= 0 {
		req.MaxPerDay = 1
	}
	if req.MaxPerWeek <= 0 {
		req.MaxPerWeek = 2
	}

	grid := make([][]model.Cell, 5)
	for i := range grid {
		grid[i] = make([]model.Cell, 4)
	}

	conflicts := []model.ConflictCell{}
	warnings := []model.ConflictCell{} // 警告（人数不足但未达最小要求）

	// 跟踪本次预览中每个用户的排班次数
	// weekCount[userID] = 本周已排班次数
	// dayCount[userID][weekday] = 当天已排班次数
	previewWeekCount := make(map[int]int)
	previewDayCount := make(map[int]map[int]int)

	for _, weekday := range req.Days {
		if weekday < 1 || weekday > 5 {
			continue
		}
		for period := 1; period <= req.Periods; period++ {
			// 使用GetAvailableUsersForSchedule按部门筛选用户
			available, err := s.availabilityDAO.GetAvailableUsersForSchedule(req.Week, weekday, period, req.Department)
			if err != nil {
				continue
			}

			counters := make(map[int]int64)
			for _, user := range available {
				count, _ := s.dutyDAO.CountByUser(user.ID)
				counters[user.ID] = count
			}

			sort.Slice(available, func(i, j int) bool {
				ci, cj := counters[available[i].ID], counters[available[j].ID]
				if ci != cj {
					return ci < cj
				}
				return rand.Intn(2) == 0
			})

			var filtered []model.User
			for _, user := range available {
				// 检查历史排班 + 本次预览已分配的次数
				historyWeekCount, _ := s.dutyDAO.CountByUserAndWeek(user.ID, req.Week)
				previewWeek := previewWeekCount[user.ID]
				totalWeekCount := int(historyWeekCount) + previewWeek
				if totalWeekCount >= req.MaxPerWeek {
					continue
				}

				historyDayCount, _ := s.dutyDAO.CountByUserAndDay(user.ID, req.Week, weekday)
				previewDay := 0
				if previewDayCount[user.ID] != nil {
					previewDay = previewDayCount[user.ID][weekday]
				}
				totalDayCount := int(historyDayCount) + previewDay
				if totalDayCount >= req.MaxPerDay {
					continue
				}
				filtered = append(filtered, user)
			}

			var selected []model.User
			if len(filtered) >= req.NeedPerCell {
				selected = filtered[:req.NeedPerCell]
			} else {
				selected = filtered
				// 检查是否达到最小人数要求
				if len(selected) < req.MinPerCell {
					conflicts = append(conflicts, model.ConflictCell{
						Weekday:    weekday,
						Period:     period,
						Need:       req.MinPerCell,
						Available:  len(filtered),
					})
				} else if len(selected) < req.NeedPerCell {
					// 未达到最大人数但达到最小人数，只警告
					warnings = append(warnings, model.ConflictCell{
						Weekday:    weekday,
						Period:     period,
						Need:       req.NeedPerCell,
						Available:  len(filtered),
					})
				}
			}

			// 更新预览计数
			for _, user := range selected {
				previewWeekCount[user.ID]++
				if previewDayCount[user.ID] == nil {
					previewDayCount[user.ID] = make(map[int]int)
				}
				previewDayCount[user.ID][weekday]++
			}

			grid[weekday-1][period-1] = model.Cell{
				Weekday: weekday,
				Period:  period,
				Users:   selected,
			}
		}
	}

	return &model.SchedulePreview{
		Week:      req.Week,
		Grid:      grid,
		Conflicts: conflicts,
		Warnings:  warnings,
	}, nil
}

func (s *ScheduleService) ConfirmSchedule(adminID int, req *model.ConfirmScheduleRequest) error {
	if len(req.Cells) == 0 {
		return errors.New("无排班数据")
	}

	s.dutyDAO.DeleteByWeek(req.Week)

	var records []model.DutyRecord
	for _, cell := range req.Cells {
		for _, userID := range cell.UserIDs {
			records = append(records, model.DutyRecord{
				Week:       req.Week,
				Weekday:    cell.Weekday,
				Period:     cell.Period,
				UserID:     userID,
				AssignedBy: adminID,
				Status:     "pending",
			})
		}
	}

	if err := s.dutyDAO.CreateBatch(records); err != nil {
		return err
	}

	// 如果开启了自动递增，尝试增加当前周次
	s.dutyDAO.IncrementCurrentWeek()

	return nil
}

func (s *ScheduleService) GetScheduleByWeek(week int) ([]model.DutyRecordWithUser, error) {
	return s.dutyDAO.GetByWeek(week)
}

func (s *ScheduleService) GetMyDuties(userID int) ([]model.DutyRecordWithUser, error) {
	return s.dutyDAO.GetByUserID(userID)
}

func (s *ScheduleService) UpdateDutyStatus(userID int, dutyID int, status string) error {
	duty, err := s.dutyDAO.GetByID(dutyID)
	if err != nil {
		return errors.New("记录不存在")
	}

	if duty.UserID != userID {
		return errors.New("无权操作")
	}

	validStatuses := map[string]bool{"pending": true, "confirmed": true, "completed": true, "cancelled": true}
	if !validStatuses[status] {
		return errors.New("无效的状态")
	}

	return s.dutyDAO.UpdateStatus(dutyID, status)
}

// GetScheduleSettings 获取排班设置
func (s *ScheduleService) GetScheduleSettings(adminID int) (*model.ScheduleSettings, error) {
	return s.dutyDAO.GetScheduleSettings(adminID)
}

// SaveScheduleSettings 保存排班设置
func (s *ScheduleService) SaveScheduleSettings(adminID int, settings *model.ScheduleSettings) error {
	settings.AdminID = adminID
	return s.dutyDAO.SaveScheduleSettings(settings)
}

// UpdateSchedule 更新排班（添加/删除人员）
func (s *ScheduleService) UpdateSchedule(adminID int, req *model.UpdateScheduleRequest) error {
	// 删除人员
	for _, userID := range req.RemoveUserIDs {
		if err := s.dutyDAO.DeleteByWeekdayPeriodAndUser(req.Week, req.Weekday, req.Period, userID); err != nil {
			return err
		}
	}

	// 添加人员
	for _, userID := range req.AddUserIDs {
		record := model.DutyRecord{
			Week:       req.Week,
			Weekday:    req.Weekday,
			Period:     req.Period,
			UserID:     userID,
			AssignedBy: adminID,
			Status:     "pending",
		}
		if err := s.dutyDAO.Create(&record); err != nil {
			return err
		}
	}

	return nil
}

// GetCurrentWeek 获取当前周次（优先根据学期起始日计算）
func (s *ScheduleService) GetCurrentWeek() (int, bool, error) {
	// 先尝试从学期起始日计算当前周次
	startDate, err := s.dutyDAO.GetSemesterStartDate()
	if err == nil && startDate != nil && *startDate != "" {
		week, err := s.CalculateCurrentWeek(*startDate)
		if err == nil && week > 0 {
			// 获取自动递增设置
			_, autoIncrement, _ := s.dutyDAO.GetCurrentWeek()
			return week, autoIncrement, nil
		}
	}
	
	// 如果没有设置学期起始日，返回数据库中存储的当前周次
	return s.dutyDAO.GetCurrentWeek()
}

// UpdateCurrentWeek 更新当前周次
func (s *ScheduleService) UpdateCurrentWeek(adminID int, currentWeek int, autoIncrement bool) error {
	return s.dutyDAO.UpdateCurrentWeek(adminID, currentWeek, autoIncrement)
}

// IncrementCurrentWeek 当前周次自动递增
func (s *ScheduleService) IncrementCurrentWeek() error {
	return s.dutyDAO.IncrementCurrentWeek()
}

// GetSemesterStartDate 获取学期起始日
func (s *ScheduleService) GetSemesterStartDate() (*string, error) {
	return s.dutyDAO.GetSemesterStartDate()
}

// UpdateSemesterStartDate 更新学期起始日
func (s *ScheduleService) UpdateSemesterStartDate(adminID int, startDate string) error {
	return s.dutyDAO.UpdateSemesterStartDate(adminID, startDate)
}

// CalculateCurrentWeek 根据学期起始日计算当前周次
func (s *ScheduleService) CalculateCurrentWeek(startDate string) (int, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	days := int(now.Sub(start).Hours() / 24)
	if days < 0 {
		return 1, nil
	}
	week := days/7 + 1
	if week > 30 {
		return 30, nil
	}
	return week, nil
}
