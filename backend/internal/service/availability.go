package service

import (
	"errors"
	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/model"
)

type AvailabilityService struct {
	availabilityDAO *dao.AvailabilityDAO
}

func NewAvailabilityService() *AvailabilityService {
	return &AvailabilityService{
		availabilityDAO: dao.NewAvailabilityDAO(),
	}
}

func (s *AvailabilityService) AddAvailability(userID int, req *model.AddAvailabilityRequest) error {
	var availabilities []model.Availability

	for _, week := range req.Weeks {
		availability := model.Availability{
			UserID:  userID,
			Week:    week,
			Weekday: req.Weekday,
			Period:  req.Period,
		}
		availabilities = append(availabilities, availability)
	}

	if len(availabilities) == 0 {
		return errors.New("无有效数据")
	}

	return s.availabilityDAO.CreateBatch(availabilities)
}

func (s *AvailabilityService) GetMyAvailability(userID int) ([]model.Availability, error) {
	return s.availabilityDAO.GetByUserID(userID)
}

func (s *AvailabilityService) GetAllAvailability() ([]model.AvailabilityWithUser, error) {
	return s.availabilityDAO.GetAllGrouped()
}

func (s *AvailabilityService) DeleteAvailability(userID int, id int) error {
	availability, err := s.availabilityDAO.GetByID(id)
	if err != nil {
		return errors.New("记录不存在")
	}

	if availability.UserID != userID {
		return errors.New("无权删除")
	}

	return s.availabilityDAO.Delete(id)
}

func (s *AvailabilityService) GetAvailableUsers(week, weekday, period int) ([]model.User, error) {
	return s.availabilityDAO.GetAvailableUsers(week, weekday, period)
}

// CreateBatch 批量创建无课时间记录
func (s *AvailabilityService) CreateBatch(userID int, availabilities []model.Availability) error {
	if len(availabilities) == 0 {
		return nil
	}
	return s.availabilityDAO.CreateBatch(availabilities)
}

// DeleteByUserID 删除用户的所有无课时间记录
func (s *AvailabilityService) DeleteByUserID(userID int) error {
	return s.availabilityDAO.DeleteByUserID(userID)
}
