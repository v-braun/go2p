package go2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMin(t *testing.T) {
	actions := createMiddlewares(5)

	min := actions.Min()
	assert.Equal(t, 1, min)

	actions = actions.nextItems(Receive, 0)
	min = actions.Min()
	assert.Equal(t, 1, min)
}

func TestMax(t *testing.T) {
	actions := createMiddlewares(5)

	max := actions.Max()
	assert.Equal(t, 5, max)

	actions = actions.nextItems(Receive, 0)
	max = actions.Max()
	assert.Equal(t, 5, max)
}

func TestByPos(t *testing.T) {
	actions := createMiddlewares(5)

	p := actions.ByPos(3)

	assert.Equal(t, 3, p.Pos())
	assert.Equal(t, "m2", p.name) // generated id's are not same as pos

	assert.Equal(t, "m2 (3)", p.String()) // generated id's are not same as pos

}
