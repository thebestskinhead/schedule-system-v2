package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
)

// ApplicationExecutor 应用类型执行器接口
type ApplicationExecutor interface {
	// GetType 返回应用类型标识
	GetType() string
	// Validate 验证申请数据
	Validate(data json.RawMessage) error
	// OnCreated 申请创建后的回调
	OnCreated(ctx context.Context, app *model.Application) error
	// OnApproved 申请批准后的回调
	OnApproved(ctx context.Context, app *model.Application, approval *model.ApplicationApproval) error
	// OnRejected 申请拒绝后的回调
	OnRejected(ctx context.Context, app *model.Application, approval *model.ApplicationApproval) error
	// GetRequiredApprovers 获取所需的审批人列表
	GetRequiredApprovers(ctx context.Context, app *model.Application) ([]int, error)
	// CanApply 检查用户是否可以提交申请
	CanApply(ctx context.Context, userID int, data json.RawMessage) (bool, string)
}

// ApplicationManager 应用管理器
type ApplicationManager struct {
	executors map[string]ApplicationExecutor
}

// NewApplicationManager 创建应用管理器
func NewApplicationManager() *ApplicationManager {
	return &ApplicationManager{
		executors: make(map[string]ApplicationExecutor),
	}
}

// Register 注册应用执行器
func (m *ApplicationManager) Register(executor ApplicationExecutor) {
	m.executors[executor.GetType()] = executor
}

// GetExecutor 获取执行器
func (m *ApplicationManager) GetExecutor(appType string) (ApplicationExecutor, bool) {
	executor, ok := m.executors[appType]
	return executor, ok
}

// ApplicationService 应用服务
type ApplicationService struct {
	applicationDao *dao.ApplicationDao
	userDao        *dao.UserDAO
	manager        *ApplicationManager
}

// NewApplicationService 创建应用服务
func NewApplicationService(applicationDao *dao.ApplicationDao, userDao *dao.UserDAO, manager *ApplicationManager) *ApplicationService {
	return &ApplicationService{
		applicationDao: applicationDao,
		userDao:        userDao,
		manager:        manager,
	}
}

// CreateApplication 创建申请
func (s *ApplicationService) CreateApplication(ctx context.Context, userID int, appType string, data json.RawMessage, reason string) (*model.Application, error) {
	// 获取执行器
	executor, ok := s.manager.GetExecutor(appType)
	if !ok {
		return nil, fmt.Errorf("未知的申请类型: %s", appType)
	}

	// 检查用户是否可以申请
	if can, msg := executor.CanApply(ctx, userID, data); !can {
		return nil, errors.New(msg)
	}

	// 验证数据
	if err := executor.Validate(data); err != nil {
		return nil, fmt.Errorf("数据验证失败: %v", err)
	}

	// 获取申请类型配置
	appTypeConfig, err := s.applicationDao.GetApplicationType(ctx, appType)
	if err != nil {
		return nil, err
	}

	// 创建申请对象
	now := time.Now()
	// 记录申请人当前部门：部门级权限授予按此推导资源范围，缺失会导致范围退化为 0（全部门通配）
	applicant, err := s.userDao.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取申请人信息失败: %v", err)
	}
	application := &model.Application{
		TypeCode:    appType,
		ApplicantID: userID,
		Department:  applicant.Department,
		Status:      model.ApplicationStatusPending,
		Data:        data,
		Title:       reason, // 使用 reason 作为标题
		Content:     reason,
		CurrentLevel: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 解析配置获取默认审批流程（须在落库前完成，TotalLevels 随首次 INSERT 一并写入）
	var typeConfig model.TypeConfig
	if len(appTypeConfig.Config) > 0 {
		if err := json.Unmarshal(appTypeConfig.Config, &typeConfig); err != nil {
			// 解析失败使用默认配置
			typeConfig = model.TypeConfig{
				Fields: []model.FieldConfig{},
				Flow:   []model.FlowStep{{Level: 1, Role: "dept_admin", Label: "部门管理员审批"}},
			}
		}
	} else {
		// 没有配置时使用默认配置
		typeConfig = model.TypeConfig{
			Fields: []model.FieldConfig{},
			Flow:   []model.FlowStep{{Level: 1, Role: "dept_admin", Label: "部门管理员审批"}},
		}
	}

	// 计算总审批层级
	if len(typeConfig.Flow) > 0 {
		application.TotalLevels = len(typeConfig.Flow)
	} else {
		application.TotalLevels = 1
	}

	// 生成申请编号并落库（撞唯一键时应用层重试 ≤3 次，见 G 项）
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		appNo, err := s.applicationDao.GenerateApplicationNo(ctx, appType)
		if err != nil {
			return nil, err
		}
		application.ApplicationNo = appNo
		if err := s.applicationDao.Create(ctx, application); err != nil {
			if db.IsDuplicateEntry(err) {
				lastErr = err
				continue // 唯一键冲突：重新生成编号后重试
			}
			return nil, err
		}
		lastErr = nil
		break
	}
	if lastErr != nil {
		return nil, fmt.Errorf("生成申请编号失败（唯一键冲突）: %w", lastErr)
	}

	// 审批流程配置在 application_types 表中，不需要为每个申请创建审批人配置
	// 实际审批人由 GetRequiredApprovers 方法根据配置动态计算

	// 获取审批人列表
	approverIDs, err := executor.GetRequiredApprovers(ctx, application)
	if err != nil {
		return nil, err
	}

	// 发送审批通知（异步）
	go s.notifyApprovers(approverIDs, application)

	// 回调
	if err := executor.OnCreated(ctx, application); err != nil {
		// 记录日志但不返回错误
		fmt.Printf("OnCreated callback error: %v\n", err)
	}

	return application, nil
}

