package handlers

import (
	"cyber-range/internal/service"
	"cyber-range/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChallengeHandler struct {
	svc *service.ChallengeService
}

func NewChallengeHandler(svc *service.ChallengeService) *ChallengeHandler {
	return &ChallengeHandler{svc: svc}
}

// Standard API response format
type APIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// List returns all available challenges
func (h *ChallengeHandler) List(c *gin.Context) {
	challenges, err := h.svc.ListChallenges(c.Request.Context())
	if err != nil {
		logger.Error(c.Request.Context(), "Failed to list challenges", "error", err)
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "Failed to fetch challenges",
		})
		return
	}

	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "success",
		Data: challenges,
	})
}

// Start launches a challenge instance for the user
func (h *ChallengeHandler) Start(c *gin.Context) {
	challengeID := c.Param("id")

	// TODO: Extract userID from JWT token (for now use mock user)
	userID := "user_mock_001"

	instance, err := h.svc.StartInstance(c.Request.Context(), userID, challengeID)
	if err != nil {
		logger.Error(c.Request.Context(), "Failed to start instance",
			"user_id", userID, "challenge_id", challengeID, "error", err)

		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  err.Error(),
		})
		return
	}

	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "Instance started successfully",
		Data: instance,
	})
}

// Stop terminates a challenge instance
func (h *ChallengeHandler) Stop(c *gin.Context) {
	challengeID := c.Param("id")
	userID := "user_mock_001" // TODO: Extract from JWT

	if err := h.svc.StopInstance(c.Request.Context(), userID, challengeID); err != nil {
		logger.Error(c.Request.Context(), "Failed to stop instance", "error", err)
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  err.Error(),
		})
		return
	}

	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "Instance stopped successfully",
		Data: map[string]string{"status": "stopped"},
	})
}

// Verify validates a flag submission
func (h *ChallengeHandler) Verify(c *gin.Context) {
	var req struct {
		ChallengeID string `json:"challenge_id" binding:"required"`
		Flag        string `json:"flag" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.PureJSON(http.StatusBadRequest, APIResponse{
			Code: 400,
			Msg:  "Invalid request format",
		})
		return
	}

	userID := "user_mock_001" // TODO: Extract from JWT

	correct, message, err := h.svc.VerifyFlag(c.Request.Context(), userID, req.ChallengeID, req.Flag)
	if err != nil {
		logger.Error(c.Request.Context(), "Failed to verify flag", "error", err)
		c.PureJSON(http.StatusInternalServerError, APIResponse{
			Code: 500,
			Msg:  "Verification failed",
		})
		return
	}

	c.PureJSON(http.StatusOK, APIResponse{
		Code: 200,
		Msg:  "success",
		Data: map[string]interface{}{
			"correct": correct,
			"message": message,
		},
	})
}
