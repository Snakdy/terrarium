package pipinstall

import cbev1 "github.com/Snakdy/container-build-engine/pkg/api/v1"

const Name = "pip-install"

type Statement struct {
	options cbev1.Options
}