// GetMyApplications 获取我的申请列表
func (s *ApplicationService) GetMyApplications(ctx context.Context, userID int, status string, page, pageSize int) ([]model.Application, int, error) {
	// 将字符串状态转换为 int，空字符串表示不筛选
	statusInt := -1
	if status != "" {
		if v, err := strconv.Atoi(status); err == nil {
			statusInt = v
		}
	}
	return s.applicationDao.GetByApplicant(ctx, userID, statusInt, page, pageSize)
}

// GetPendingApprovals 获取待我审批的申请
func (s *ApplicationService) GetPendingApprovals(ctx context.Context, userID int, isAdmin bool, department string, page, pageSize int) ([]model.Application, int, error) {
	return s.applicationDao.GetPendingByApprover(ctx, userID, isAdmin, department, page, pageSize)
}

// GetApplicationDetail 获取申请详情
func (s *ApplicationService) GetApplicationDetail(ctx context.Context, appID int, userID int, isAdmin bool) (*model.Application, error) {
	app, err := s.applicationDao.GetByID(ctx, appID)
	if err != nil {
		return nil, err
	}

	// 检查权限：申请人自己、审批人或管理员可以查看
	if app.ApplicantID != userID && !isAdmin {
		// 审批人可查看（复用审批权限判定：按申请类型流程配置 + 身份 + 部门）
		canView, err := s.canApprove(ctx, app, userID)
		if err != nil {
			return nil, err
		}
		if !canView {
			return nil, fmt.Errorf("无权查看此申请")
		}
	}

	return app, nil
}

