package app

type VMMeta struct {
	Image string `yml:"image"`
	OS    string `yml:"os"`
}

func (h *App) getImageMeta(image string) (*VMMeta, bool) {
	meta, ok := h.vms[image]
	return meta, ok
}
