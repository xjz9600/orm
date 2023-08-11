package round_robin

import (
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/balancer"
	"testing"
)

func TestWeightBalancer_Pick(t *testing.T) {
	b := &WeightBalancer{
		connections: []*weightConn{
			{
				c: SubConn{
					name: "weight-5",
				},
				weight:          5,
				efficientWeight: 5,
				currentWeight:   5,
			},
			{
				c: SubConn{
					name: "weight-4",
				},
				weight:          4,
				efficientWeight: 4,
				currentWeight:   4,
			},
			{
				c: SubConn{
					name: "weight-3",
				},
				weight:          3,
				efficientWeight: 3,
				currentWeight:   3,
			},
		},
	}
	res, err := b.Pick(balancer.PickInfo{})
	assert.NoError(t, err)
	assert.Equal(t, res.SubConn.(SubConn).name, "weight-5")

	res, err = b.Pick(balancer.PickInfo{})
	assert.NoError(t, err)
	assert.Equal(t, res.SubConn.(SubConn).name, "weight-4")

	res, err = b.Pick(balancer.PickInfo{})
	assert.NoError(t, err)
	assert.Equal(t, res.SubConn.(SubConn).name, "weight-3")

	res, err = b.Pick(balancer.PickInfo{})
	assert.NoError(t, err)
	assert.Equal(t, res.SubConn.(SubConn).name, "weight-5")

	res, err = b.Pick(balancer.PickInfo{})
	assert.NoError(t, err)
	assert.Equal(t, res.SubConn.(SubConn).name, "weight-4")
}
