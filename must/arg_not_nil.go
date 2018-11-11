package must

import (
	"fmt"

	"github.com/pkg/errors"
)

// ArgNotNil panics if provided item is nil
func ArgNotNil(item interface{}, name string) {

	if item == nil {
		panic(errors.New(fmt.Sprintf("argument %s is nil", name)))
	}
}
