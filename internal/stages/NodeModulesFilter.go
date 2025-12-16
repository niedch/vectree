package stages

import (
	"context"
	"log"
	"strings"
)

type NodeModulesFilter struct{}

func NewNodeModulesFilter() *NodeModulesFilter {
	return &NodeModulesFilter{}
}

func (l NodeModulesFilter) Run(ctx context.Context, in <-chan string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)

		for filePath := range in {
			if strings.Contains(filePath, "node_modules") {
				log.Println("Ignoring File", filePath)
				continue
			}

			select {
			case out <- filePath:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}
