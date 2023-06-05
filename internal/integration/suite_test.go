package integration

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"orm"
)

type Suite struct {
	suite.Suite
	driver string
	dsn    string
	db     *orm.DB
}

func (i *Suite) SetupSuite() {
	db, err := orm.Open(i.driver, i.dsn)
	assert.NoError(i.T(), err)
	err = db.Wait()
	assert.NoError(i.T(), err)
	i.db = db
}
