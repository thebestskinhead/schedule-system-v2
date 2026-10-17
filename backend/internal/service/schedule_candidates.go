package service

import (
	"errors"
	"fmt"
	"schedule-system-v2/backend/internal/model"
	"strings"
)

func candidateFor(user model.User, weekday, period, maxDay, maxWeek int, cells []model.ConfirmCell, available bool) model.ScheduleCandidate {
	c := model.ScheduleCandidate{ID: user.ID, Name: user.Name, StudentID: user.StudentID, Reasons: []string{}}
	selected := false
	// Count unique slots so repeated input cannot skew rule checks.
	seen := map[[2]int]bool{}
	for _, cell := range cells {
		key := [2]int{cell.Weekday, cell.Period}
		for _, id := range cell.UserIDs {
			if id != user.ID || seen[key] {
				continue
			}
			seen[key] = true
			c.WeekCount++
			if cell.Weekday == weekday {
				c.DayCount++
			}
			if cell.Weekday == weekday && cell.Period == period {
				selected = true
			}
		}
	}
	if !available {
		c.Reasons = append(c.Reasons, "该时段有课、未填报无课时间或已请假")
	}
	if selected {
		c.Reasons = append(c.Reasons, "该时段已添加")
	}
	if c.DayCount >= maxDay {
		c.Reasons = append(c.Reasons, "已达到日最大排班次数")
	}
	if c.WeekCount >= maxWeek {
		c.Reasons = append(c.Reasons, "已达到周最大排班次数")
	}
	c.Eligible = len(c.Reasons) == 0
	return c
}

func (s *ScheduleService) GetScheduleCandidates(req *model.CandidateRequest) ([]model.ScheduleCandidate, error) {
	users, err := s.userDAO.ListByDepartment(req.Department, model.EmploymentStatusActive)
	if err != nil {
		return nil, err
	}
	available, err := s.availabilityDAO.GetAvailableUsersForSchedule(req.Week, req.Weekday, req.Period, req.Department)
	if err != nil {
		return nil, err
	}
	ids := map[int]bool{}
	for _, u := range available {
		ids[u.ID] = true
	}
	result := []model.ScheduleCandidate{}
	for _, user := range users {
		if user.DeptRole == model.RoleDeptAdmin {
			continue
		}
		result = append(result, candidateFor(user, req.Weekday, req.Period, req.MaxPerDay, req.MaxPerWeek, req.Cells, ids[user.ID]))
	}
	return result, nil
}

// Validate the whole replacement draft before any writes, using the same constraints as the picker.
func (s *ScheduleService) validateSchedule(req *model.ConfirmScheduleRequest) error {
	maxDay, maxWeek := req.MaxPerDay, req.MaxPerWeek
	if maxDay == 0 || maxWeek == 0 {
		settings, err := s.GetScheduleSettings()
		if err != nil {
			return err
		}
		if maxDay == 0 {
			maxDay = settings.MaxPerDay
		}
		if maxWeek == 0 {
			maxWeek = settings.MaxPerWeek
		}
	}
	if maxDay < 1 || maxWeek < 1 {
		return errors.New("排班次数上限必须大于0")
	}
	ids := []int{}
	for _, cell := range req.Cells {
		if cell.Weekday < 1 || cell.Weekday > 5 || cell.Period < 1 || cell.Period > 4 {
			return errors.New("排班时段无效")
		}
		ids = append(ids, cell.UserIDs...)
	}
	users, err := s.userDAO.GetUsersByIDs(ids, model.EmploymentStatusActive)
	if err != nil {
		return err
	}
	userMap := map[int]model.User{}
	for _, user := range users {
		userMap[user.ID] = user
	}
	accepted := []model.ConfirmCell{}
	availableBySlot := map[[2]int]map[int]bool{}
	for _, cell := range req.Cells {
		key := [2]int{cell.Weekday, cell.Period}
		if availableBySlot[key] == nil {
			available, err := s.availabilityDAO.GetAvailableUsers(req.Week, cell.Weekday, cell.Period)
			if err != nil {
				return err
			}
			availableBySlot[key] = map[int]bool{}
			for _, u := range available {
				availableBySlot[key][u.ID] = true
			}
		}
		for _, id := range cell.UserIDs {
			user, ok := userMap[id]
			if !ok || user.DeptRole == model.RoleDeptAdmin {
				return errors.New("包含不符合排班条件的人员")
			}
			c := candidateFor(user, cell.Weekday, cell.Period, maxDay, maxWeek, accepted, availableBySlot[key][id])
			if !c.Eligible {
				return fmt.Errorf("%s：%s", user.Name, strings.Join(c.Reasons, "；"))
			}
			accepted = append(accepted, model.ConfirmCell{Weekday: cell.Weekday, Period: cell.Period, UserIDs: []int{id}})
		}
	}
	return nil
}
