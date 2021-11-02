package types

const (
	// PVCDeleteDateAnnotationName delete mark annotation for pvc
	PVCDeleteDateAnnotationName = "delete-time.pvc.hakurei.cn"
	// PVCDeleteRetentionSeconds how many seconds of pvc retention
	PVCDeleteRetentionSeconds = 180 * 24 * 60 * 60 // 180 days
	ProxySQLWriterGroup       = 10
	ProxySQLReaderGroup       = 20
)
