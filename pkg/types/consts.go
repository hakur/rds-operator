package types

const (
	// PVCDeleteDateAnnotationName delete mark annotation for pvc
	PVCDeleteDateAnnotationName = "delete-time.pvc.hakurei.cn"
	// PVCDeleteRetentionDays how many seconds of pvc retention
	PVCDeleteRetentionDays = 180 * 24 * 60 * 60 // 180 days
)
