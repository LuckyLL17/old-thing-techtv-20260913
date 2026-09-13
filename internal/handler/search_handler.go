package handler

import (
	"upcycle-hub/api/dto"
	"upcycle-hub/internal/service"
	apperr "upcycle-hub/pkg/errors"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchSvc  *service.SearchService
	recommendSvc *service.RecommendService
}

func NewSearchHandler(s *service.SearchService, r *service.RecommendService) *SearchHandler {
	return &SearchHandler{searchSvc: s, recommendSvc: r}
}

func (h *SearchHandler) Search(c *gin.Context) {
	var req dto.SearchReq
	if err := c.ShouldBindQuery(&req); err != nil {
		Fail(c, apperr.Wrap(apperr.CodeValidation, "参数错误", err))
		return
	}
	if req.Keyword == "" {
		OK(c, &service.SearchResult{Keyword: ""})
		return
	}
	if req.CategoryID > 0 || req.Difficulty != "" {
		list, total, err := h.searchSvc.Tutorials(req.Keyword, req.CategoryID, req.Difficulty, req.Page, req.Size)
		if err != nil {
			Fail(c, err)
			return
		}
		PageOK(c, list, total, req.Page, req.Size)
		return
	}
	res, err := h.searchSvc.All(req.Keyword, req.Page, req.Size)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, res)
}

func (h *SearchHandler) Home(c *gin.Context) {
	hd, err := h.recommendSvc.Home()
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, hd)
}

func (h *SearchHandler) Random(c *gin.Context) {
	n := 6
	if v := c.Query("n"); v != "" {
		if x, err := atoi(v); err == nil && x > 0 && x <= 20 {
			n = x
		}
	}
	list, err := h.recommendSvc.RandomInspiration(n)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, list)
}

func (h *SearchHandler) Top(c *gin.Context) {
	n := 10
	if v := c.Query("n"); v != "" {
		if x, err := atoi(v); err == nil && x > 0 && x <= 50 {
			n = x
		}
	}
	list, err := h.recommendSvc.TopTutorials(n)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, list)
}

func atoi(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, apperr.ErrBadRequest
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
