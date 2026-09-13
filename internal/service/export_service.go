package service

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
	"upcycle-hub/internal/domain"
	"upcycle-hub/internal/repository"
	apperr "upcycle-hub/pkg/errors"
	"upcycle-hub/pkg/utils"
)

// ExportService 负责把教程渲染成一份自包含的 A4 打印版 HTML 快照。
// 快照在触发那一刻生成并落盘，之后教程修改、图片替换都不影响已生成的文件。
type ExportService struct {
	tutorialRepo *repository.TutorialRepo
	exportRepo   *repository.ExportRepo
	uploadDir    string
}

func NewExportService(tr *repository.TutorialRepo, er *repository.ExportRepo, uploadDir string) *ExportService {
	return &ExportService{tutorialRepo: tr, exportRepo: er, uploadDir: uploadDir}
}

// Export 生成教程的 A4 打印快照，返回记录（含可访问的 FileURL）。
func (s *ExportService) Export(tutorialID, userID uint64) (*domain.TutorialExport, error) {
	t, err := s.tutorialRepo.GetByID(tutorialID, true)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	html, err := s.renderHTML(t, now)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(s.uploadDir, "exports")
	if err := utils.EnsureDir(dir); err != nil {
		return nil, apperr.Wrap(apperr.CodeFile, "创建导出目录失败", err)
	}
	fname := fmt.Sprintf("tutorial_%d_v%d_%d.html", t.ID, t.Version, now.UnixNano())
	full := filepath.Join(dir, fname)
	if err := os.WriteFile(full, []byte(html), 0644); err != nil {
		return nil, apperr.Wrap(apperr.CodeFile, "写入导出文件失败", err)
	}
	rec := &domain.TutorialExport{
		TutorialID: t.ID,
		Version:    t.Version,
		Title:      t.Title,
		FileName:   fname,
		FileURL:    "/uploads/exports/" + fname,
		FileSize:   int64(len(html)),
		CreatedBy:  userID,
	}
	if err := s.exportRepo.Create(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

func (s *ExportService) List(tutorialID uint64) ([]*domain.TutorialExport, error) {
	if _, err := s.tutorialRepo.GetByID(tutorialID, false); err != nil {
		return nil, err
	}
	return s.exportRepo.ListByTutorial(tutorialID, 20)
}

// ---------- 渲染 ----------

type exportItem struct {
	Name     string
	Quantity string
	Unit     string
	Notes    string
}

type exportStep struct {
	Index            int
	Title            string
	Content          string
	Image            template.URL
	Reminder         string
	EstimatedMinutes int
}

type exportData struct {
	Title       string
	Summary     string
	Difficulty  string
	Hours       string
	Category    string
	Author      string
	Tags        []string
	CoverBefore template.URL
	CoverAfter  template.URL
	Materials   []exportItem
	Tools       []exportItem
	Steps       []exportStep
	Version     int
	ExportedAt  string
}

func (s *ExportService) renderHTML(t *domain.Tutorial, now time.Time) (string, error) {
	d := &exportData{
		Title:      t.Title,
		Summary:    t.Summary,
		Difficulty: difficultyLabel(t.Difficulty),
		Hours:      formatHours(t.EstimatedHours),
		Version:    t.Version,
		ExportedAt: now.Format("2006-01-02 15:04"),
	}
	if t.Category != nil {
		d.Category = t.Category.Name
	}
	if t.User != nil {
		d.Author = t.User.Nickname
		if d.Author == "" {
			d.Author = t.User.Username
		}
	}
	for _, tag := range t.Tags {
		d.Tags = append(d.Tags, tag.Name)
	}
	d.CoverBefore = s.inlineImage(t.CoverBefore)
	d.CoverAfter = s.inlineImage(t.CoverAfter)
	for _, m := range t.Materials {
		item := exportItem{Name: m.Name, Quantity: m.Quantity, Unit: m.Unit, Notes: m.Notes}
		if m.IsTool {
			d.Tools = append(d.Tools, item)
		} else {
			d.Materials = append(d.Materials, item)
		}
	}
	for i, st := range t.Steps {
		d.Steps = append(d.Steps, exportStep{
			Index:            i + 1,
			Title:            st.Title,
			Content:          st.Content,
			Image:            s.inlineImage(st.Image),
			Reminder:         st.Reminder,
			EstimatedMinutes: st.EstimatedMinutes,
		})
	}
	tpl, err := template.New("export").Funcs(template.FuncMap{
		"inc": func(i int) int { return i + 1 },
	}).Parse(exportTemplate)
	if err != nil {
		return "", apperr.Wrap(apperr.CodeInternal, "解析导出模板失败", err)
	}
	var sb strings.Builder
	if err := tpl.Execute(&sb, d); err != nil {
		return "", apperr.Wrap(apperr.CodeInternal, "渲染导出页面失败", err)
	}
	return sb.String(), nil
}

// inlineImage 把本站 /uploads/ 下的图片读出来转成 base64 data URI 内联进 HTML，
// 保证快照完全自包含；外链图片保留原 URL。返回值经 safeImageURL 白名单校验，
// 只允许 data:image、本站相对路径和 http(s) 地址，其余一律丢弃。
func (s *ExportService) inlineImage(ref string) template.URL {
	if ref == "" {
		return ""
	}
	if !strings.HasPrefix(ref, "/uploads/") {
		return safeImageURL(ref)
	}
	rel := strings.TrimPrefix(ref, "/uploads/")
	// Clean("/"+rel) 规整掉 ../，确保最终路径一定落在 uploadDir 内
	full := filepath.Join(s.uploadDir, filepath.Clean("/"+rel))
	data, err := os.ReadFile(full)
	if err != nil {
		return safeImageURL(ref)
	}
	mime := "image/jpeg"
	switch strings.ToLower(filepath.Ext(full)) {
	case ".png":
		mime = "image/png"
	case ".gif":
		mime = "image/gif"
	case ".webp":
		mime = "image/webp"
	}
	return template.URL("data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data))
}

func safeImageURL(ref string) template.URL {
	if strings.HasPrefix(ref, "/") && !strings.HasPrefix(ref, "//") {
		return template.URL(ref)
	}
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return template.URL(ref)
	}
	return ""
}

func difficultyLabel(d string) string {
	switch d {
	case domain.DifficultyEasy:
		return "简单"
	case domain.DifficultyHard:
		return "困难"
	default:
		return "中等"
	}
}

func formatHours(h float64) string {
	if h == float64(int64(h)) {
		return fmt.Sprintf("%d", int64(h))
	}
	return fmt.Sprintf("%.1f", h)
}
