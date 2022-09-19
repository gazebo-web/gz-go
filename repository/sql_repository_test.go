package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
	utilsgorm "gitlab.com/ignitionrobotics/web/ign-go/v6/database/gorm"
	"testing"
)

func TestRepository(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

type RepositoryTestSuite struct {
	suite.Suite
	db         *gorm.DB
	Repository Repository
}

type Test struct {
	ModelSQL
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func (t Test) GetID() uint {
	return t.ID
}

func (t Test) TableName() string {
	return "test"
}

func (suite *RepositoryTestSuite) SetupSuite() {
	db, err := utilsgorm.GetTestDBFromEnvVars()
	suite.Require().NoError(err)
	suite.db = db

	suite.Repository = NewRepository(suite.db, &Test{})
}

func (suite *RepositoryTestSuite) SetupTest() {
	suite.Require().NoError(suite.db.DropTableIfExists(&Test{}).Error)
	suite.Require().NoError(suite.db.AutoMigrate(&Test{}).Error)

	test1 := &Test{
		Name:  "Test1",
		Value: 1,
	}
	test2 := &Test{
		Name:  "Test2",
		Value: 2,
	}

	test3 := &Test{
		Name:  "Test3",
		Value: 3,
	}

	res, err := suite.Repository.CreateBulk([]Model{test1, test2, test3})
	suite.Require().NoError(err)
	suite.Require().Len(res, 3)
}

func (suite *RepositoryTestSuite) TearDownSuite() {
	suite.Require().NoError(suite.db.DropTableIfExists(&Test{}).Error)
	suite.Require().NoError(suite.db.Close())
}

func (suite *RepositoryTestSuite) TestImplementsInterface() {
	var expected *Repository
	suite.Assert().Implements(expected, new(repositorySQL))
}

func (suite *RepositoryTestSuite) TestCreateOne() {
	// Creating one record should not fail.
	res, err := suite.Repository.CreateBulk([]Model{&Test{
		Name:  "test",
		Value: 999,
	}})
	suite.Assert().NoError(err)
	suite.Assert().Len(res, 1)

	var count int64
	err = suite.db.Model(&Test{}).Count(&count).Error
	suite.Require().NoError(err)
	suite.Assert().Equal(int64(4), count)
}

func (suite *RepositoryTestSuite) TestCreateMultiple() {
	// Creating multiple records should not fail
	res, err := suite.Repository.CreateBulk([]Model{
		&Test{
			Name:  "test",
			Value: 999,
		},
		&Test{
			Name:  "test",
			Value: 999,
		},
		&Test{
			Name:  "test",
			Value: 999,
		},
	})
	suite.Assert().NoError(err)
	suite.Assert().Len(res, 3)

	// And those records should be in the database.
	var count int64
	err = suite.db.Model(&Test{}).Count(&count).Error
	suite.Require().NoError(err)
	suite.Assert().Equal(int64(6), count)
}

func (suite *RepositoryTestSuite) TestFind() {
	var t []Test

	// Finding multiple records should not fail.
	err := suite.Repository.Find(&t, Where("name IN (?)", []string{"Test1", "Test2"}))
	suite.Require().NoError(err)

	suite.Assert().Len(t, 2)
}

func (suite *RepositoryTestSuite) TestFindOne() {
	var t Test

	// Finding one should not fail.
	suite.Assert().NoError(suite.Repository.FindOne(&t, Filter{
		Template: "name = ?",
		Values:   []interface{}{"Test1"},
	}, Filter{
		Template: "value = ?",
		Values:   []interface{}{1},
	}))

	suite.Assert().Equal("Test1", t.Name)
	suite.Assert().Equal(1, t.Value)
}

func (suite *RepositoryTestSuite) TestUpdate() {
	filter := Filter{
		Template: "name = ?",
		Values:   []interface{}{"Test1"},
	}

	// Update record using filters should not fail.
	suite.Assert().NoError(suite.Repository.Update(map[string]interface{}{"name": "Test111", "value": 12345}, filter))

	var t Test

	// Finding the old record should fail.
	suite.Assert().Error(suite.Repository.FindOne(&t, filter))

	// Finding the correct record should not fail.
	suite.Assert().NoError(suite.Repository.FindOne(&t, Filter{
		Template: "name = ?",
		Values:   []interface{}{"Test111"},
	}))

	// The updated values should be in the record
	suite.Assert().Equal("Test111", t.Name)
	suite.Assert().Equal(12345, t.Value)
}

func (suite *RepositoryTestSuite) TestDelete() {
	filter := Filter{
		Template: "name = ?",
		Values:   []interface{}{"Test1"},
	}

	// Deleting should not fail.
	suite.Assert().NoError(suite.Repository.Delete(filter))

	// Finding one should fail, the record no longer exists.
	var t Test
	suite.Assert().Error(suite.Repository.FindOne(&t, filter))

	// Deleting is idempotent, deleting twice should not return an error.
	suite.Assert().NoError(suite.Repository.Delete(filter))
}

func (suite *RepositoryTestSuite) TestFirstOrCreate() {
	var test Test
	suite.Require().NoError(suite.Repository.FirstOrCreate(&test, Filter{
		Template: "value = ?",
		Values:   []interface{}{1},
	}))

	suite.Assert().Equal(uint(1), test.ID)
	suite.Assert().Equal("Test1", test.Name)

	test = Test{
		Name:  "Test4",
		Value: 4,
	}

	suite.Require().NoError(suite.Repository.FirstOrCreate(&test, Filter{
		Template: "value = ?",
		Values:   []interface{}{4},
	}))

	suite.Assert().Equal(uint(4), test.ID)
	suite.Assert().Equal("Test4", test.Name)
}

func (suite *RepositoryTestSuite) TestLast() {
	// Create an extra entity
	test3 := Test{
		Name:  "Test3",
		Value: 3,
	}
	created, err := suite.Repository.Create(&test3)
	suite.Require().NoError(err)

	var out Test
	err = suite.Repository.Last(&out, Filter{
		Template: "name = ?",
		Values:   []interface{}{"Test3"},
	})
	suite.Assert().NoError(err)

	suite.Assert().Equal(created.GetID(), out.GetID())
}

func (suite *RepositoryTestSuite) TestCount() {
	count, err := suite.Repository.Count(Filter{
		Template: "name = ?",
		Values:   []interface{}{"Test3"},
	})
	suite.Assert().NoError(err)
	suite.Assert().Equal(uint64(1), count)

	count, err = suite.Repository.Count(Filter{
		Template: "name LIKE ?",
		Values:   []interface{}{"Test%"},
	})
	suite.Assert().NoError(err)
	suite.Assert().Equal(uint64(3), count)
}