// ProcessApproval 处理审批（使用乐观锁防止并发问题）
func (s *ApplicationService) ProcessApproval(ctx context.Context, appID, approverID int, action model.ApprovalAction, comment string) error {
	// 获取申请
	app, err := s.applicationDao.GetByID(ctx, appID)
	if err != nil {
		return err
	}

	// 检查状态（乐观锁检查点1）
	if app.Status != model.ApplicationStatusPending && app.Status != model.ApplicationStatusProcessing {
		return fmt.Errorf("申请当前状态不可审批")
	}

	// 检查权限
	canApprove, err := s.canApprove(ctx, app, approverID)
	if err != nil {
		return err
	}
	if !canApprove {
		return fmt.Errorf("无权审批此申请")
	}

	// 获取执行器
	executor, ok := s.manager.GetExecutor(app.TypeCode)
	if !ok {
		return fmt.Errorf("未知的申请类型: %s", app.TypeCode)
	}

	// 仅评论：不改变状态，仅插入审批记录（SP 的 UPDATE 无变更会误触 SIGNAL，故单独处理）
	if action == model.ApprovalActionComment {
		approval := &model.ApplicationApproval{
			ApplicationID: appID,
			ApproverID:    approverID,
			Action:        action,
			Comment:       comment,
			Level:         app.CurrentLevel,
			CreatedAt:     time.Now(),
		}
		if err := s.applicationDao.CreateApproval(ctx, approval); err != nil {
			return err
		}
		return nil
	}

	// 计算期望状态（乐观锁）与目标状态
	expectedStatus := int(app.Status)
	var newStatus int
	switch action {
	case model.ApprovalActionApprove:
		if app.CurrentLevel < app.TotalLevels {
			newStatus = int(model.ApplicationStatusProcessing)
		} else {
			newStatus = int(model.ApplicationStatusApproved)
		}
	case model.ApprovalActionReject:
		newStatus = int(model.ApplicationStatusRejected)
	default:
		return fmt.Errorf("不支持的审批操作")
	}

	// 原子推进：写审批记录 + 更新状态（乐观锁），替代原「创建审批记录 + 更新状态」两段
	spErr := db.WithRetry(ctx, func() error {
		_, e := db.GetDB().Exec("CALL sp_advance_application(?, ?, ?, ?, ?, ?)",
			appID, approverID, int(action), comment, expectedStatus, newStatus)
		return e
	})
	if spErr != nil {
		if biz, ok := db.BusinessErrorFromSQLError(spErr); ok {
			return biz
		}
		return spErr
	}

	// 多级审批：状态进入 Processing 时推进 current_level 以路由下一级
	// （注：sp_advance_application 不维护 current_level，此处为 SP 之外的必要补充更新）
	if action == model.ApprovalActionApprove && app.CurrentLevel < app.TotalLevels {
		app.CurrentLevel++
		app.Status = model.ApplicationStatusProcessing
		app.UpdatedAt = time.Now()
		if _, e := db.GetDB().Exec("UPDATE applications SET current_level = ?, updated_at = ? WHERE id = ?",
			app.CurrentLevel, app.UpdatedAt, appID); e != nil {
			fmt.Printf("advance current_level error: %v\n", e)
		}
		nextApprovers, err := executor.GetRequiredApprovers(ctx, app)
		if err == nil {
			go s.notifyApprovers(nextApprovers, app)
		}
		return nil
	}

	// 最终批准/拒绝：SP 提交成功后执行 executor 回调（失败告警不阻断）
	approval := &model.ApplicationApproval{
		ApplicationID: appID,
		ApproverID:    approverID,
		Action:        action,
		Comment:       comment,
		Level:         app.CurrentLevel,
		CreatedAt:     time.Now(),
	}
	if action == model.ApprovalActionApprove {
		if err := executor.OnApproved(ctx, app, approval); err != nil {
			fmt.Printf("OnApproved callback error: %v\n", err)
		}
	} else if action == model.ApprovalActionReject {
		if err := executor.OnRejected(ctx, app, approval); err != nil {
			fmt.Printf("OnRejected callback error: %v\n", err)
		}
	}

	return nil
}

// canApprove 检查用户是否可以审批
func (s *ApplicationService) canApprove(ctx context.Context, app *model.Application, userID int) (bool, error) {
	// 获取用户信息
	user, err := s.userDao.GetByID(userID)
	if err != nil {
		return false, err
	}

	// 系统管理员可以审批所有
	if user.Role == model.RoleAdmin {
		return true, nil
	}

	// 获取申请类型配置
	appType, err := s.applicationDao.GetApplicationType(ctx, app.TypeCode)
	if err != nil {
		return false, err
	}

	var typeConfig model.TypeConfig
	if len(appType.Config) > 0 {
		if err := json.Unmarshal(appType.Config, &typeConfig); err != nil {
			// 解析失败使用默认配置
			typeConfig = model.TypeConfig{
				Flow: []model.FlowStep{{Level: 1, Role: "dept_admin", Label: "部门管理员审批"}},
			}
		}
	} else {
		typeConfig = model.TypeConfig{
			Flow: []model.FlowStep{{Level: 1, Role: "dept_admin", Label: "部门管理员审批"}},
		}
	}

	// 获取当前审批层级配置
	var currentStep *model.FlowStep
	for _, step := range typeConfig.Flow {
		if step.Level == app.CurrentLevel {
			currentStep = &step
			break
		}
	}

	if currentStep == nil {
		// 使用默认逻辑：部门管理员可以审批本部门申请
		applicant, err := s.userDao.GetByID(app.ApplicantID)
		if err != nil {
			return false, err
		}
		if user.DeptRole != model.RoleDeptAdmin || user.Department != applicant.Department {
			return false, nil
		}
		// 权限范围检查：部门管理员不能审批全局权限申请
		if app.TypeCode == "temp_permission" && len(app.Data) > 0 {
			var permData struct {
				Permission string `json:"permission"`
			}
			if err := json.Unmarshal(app.Data, &permData); err == nil {
				if isGlobalPermission(permData.Permission) {
					return false, nil
				}
			}
		}
		return true, nil
	}

	// 检查审批角色
	switch currentStep.Role {
	case "admin":
		return user.Role == model.RoleAdmin, nil
	case "dept_admin":
		if user.DeptRole != model.RoleDeptAdmin {
			return false, nil
		}
		// 部门管理员可以审批本部门申请
		applicant, err := s.userDao.GetByID(app.ApplicantID)
		if err != nil {
			return false, err
		}
		if user.Department != applicant.Department {
			return false, nil
		}
		// 权限范围检查：部门管理员不能审批全局权限申请
		if app.TypeCode == "temp_permission" && len(app.Data) > 0 {
			var permData struct {
				Permission string `json:"permission"`
			}
			if err := json.Unmarshal(app.Data, &permData); err == nil {
				if isGlobalPermission(permData.Permission) {
					return false, nil
				}
			}
		}
		return true, nil
	case "office_admin":
		return user.Department == model.DepartmentOffice && user.DeptRole == model.RoleDeptAdmin, nil
	default:
		return false, nil
	}
}

