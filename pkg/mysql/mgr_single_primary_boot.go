package mysql

func NewMGRSinglePrimaryBoot() (t *MGRSinglePrimaryBoot) {
	t = new(MGRSinglePrimaryBoot)
	return t
}

type MGRSinglePrimaryBoot struct {
}
