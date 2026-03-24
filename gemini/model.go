package gemini

// Model constants for common Gemini models.
const (
	ModelAuto          = "auto"
	ModelPro           = "pro"
	ModelFlash         = "flash"
	ModelFlashLite     = "flash-lite"
	ModelGemini25Pro   = "gemini-2.5-pro"
	ModelGemini25Flash = "gemini-2.5-flash"
	ModelGemini3Pro    = "gemini-3-pro-preview"
	ModelGemini3Flash  = "gemini-3-flash-preview"
)

// ApprovalMode constants.
const (
	ApprovalDefault  = "default"
	ApprovalAutoEdit = "auto_edit"
	ApprovalYolo     = "yolo"
)

// Exit code constants.
const (
	ExitSuccess      = 0
	ExitError        = 1
	ExitInvalidInput = 42
	ExitTurnLimit    = 53
)
