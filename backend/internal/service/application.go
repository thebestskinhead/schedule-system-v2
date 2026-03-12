package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"schedule-system-v2/backend/internal/dao"
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
		return nil, fmt.Errorf(msg)
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

	// 生成申请编号
	appNo, err := s.applicationDao.GenerateApplicationNo(ctx, appType)
	if err != nil {
		return nil, err
	}

	// 解析配置获取默认审批流程
	var typeConfig model.TypeConfig
	if len(appTypeConfig.Config) > 0 {
		if err := json.Unmarshal(appTypeConfig.Config, &typeConfig); err != nil {
			// 如果解析失败，使用默认配置
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

	// 创建申请
	now := time.Now()
	application := &model.Application{
		ApplicationNo: appNo,
		TypeCode:      appType,
		ApplicantID:   userID,
		Status:        model.ApplicationStatusPending,
		Data:          data,
		Title:         reason, // 使用 reason 作为标题
		Content:       reason,
		CurrentLevel:  1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// 计算总审批层级
	if len(typeConfig.Flow) > 0 {
		application.TotalLevels = len(typeConfig.Flow)
	} else {
		application.TotalLevels = 1
	}

	// 保存申请
	if err := s.applicationDao.Create(ctx, application); err != nil {
		return nil, err
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
	return s.applicationDao.GetByApplicant(ctx, userID, status, page, pageSize)
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
		// 检查是否是审批人
		isApprover, err := s.applicationDao.IsApproverForApplication(ctx, appID, userID)
		if err != nil {
			return nil, err
		}
		if !isApprover {
			return nil, fmt.Errorf("无权查看此申请")
		}
	}

	return app, nil
}

// ProcessApproval 处理审批
func (s *ApplicationService) ProcessApproval(ctx context.Context, appID, approverID int, action model.ApprovalAction, comment string) error {
	// 获取申请
	app, err := s.applicationDao.GetByID(ctx, appID)
	if err != nil {
		return err
	}

	// 检查状态
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

	// 创建审批记录
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

	// 处理审批结果
	now := time.Now()
	switch action {
	case model.ApprovalActionApprove:
		// 是否还有下一级审批
		if app.CurrentLevel < app.TotalLevels {
			// 进入下一级
			app.CurrentLevel++
			app.Status = model.ApplicationStatusProcessing
			app.UpdatedAt = now
			if err := s.applicationDao.Update(ctx, app); err != nil {
				return err
			}

			// 通知下一级审批人
			nextApprovers, err := executor.GetRequiredApprovers(ctx, app)
			if err != nil {
				return err
			}
			go s.notifyApprovers(nextApprovers, app)
		} else {
			// 最终批准
			app.Status = model.ApplicationStatusApproved
			app.UpdatedAt = now
			if err := s.applicationDao.Update(ctx, app); err != nil {
				return err
			}

			// 执行批准回调
			if err := executor.OnApproved(ctx, app, approval); err != nil {
				fmt.Printf("OnApproved callback error: %v\n", err)
			}
		}

	case model.ApprovalActionReject:
		app.Status = model.ApplicationStatusRejected
		app.UpdatedAt = now
		if err := s.applicationDao.Update(ctx, app); err != nil {
			return err
		}

		// 执行拒绝回调
		if err := executor.OnRejected(ctx, app, approval); err != nil {
			fmt.Printf("OnRejected callback error: %v\n", err)
		}

	case model.ApprovalActionComment:
		// 仅评论，不改变状态
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
	if user.Role == "admin" {
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
		return user.DeptRole == "dept_admin" && user.Department == applicant.Department, nil
	}

	// 检查审批角色
	switch currentStep.Role {
	case "admin":
		return user.Role == "admin", nil
	case "dept_admin":
		if user.DeptRole != "dept_admin" {
			return false, nil
		}
		// 部门管理员可以审批本部门申请
		applicant, err := s.userDao.GetByID(app.ApplicantID)
		if err != nil {
			return false, err
		}
		return user.Department == applicant.Department, nil
	case "office_admin":
		return user.Department == "办公室" && user.DeptRole == "dept_admin", nil
	default:
		return false, nil
	}
}

// CancelApplication 取消申请
func (s *ApplicationService) CancelApplication(ctx context.Context, appID, userID int) error {
	app, err := s.applicationDao.GetByID(ctx, appID)
	if err != nil {
		return err
	}

	// 只能取消自己的申请
	if app.ApplicantID != userID {
		return fmt.Errorf("只能取消自己的申请")
	}

	// 只能取消待审批的申请
	if app.Status != model.ApplicationStatusPending {
		return fmt.Errorf("只能取消待审批的申请")
	}

	app.Status = model.ApplicationStatusWithdrawn
	app.UpdatedAt = time.Now()
	return s.applicationDao.Update(ctx, app)
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
	myPending, err := s.applicationDao.CountByApplicantAndStatus(ctx, userID, string(model.ApplicationStatusPending))
	if err != nil {
		return nil, err
	}
	myApproved, err := s.applicationDao.CountByApplicantAndStatus(ctx, userID, string(model.ApplicationStatusApproved))
	if err != nil {
		return nil, err
	}
	myRejected, err := s.applicationDao.CountByApplicantAndStatus(ctx, userID, string(model.ApplicationStatusRejected))
	if err != nil {
		return nil, err
	}

	stats["my_applications"] = map[string]int{
		"pending":  myPending,
		"approved": myApproved,
		"rejected": myRejected,
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
