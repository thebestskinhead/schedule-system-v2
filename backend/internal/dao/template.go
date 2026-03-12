package dao

import (
	"encoding/json"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
)

type TemplateDAO struct{}

func NewTemplateDAO() *TemplateDAO {
	return &TemplateDAO{}
}

// GetAllTemplates 获取所有模板
func (d *TemplateDAO) GetAllTemplates(adminID int) ([]model.ExportTemplate, error) {
	query := `SELECT id, admin_id, name, description, config, is_default, created_at, updated_at 
			  FROM export_templates 
			  WHERE admin_id = ? OR is_default = 1
			  ORDER BY is_default DESC, id DESC`

	rows, err := db.GetDB().Query(query, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []model.ExportTemplate
	for rows.Next() {
		var t model.ExportTemplate
		err := rows.Scan(&t.ID, &t.AdminID, &t.Name, &t.Description, &t.Config, &t.IsDefault, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			continue
		}
		templates = append(templates, t)
	}
	return templates, nil
}

// GetTemplateByID 根据ID获取模板
func (d *TemplateDAO) GetTemplateByID(id int) (*model.ExportTemplate, error) {
	query := `SELECT id, admin_id, name, description, config, is_default, created_at, updated_at 
			  FROM export_templates WHERE id = ?`

	var t model.ExportTemplate
	err := db.GetDB().QueryRow(query, id).Scan(
		&t.ID, &t.AdminID, &t.Name, &t.Description, &t.Config,
		&t.IsDefault, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetDefaultTemplate 获取默认模板
func (d *TemplateDAO) GetDefaultTemplate() (*model.ExportTemplate, error) {
	query := `SELECT id, admin_id, name, description, config, is_default, created_at, updated_at 
			  FROM export_templates WHERE is_default = 1 LIMIT 1`

	var t model.ExportTemplate
	err := db.GetDB().QueryRow(query).Scan(
		&t.ID, &t.AdminID, &t.Name, &t.Description, &t.Config,
		&t.IsDefault, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateTemplate 创建模板
func (d *TemplateDAO) CreateTemplate(template *model.ExportTemplate) error {
	query := `INSERT INTO export_templates (admin_id, name, description, config, is_default) 
			  VALUES (?, ?, ?, ?, ?)`

	result, err := db.GetDB().Exec(query, template.AdminID, template.Name, template.Description, template.Config, template.IsDefault)
	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	template.ID = int(id)
	return nil
}

// UpdateTemplate 更新模板
func (d *TemplateDAO) UpdateTemplate(template *model.ExportTemplate) error {
	query := `UPDATE export_templates SET name = ?, description = ?, config = ?, is_default = ? WHERE id = ?`
	_, err := db.GetDB().Exec(query, template.Name, template.Description, template.Config, template.IsDefault, template.ID)
	return err
}

// DeleteTemplate 删除模板
func (d *TemplateDAO) DeleteTemplate(id int) error {
	query := `DELETE FROM export_templates WHERE id = ? AND is_default = 0`
	_, err := db.GetDB().Exec(query, id)
	return err
}

// ClearDefaultTemplates 清除默认模板标记
func (d *TemplateDAO) ClearDefaultTemplates() error {
	_, err := db.GetDB().Exec(`UPDATE export_templates SET is_default = 0 WHERE is_default = 1`)
	return err
}

// ParseTemplateConfig 解析模板配置
func ParseTemplateConfig(t *model.ExportTemplate) (*model.ExportTemplateConfig, error) {
	var config model.ExportTemplateConfig
	if err := json.Unmarshal(t.Config, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
