package handlers

import (
	"log/slog"
	"net/http"
	"payments-dashboard/internal/models"
	"payments-dashboard/internal/repository"
	"payments-dashboard/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo      *repository.Repository
	importSvc *services.ImportService
	log       *slog.Logger
}

func New(repo *repository.Repository, importSvc *services.ImportService, log *slog.Logger) *Handler {
	return &Handler{repo: repo, importSvc: importSvc, log: log}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.GET("/dashboard/summary", h.GetDashboardSummary)
		api.GET("/clients", h.GetClients)
		api.GET("/projects", h.GetProjects)
		api.GET("/payments", h.GetPayments)
		api.GET("/payments/:id", h.GetPayment)
		api.PUT("/payments/:id/act", h.UpdateAct)
		api.POST("/import/bank-statement", h.ImportBankStatement)
		api.POST("/import/bank-statement/raw", h.ImportBankStatementRaw)
	}
}

func (h *Handler) GetDashboardSummary(c *gin.Context) {
	summary, err := h.repo.GetDashboardSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *Handler) GetClients(c *gin.Context) {
	clients, err := h.repo.GetAllClients()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, clients)
}

func (h *Handler) GetProjects(c *gin.Context) {
	summaries, err := h.repo.GetProjectSummaries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summaries)
}

func (h *Handler) GetPayments(c *gin.Context) {
	var filter models.PaymentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payments, err := h.repo.GetPayments(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payments)
}

func (h *Handler) GetPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	payment, err := h.repo.GetPaymentByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if payment.Act != nil {
		payment.Act.Status = payment.Act.CalculateStatus(payment.PaymentDate)
	}
	c.JSON(http.StatusOK, payment)
}

type UpdateActRequest struct {
	IsSent         bool   `json:"is_sent"`
	IsSigned       bool   `json:"is_signed"`
	ManagerComment string `json:"manager_comment"`
}

func (h *Handler) UpdateAct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateActRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	act, err := h.repo.UpsertAct(uint(id), req.IsSent, req.IsSigned, req.ManagerComment)
	if err != nil {
		h.log.Error("failed to upsert act", "payment_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.log.Info("act updated",
		"payment_id", id, "is_sent", req.IsSent, "is_signed", req.IsSigned, "status", act.Status)
	c.JSON(http.StatusOK, act)
}

// ImportBankStatement accepts already-parsed bank operations and imports them
// idempotently. Re-posting the same batch will not create duplicates.
//
// @Summary Импорт банковской выписки
// @Description Принимает массив распарсенных операций и создаёт оплаты идемпотентно (с retry на блокировках).
// @Tags import
// @Accept json
// @Produce json
// @Param operations body handlers.ImportRequest true "Операции из выписки"
// @Success 200 {object} services.ImportResult
// @Router /import/bank-statement [post]
func (h *Handler) ImportBankStatement(c *gin.Context) {
	var req ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.importSvc.Import(req.Operations)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

type ImportRequest struct {
	Operations []services.BankOperation `json:"operations"`
}

// ImportBankStatementRaw accepts raw statement text, extracts operations via the
// configured AI extractor (LLM if enabled, otherwise regex), then imports them
// idempotently.
//
// @Summary Импорт выписки из сырого текста (AI/regex)
// @Description Принимает сырой текст выписки, извлекает операции через LLM (если задан AI_PROVIDER) или regex-фолбэк, затем идемпотентно импортирует.
// @Tags import
// @Accept json
// @Produce json
// @Param body body handlers.ImportRawRequest true "Сырой текст выписки"
// @Success 200 {object} services.RawImportResult
// @Router /import/bank-statement/raw [post]
func (h *Handler) ImportBankStatementRaw(c *gin.Context) {
	var req ImportRawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Text) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
		return
	}
	result, err := h.importSvc.ImportRawText(c.Request.Context(), req.Text)
	if err != nil {
		h.log.Error("raw import failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

type ImportRawRequest struct {
	Text string `json:"text"`
}
