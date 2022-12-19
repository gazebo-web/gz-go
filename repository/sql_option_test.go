package repository

import (
	"fmt"
	utilsgorm "github.com/gazebo-web/gz-go/v6/database/gorm"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SQLOptionsReferenceModel struct {
	ModelSQL
	Value       int `json:"value"`
	ReferenceID *uint
	Reference   *SQLOptionsReferenceModel
}

func (t SQLOptionsReferenceModel) TableName() string {
	return "sql_options_reference_model"
}

type SQLOptionsTestModel struct {
	ModelSQL
	Name         string `json:"name"`
	Value        int    `json:"value"`
	Even         bool   `json:"even"`
	LTE5         bool   `json:"lte_5"`
	Reference1ID *uint
	Reference1   *SQLOptionsReferenceModel
	Reference2ID *uint
	Reference2   *SQLOptionsReferenceModel
	Reference3ID *uint
	Reference3   *SQLOptionsReferenceModel
}

func (t SQLOptionsTestModel) GetID() uint {
	return t.ID
}

func (t SQLOptionsTestModel) TableName() string {
	return "sql_options_test_model"
}

func TestSQLOptions(t *testing.T) {
	suite.Run(t, new(SQLOptionsTestSuite))
}

type SQLOptionsTestSuite struct {
	suite.Suite
	db         *gorm.DB
	repository Repository
	models     []interface{}
}

func (s *SQLOptionsTestSuite) SetupSuite() {
	db, err := utilsgorm.GetTestDBFromEnvVars()
	s.Require().NoError(err)
	s.db = db.Model(new(SQLOptionsTestModel))

	s.models = []interface{}{new(SQLOptionsReferenceModel), new(SQLOptionsTestModel)}
	s.repository = NewRepository(s.db, &SQLOptionsTestModel{})
}

func (s *SQLOptionsTestSuite) SetupTest() {

	s.Require().NoError(s.db.DropTableIfExists(s.models...).Error)
	s.Require().NoError(s.db.AutoMigrate(s.models...).Error)
	// Create test entries
	for i := 0; i < 10; i++ {
		refref := SQLOptionsReferenceModel{
			Value: 100 * (i + 1),
		}
		s.db.Create(&refref)

		ref := SQLOptionsReferenceModel{
			Value:     i + 1,
			Reference: &refref,
		}
		s.db.Create(&ref)

		s.db.Create(&SQLOptionsTestModel{
			Name:       fmt.Sprintf("Test %d", i+1),
			Value:      i + 1,
			Even:       (i+1)%2 == 0,
			LTE5:       (i + 1) <= 5,
			Reference1: &ref,
			Reference2: &ref,
			Reference3: &ref,
		})
	}

	// Check that the set of results is as expected
	s.validateNoOptionFind()
}

func (s *SQLOptionsTestSuite) TearDownSuite() {
	// Options should not modify the set of results
	s.validateNoOptionFind()

	s.Require().NoError(s.db.DropTableIfExists(s.models).Error)
	s.Require().NoError(s.db.Close())
}

func (s *SQLOptionsTestSuite) TestSQLOptionImplementsOption() {
	s.Assert().Implements((*Option)(nil), new(SQLOption))
}

func (s *SQLOptionsTestSuite) getValues(values []*SQLOptionsTestModel) (out []int) {
	for _, v := range values {
		out = append(out, v.Value)
	}
	return out
}

func (s *SQLOptionsTestSuite) validateNoOptionFind() {
	var out []*SQLOptionsTestModel
	s.Require().NoError(s.repository.Find(&out))
	s.Assert().EqualValues([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, s.getValues(out))
}

func (s *SQLOptionsTestSuite) TestFindWhereOption() {
	var out []*SQLOptionsTestModel

	// Single condition
	s.Require().NoError(s.repository.Find(&out, Where("value % ? = 0", 2)))
	s.Assert().EqualValues([]int{2, 4, 6, 8, 10}, s.getValues(out))

	// Multiple conditions
	s.Require().NoError(s.repository.Find(&out, Where("value % ? = 0 AND (value < ? OR value > ?)", 2, 5, 7)))
	s.Assert().EqualValues([]int{2, 4, 8, 10}, s.getValues(out))
}

func (s *SQLOptionsTestSuite) TestFindMaxResultsOption() {
	var out []*SQLOptionsTestModel

	// Single option truncates number of results
	expected := 5
	s.Require().NoError(s.repository.Find(&out, MaxResults(expected)))
	s.Assert().Len(out, expected)

	// Multiple options passed uses the latest one
	expected = 5
	s.Require().NoError(s.repository.Find(&out, MaxResults(expected+2), MaxResults(expected)))
	s.Assert().Len(out, expected)
}

func (s *SQLOptionsTestSuite) TestFindOffsetOption() {
	var out []*SQLOptionsTestModel

	// Single offset
	s.Require().NoError(s.repository.Find(&out, MaxResults(10), Offset(5)))
	s.Assert().EqualValues([]int{6, 7, 8, 9, 10}, s.getValues(out))

	// Multiple offsets uses the latest one
	s.Require().NoError(s.repository.Find(&out, MaxResults(10), Offset(7), Offset(5)))
	s.Assert().EqualValues([]int{6, 7, 8, 9, 10}, s.getValues(out))

	// Using offset without max results does not change the offset
	s.Require().NoError(s.repository.Find(&out, Offset(5)))
	s.Assert().EqualValues([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, s.getValues(out))
}

func (s *SQLOptionsTestSuite) TestFindSelectAndGroupByOptions() {
	var out []*SQLOptionsTestModel

	// GroupBy fails if unaggregated fields are returned
	s.Assert().Error(s.repository.Find(&out, GroupBy("even")))

	// Single field group
	s.Assert().NoError(s.repository.Find(&out, Fields("SUM(value) value"), GroupBy("even")))
	s.Assert().EqualValues([]int{1 + 3 + 5 + 7 + 9, 2 + 4 + 6 + 8 + 10}, s.getValues(out))

	// Multiple field group
	s.Assert().NoError(s.repository.Find(&out, Fields("SUM(value) value"), GroupBy("even", "lte5")))
	s.Assert().EqualValues([]int{1 + 3 + 5, 2 + 4, 6 + 8 + 10, 7 + 9}, s.getValues(out))
}

func (s *SQLOptionsTestSuite) TestFindPreloadOption() {
	var out []*SQLOptionsTestModel

	// No related field should be present by default
	s.Assert().NoError(s.repository.Find(&out, Where("id = ?", 1)))
	s.Assert().Nil(out[0].Reference1)
	s.Assert().Nil(out[0].Reference2)
	s.Assert().Nil(out[0].Reference3)

	// Preloading a reference should bring the reference and its related fields should not be present
	s.Assert().NoError(s.repository.Find(&out, Where("id = ?", 1), Preload("Reference1")))
	s.Assert().NotNil(out[0].Reference1)
	s.Assert().Nil(out[0].Reference1.Reference)
	s.Assert().Nil(out[0].Reference2)
	s.Assert().Nil(out[0].Reference3)
	s.Assert().Equal(out[0].Reference1.Value, out[0].Value)

	// Multiple preloads should bring all references and its related fields should not be present
	s.Assert().NoError(s.repository.Find(&out, Where("id = ?", 1), Preload("Reference1"), Preload("Reference2")))
	s.Assert().NotNil(out[0].Reference1)
	s.Assert().Nil(out[0].Reference1.Reference)
	s.Assert().NotNil(out[0].Reference2)
	s.Assert().Nil(out[0].Reference2.Reference)
	s.Assert().Nil(out[0].Reference3)
	s.Assert().Equal(out[0].Reference1.Value, out[0].Value)
	s.Assert().Equal(out[0].Reference2.Value, out[0].Value)

	// Preloading a nested reference should bring the reference and its nested field should be present
	s.Assert().NoError(s.repository.Find(&out, Where("id = ?", 1), Preload("Reference1.Reference")))
	s.Assert().NotNil(out[0].Reference1)
	s.Assert().NotNil(out[0].Reference1.Reference)
	s.Assert().Equal(out[0].Reference1.Value, out[0].Value)
	s.Assert().Equal(out[0].Reference1.Reference.Value, 100*out[0].Reference1.Value)

	// Filtering preloads with matching conditions should make reference present
	s.Assert().NoError(s.repository.Find(&out, Where("id = ?", 1), Preload("Reference1", "value = ?", 1)))
	s.Assert().NotNil(out[0].Reference1)

	// Filtering preloads with non-matching conditions should prevent the reference from being present
	s.Assert().NoError(s.repository.Find(&out, Where("id = ?", 1), Preload("Reference1", "value = ?", 1000)))
	s.Assert().Nil(out[0].Reference1)
}

func (s *SQLOptionsTestSuite) TestFindOrderByOption() {
	var out []*SQLOptionsTestModel

	// Single descending order
	s.Require().NoError(s.repository.Find(&out, OrderBy(Descending("even"))))
	s.Assert().EqualValues([]int{2, 4, 6, 8, 10, 1, 3, 5, 7, 9}, s.getValues(out))

	// Single ascending order
	s.Require().NoError(s.repository.Find(&out, OrderBy(Ascending("even"))))
	s.Assert().EqualValues([]int{1, 3, 5, 7, 9, 2, 4, 6, 8, 10}, s.getValues(out))

	// // Multiple orders in a single call should stack
	s.Require().NoError(s.repository.Find(&out, OrderBy(Descending("even"), Descending("value"))))
	s.Assert().EqualValues([]int{10, 8, 6, 4, 2, 9, 7, 5, 3, 1}, s.getValues(out))

	// Multiple orders in multiple calls should be combined in the order the options were applied
	s.Require().NoError(s.repository.Find(&out, OrderBy(Descending("even")), OrderBy(Descending("value"))))
	s.Assert().EqualValues([]int{10, 8, 6, 4, 2, 9, 7, 5, 3, 1}, s.getValues(out))
}
