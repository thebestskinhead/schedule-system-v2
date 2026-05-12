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
	userDAO         *dao.UserDAO
}

func NewScheduleService() *ScheduleService {
	return &ScheduleService{
		availabilityDAO: dao.NewAvailabilityDAO(),
		dutyDAO:         dao.NewDutyDAO(),
		userDAO:         dao.NewUserDAO(),
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

	// ---------- 1. 预拉取阶段（3次数据库查询） ----------

	// 1a. 拉取可用性矩阵
	matrixItems, err := s.availabilityDAO.GetAvailabilityMatrix(req.Week, req.Department)
	if err != nil {
		return nil, err
	}

	// 构建：slotKey -> []userID 的映射，以及 userID 集合
	type slotKey struct{ weekday, period int }
	availMap := make(map[slotKey][]int)
	userSet := make(map[int]struct{})
	for _, item := range matrixItems {
		k := slotKey{item.Weekday, item.Period}
		availMap[k] = append(availMap[k], item.UserID)
		userSet[item.UserID] = struct{}{}
	}

	// 收集所有 userID
	userIDs := make([]int, 0, len(userSet))
	for uid := range userSet {
		userIDs = append(userIDs, uid)
	}

	// 1b. 拉取用户基础信息（ID -> User 映射）
	usersList, _ := s.userDAO.GetUsersByIDs(userIDs)
	userMap := make(map[int]model.User)
	for _, u := range usersList {
		userMap[u.ID] = u
	}

	// 1c. 拉取历史计数（全历史、本周、本周各天）
	historyTotal, _ := s.dutyDAO.BatchCountByUsers(userIDs)
	historyWeek, _ := s.dutyDAO.BatchCountByWeek(req.Week, userIDs)
	historyDay, _ := s.dutyDAO.BatchCountByWeekDay(req.Week, userIDs)

	// 构建内存只读快照
	counts := model.UserDutyCounts{
		Total: historyTotal,
		ByWeek: historyWeek,
		ByDay: historyDay,
	}

	// ---------- 2. GRASP 多启动阶段（50次迭代） ----------
	const iterations = 50
	const rclSize = 5
	const maxRepairRounds = 20

	var bestSolution *model.SchedulePreview
	var bestScore int64 = -1
	var bestConflictCount = -1

	for iter := 0; iter < iterations; iter++ {
		grid := make([][]model.Cell, 5)
		for i := range grid {
			grid[i] = make([]model.Cell, 4)
		}
		conflicts := []model.ConflictCell{}
		warnings := []model.ConflictCell{}

		// 本次预览的临时计数器
		previewWeekCount := make(map[int]int)
		previewDayCount := make(map[int]map[int]int)

		// 生成20个时段的随机顺序
		slots := make([]slotKey, 0, len(req.Days)*req.Periods)
		for _, wd := range req.Days {
			if wd < 1 || wd > 5 {
				continue
			}
			for p := 1; p <= req.Periods; p++ {
				slots = append(slots, slotKey{wd, p})
			}
		}
		rand.Shuffle(len(slots), func(i, j int) { slots[i], slots[j] = slots[j], slots[i] })

		// 逐个时段处理
		for _, sk := range slots {
			availableIDs := availMap[sk]
			if len(availableIDs) == 0 {
				// 无人可用，标记 conflict
				if req.MinPerCell > 0 {
					conflicts = append(conflicts, model.ConflictCell{
						Weekday:   sk.weekday,
						Period:    sk.period,
						Need:      req.MinPerCell,
						Available: 0,
					})
				}
				grid[sk.weekday-1][sk.period-1] = model.Cell{Weekday: sk.weekday, Period: sk.period, Users: nil}
				continue
			}

			// 过滤：MaxPerWeek / MaxPerDay
			var filtered []int
			for _, uid := range availableIDs {
				histWeek := counts.ByWeek[uid]
				prevWeek := int64(previewWeekCount[uid])
				if histWeek+prevWeek >= int64(req.MaxPerWeek) {
					continue
				}
				histDay := int64(0)
				if counts.ByDay[uid] != nil {
					histDay = counts.ByDay[uid][sk.weekday]
				}
				prevDay := int64(0)
				if previewDayCount[uid] != nil {
					prevDay = int64(previewDayCount[uid][sk.weekday])
				}
				if histDay+prevDay >= int64(req.MaxPerDay) {
					continue
				}
				filtered = append(filtered, uid)
			}

			// 排序：按全历史次数升序（严格弱序，不用随机）
			sort.Slice(filtered, func(i, j int) bool {
				return counts.Total[filtered[i]] < counts.Total[filtered[j]]
			})

			// RCL：取前 rclSize 人，从中随机抽取 NeedPerCell 个
			rcl := filtered
			if len(rcl) > rclSize {
				rcl = rcl[:rclSize]
			}

			need := req.NeedPerCell
			if len(rcl) < need {
				need = len(rcl)
			}

			// 随机抽取 need 个（不重复）
			perm := rand.Perm(len(rcl))
			selectedIDs := make([]int, need)
			for i := 0; i < need; i++ {
				selectedIDs[i] = rcl[perm[i]]
			}

			// 构建 selected Users（从 userMap 补全信息）
			selected := make([]model.User, need)
			for i, uid := range selectedIDs {
				if u, ok := userMap[uid]; ok {
					selected[i] = u
				} else {
					selected[i] = model.User{ID: uid}
				}
			}

			// 检查 conflict / warning
			if len(filtered) < req.MinPerCell {
				conflicts = append(conflicts, model.ConflictCell{
					Weekday:   sk.weekday,
					Period:    sk.period,
					Need:      req.MinPerCell,
					Available: len(filtered),
				})
			} else if len(filtered) < req.NeedPerCell {
				warnings = append(warnings, model.ConflictCell{
					Weekday:   sk.weekday,
					Period:    sk.period,
					Need:      req.NeedPerCell,
					Available: len(filtered),
				})
			}

			// 更新预览计数
			for _, uid := range selectedIDs {
				previewWeekCount[uid]++
				if previewDayCount[uid] == nil {
					previewDayCount[uid] = make(map[int]int)
				}
				previewDayCount[uid][sk.weekday]++
			}

			grid[sk.weekday-1][sk.period-1] = model.Cell{
				Weekday: sk.weekday,
				Period:  sk.period,
				Users:   selected,
			}
		}

		// ---------- 回溯修复 ----------
		repairRounds := 0
		for repairRounds < maxRepairRounds && len(conflicts) > 0 {
			prevConflictCount := len(conflicts)
			improved := false

			// 逐个处理 conflict 时段
			newConflicts := make([]model.ConflictCell, 0, len(conflicts))
			for _, cf := range conflicts {
				cell := &grid[cf.Weekday-1][cf.Period-1]
				currentCount := len(cell.Users)
				if currentCount >= req.MinPerCell {
					continue // 已修复
				}

				// 找因 MaxPerDay/MaxPerWeek 被过滤、且本次已排 >= 2 班的人
				found := false
				for uid := range userSet {
					// 检查该用户是否在当前 conflict 时段可用
					availableHere := false
					for _, aid := range availMap[slotKey{cf.Weekday, cf.Period}] {
						if aid == uid {
							availableHere = true
							break
						}
					}
					if !availableHere {
						continue
					}

						// 检查是否因 MaxPerDay/MaxPerWeek 被过滤
						histWeek := counts.ByWeek[uid]
						prevWeek := int64(previewWeekCount[uid])
						if histWeek+prevWeek >= int64(req.MaxPerWeek) {
							// 被 MaxPerWeek 过滤，检查是否已排 >= 2
							if previewWeekCount[uid] >= 2 {
								// 尝试撤销其旧班（优先同天）
								if s.tryRevokeAndAssign(grid, uid, cf.Weekday, cf.Period, previewWeekCount, previewDayCount, counts, req, userMap) {
									found = true
									improved = true
									break
								}
							}
							continue
						}

						histDay := int64(0)
						if counts.ByDay[uid] != nil {
							histDay = counts.ByDay[uid][cf.Weekday]
						}
						prevDay := int64(0)
						if previewDayCount[uid] != nil {
							prevDay = int64(previewDayCount[uid][cf.Weekday])
						}
						if histDay+prevDay >= int64(req.MaxPerDay) {
							// 被 MaxPerDay 过滤，检查是否已排 >= 2
							if previewWeekCount[uid] >= 2 {
								if s.tryRevokeAndAssign(grid, uid, cf.Weekday, cf.Period, previewWeekCount, previewDayCount, counts, req, userMap) {
									found = true
									improved = true
									break
								}
							}
							continue
						}
				}

				if !found {
					newConflicts = append(newConflicts, cf)
				}
			}

			conflicts = newConflicts
			repairRounds++

			// 三重保险：无改善 或 冲突数未降 则退出
			if !improved || len(conflicts) >= prevConflictCount {
				break
			}
		}

		// ---------- 评估 ----------
		score := int64(0)
		for uid, cnt := range previewWeekCount {
			score += counts.Total[uid] + int64(cnt)
		}

		// ---------- 记录最优 ----------
		candidate := &model.SchedulePreview{
			Week:      req.Week,
			Grid:      grid,
			Conflicts: conflicts,
			Warnings:  warnings,
		}

		if len(conflicts) == 0 {
			// 零冲突解 -> 最优可行解（取 score 最小，即负担最均衡）
			if bestSolution == nil || len(bestSolution.Conflicts) > 0 || score < bestScore {
				bestSolution = candidate
				bestScore = score
				bestConflictCount = 0
			}
		} else {
			// 非零冲突 -> 只记录冲突最少的次优解
			if bestSolution == nil || len(conflicts) < bestConflictCount {
				bestSolution = candidate
				bestScore = score
				bestConflictCount = len(conflicts)
			}
		}
	}

	// ---------- 3. 输出 ----------
	if bestSolution == nil {
		// 兜底：理论上不会发生，但保留安全网
		return &model.SchedulePreview{
			Week:      req.Week,
			Grid:      make([][]model.Cell, 5),
			Conflicts: []model.ConflictCell{},
			Warnings:  []model.ConflictCell{},
		}, nil
	}
	return bestSolution, nil
}

// tryRevokeAndAssign 尝试撤销某用户的旧班并填入新空档
func (s *ScheduleService) tryRevokeAndAssign(
	grid [][]model.Cell,
	uid int,
	targetWeekday, targetPeriod int,
	previewWeekCount map[int]int,
	previewDayCount map[int]map[int]int,
	counts model.UserDutyCounts,
	req *model.ScheduleRequest,
	userMap map[int]model.User,
) bool {
	// 优先撤销同天的旧班
	for p := 1; p <= 4; p++ {
		if p == targetPeriod {
			continue
		}
		cell := &grid[targetWeekday-1][p-1]
		for i, u := range cell.Users {
			if u.ID == uid {
				// 撤销
				cell.Users = append(cell.Users[:i], cell.Users[i+1:]...)
				previewWeekCount[uid]--
				previewDayCount[uid][targetWeekday]--

							// 填入目标空档
							targetCell := &grid[targetWeekday-1][targetPeriod-1]
							if u, ok := userMap[uid]; ok {
								targetCell.Users = append(targetCell.Users, u)
							} else {
								targetCell.Users = append(targetCell.Users, model.User{ID: uid})
							}
							previewWeekCount[uid]++
							previewDayCount[uid][targetWeekday]++
							return true
			}
		}
	}

	// 同天没有，尝试撤销其他天的旧班
	for wd := 1; wd <= 5; wd++ {
		if wd == targetWeekday {
			continue
		}
		for p := 1; p <= 4; p++ {
			cell := &grid[wd-1][p-1]
			for i, u := range cell.Users {
				if u.ID == uid {
					// 撤销后检查 MaxPerDay 约束
					if previewDayCount[uid] != nil && previewDayCount[uid][targetWeekday] > 0 {
						// 目标天已有班，撤销其他天的班不会违反 MaxPerDay
					} else {
						// 目标天没有班，可以直接填
					}

					cell.Users = append(cell.Users[:i], cell.Users[i+1:]...)
					previewWeekCount[uid]--
					previewDayCount[uid][wd]--

						targetCell := &grid[targetWeekday-1][targetPeriod-1]
						if u, ok := userMap[uid]; ok {
							targetCell.Users = append(targetCell.Users, u)
						} else {
							targetCell.Users = append(targetCell.Users, model.User{ID: uid})
						}
						previewWeekCount[uid]++
						if previewDayCount[uid] == nil {
							previewDayCount[uid] = make(map[int]int)
						}
						previewDayCount[uid][targetWeekday]++
						return true
				}
			}
		}
	}

	return false
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

	// 同步当前周次（自动计算或递增）
	s.SyncCurrentWeek()

	return nil
}

// SyncCurrentWeek 同步当前周次
// 若存在学期起始日且开启自动计算，将 current_week 同步为计算值；
// 若无学期起始日且开启自动计算，则 current_week + 1。
func (s *ScheduleService) SyncCurrentWeek() error {
	settings, err := s.dutyDAO.GetScheduleSettings()
	if err != nil || settings == nil {
		return nil
	}

	if !settings.AutoIncrement {
		return nil
	}

	// 有学期起始日：同步为基于日期的计算值
	if settings.SemesterStartDate != nil && !settings.SemesterStartDate.IsZero() {
		week, err := s.CalculateCurrentWeek(settings.SemesterStartDate.Format("2006-01-02"))
		if err != nil {
			return err
		}
		return s.dutyDAO.UpdateCurrentWeekDirect(week)
	}

	// 无学期起始日：原有递增逻辑
	return s.dutyDAO.IncrementCurrentWeek()
}

func (s *ScheduleService) GetScheduleByWeek(week int) ([]model.DutyRecordWithUser, error) {
	return s.dutyDAO.GetByWeek(week)
}

func (s *ScheduleService) GetScheduleByWeekAndDepartment(week int, department string) ([]model.DutyRecordWithUser, error) {
	return s.dutyDAO.GetByWeekAndDepartment(week, department)
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
func (s *ScheduleService) GetScheduleSettings() (*model.ScheduleSettings, error) {
	return s.dutyDAO.GetScheduleSettings()
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

// GetCurrentWeek 获取当前周次
// 当 auto_increment=1 且存在学期起始日时，基于日期实时计算；
// 否则返回数据库中存储的固定 current_week。
func (s *ScheduleService) GetCurrentWeek() (int, bool, error) {
	settings, err := s.dutyDAO.GetScheduleSettings()
	if err != nil || settings == nil {
		return 1, false, nil
	}

	if settings.AutoIncrement && settings.SemesterStartDate != nil && !settings.SemesterStartDate.IsZero() {
		week, err := s.CalculateCurrentWeek(settings.SemesterStartDate.Format("2006-01-02"))
		if err == nil && week > 0 {
			return week, settings.AutoIncrement, nil
		}
	}

	return settings.CurrentWeek, settings.AutoIncrement, nil
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
