package stages

import (
	"context"
	"strings"
)

type NodeModulesFilter struct{}

func NewNodeModulesFilter() *NodeModulesFilter {
	return &NodeModulesFilter{}
}

func (l NodeModulesFilter) Run(ctx context.Context, in <-chan FileRef) <-chan FileRef {
	out := make(chan FileRef)

	go func() {
		defer close(out)

		for ref := range in {
			if strings.Contains(ref.Path, "node_modules") {
				continue
			}

			select {
			case out <- ref:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}