// CancelApplication 取消申请（使用乐观锁防止并发问题）
func (s *ApplicationService) CancelApplication(ctx context.Context, appID, userID int) error {
	app, err := s.applicationDao.GetByID(ctx, appID)
	if err != nil {
		return err
	}

	// 只能取消自己的申请
	if app.ApplicantID != userID {
		return fmt.Errorf("只能取消自己的申请")
	}

	// 只能取消待审批的申请（乐观锁：检查状态）
	if app.Status != model.ApplicationStatusPending {
		return fmt.Errorf("只能取消待审批的申请")
	}

	// 使用乐观锁更新状态：只有状态仍为 Pending 时才更新
	success, err := s.applicationDao.UpdateStatusWithCheck(ctx, appID, model.ApplicationStatusPending, model.ApplicationStatusWithdrawn)
	if err != nil {
		return err
	}
	if !success {
		// 状态已被其他操作修改，说明申请已被处理
		return fmt.Errorf("申请已被处理，无法撤回")
	}

	return nil
}

// GetApplicationTypes 获取所有申请类型
func (s *ApplicationService) GetApplicationTypes(ctx context.Context) ([]model.ApplicationType, error) {
	return s.applicationDao.GetAllTypes(ctx)
}

// notifyApprovers 通知审批人（简化实现，可扩展为邮件/短信通知）
func (s *ApplicationService) notifyApprovers(approverIDs []int, app *model.Application) {
	// TODO: 实现通知逻辑，如发送邮件或站内信
	fmt.Printf("通知审批人 %v 处理申请 %s\n", approverIDs, app.ApplicationNo)
}

// GetApplicationStats 获取申请统计
func (s *ApplicationService) GetApplicationStats(ctx context.Context, userID int, isAdmin bool, department string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 我的申请统计
	myPending, err := s.applicationDao.CountByApplicantAndStatus(ctx, userID, int(model.ApplicationStatusPending))
	if err != nil {
		return nil, err
	}
	myProcessing, err := s.applicationDao.CountByApplicantAndStatus(ctx, userID, int(model.ApplicationStatusProcessing))
	if err != nil {
		return nil, err
	}
	myApproved, err := s.applicationDao.CountByApplicantAndStatus(ctx, userID, int(model.ApplicationStatusApproved))
	if err != nil {
		return nil, err
	}
	myRejected, err := s.applicationDao.CountByApplicantAndStatus(ctx, userID, int(model.ApplicationStatusRejected))
	if err != nil {
		return nil, err
	}

	stats["my_applications"] = map[string]int{
		"pending":    myPending,
		"processing": myProcessing,
		"approved":   myApproved,
		"rejected":   myRejected,
	}

	// 待我审批统计
	if isAdmin || department != "" {
		pendingApproval, err := s.applicationDao.CountPendingByApprover(ctx, userID, isAdmin, department)
		if err != nil {
			return nil, err
		}
		stats["pending_approval"] = pendingApproval
	}

	return stats, nil
}
