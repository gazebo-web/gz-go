package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go/v4"
	"fmt"
	"github.com/gazebo-web/gz-go/v8/errors"
	"github.com/gazebo-web/gz-go/v8/repository"
	"github.com/stretchr/testify/suite"
	"net/http"
	"os"
	"testing"
)

func TestFirestoreRepository(t *testing.T) {
	suite.Run(t, new(FirestoreRepositoryTestSuite))
}

type FirestoreRepositoryTestSuite struct {
	suite.Suite
	fs              *firestore.Client
	repository      repository.Repository
	client          *firebase.App
	emulatorAddress string
	projectID       string
}

func (suite *FirestoreRepositoryTestSuite) SetupSuite() {
	var err error

	suite.projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	suite.Require().NotEmpty(suite.projectID, "GOOGLE_CLOUD_PROJECT env var should be set")
	suite.emulatorAddress = os.Getenv("FIRESTORE_EMULATOR_HOST")
	suite.Require().NotEmpty(suite.emulatorAddress, "FIRESTORE_EMULATOR_HOST env var should be set")

	ctx := context.Background()
	suite.client, err = firebase.NewApp(ctx, nil)
	suite.Require().NoError(err)

	suite.fs, err = suite.client.Firestore(ctx)
	suite.Require().NoError(err)

	suite.repository = NewFirestoreRepository[Test](suite.fs)
}

func (suite *FirestoreRepositoryTestSuite) TearDownSuite() {
	suite.clearFirestoreData()
	suite.Require().NoError(suite.fs.Close())
}

func (suite *FirestoreRepositoryTestSuite) clearFirestoreData() {
	var client http.Client

	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf(
			"http://%s/emulator/v1/projects/%s/databases/(default)/documents",
			suite.emulatorAddress,
			suite.projectID,
		),
		nil,
	)
	suite.Require().NoError(err)

	_, err = client.Do(req)
	suite.Require().NoError(err)
}

func (suite *FirestoreRepositoryTestSuite) TestFirstOrCreate() {
	err := suite.repository.FirstOrCreate(context.Background(), nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, errors.ErrMethodNotImplemented)
}

func (suite *FirestoreRepositoryTestSuite) TestCreate() {
	_, err := suite.repository.Create(context.Background(), nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, errors.ErrMethodNotImplemented)
}

func (suite *FirestoreRepositoryTestSuite) TestCreateBulk() {
	_, err := suite.repository.CreateBulk(context.Background(), nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, errors.ErrMethodNotImplemented)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_All() {
	var found []Test

	suite.setupMockData()

	suite.Require().NoError(suite.repository.Find(context.Background(), &found))
	suite.Assert().Len(found, 3)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_MaxResults() {
	var found []Test

	suite.setupMockData()

	// Calling with max results should return the same amount of elements
	suite.Require().NoError(suite.repository.Find(context.Background(), &found, MaxResults(1)))
	suite.Assert().Len(found, 1)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_Offset() {
	var found []Test

	suite.setupMockData()

	// Calling with offset should return total - offset elements.
	suite.Require().NoError(suite.repository.Find(context.Background(), &found, Offset(1)))
	suite.Assert().Len(found, 2)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_OrderBy_Ascending() {
	var found []Test

	suite.setupMockData()

	suite.Require().NoError(suite.repository.Find(context.Background(), &found, OrderBy(
		Ascending("Value"),
	)))
	suite.Assert().Len(found, 3)
	suite.Assert().Equal(1, found[0].Value)
	suite.Assert().Equal(2, found[1].Value)
	suite.Assert().Equal(3, found[2].Value)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_OrderBy_Descending() {
	var found []Test

	suite.setupMockData()

	suite.Require().NoError(suite.repository.Find(context.Background(), &found, OrderBy(
		Descending("Value"),
	)))
	suite.Assert().Len(found, 3)
	suite.Assert().Equal(3, found[0].Value)
	suite.Assert().Equal(2, found[1].Value)
	suite.Assert().Equal(1, found[2].Value)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_Where() {
	var found []Test

	suite.setupMockData()

	suite.Require().NoError(suite.repository.Find(context.Background(), &found, Where("Value", "==", 1)))
	suite.Assert().Len(found, 1)
	suite.Assert().Equal(1, found[0].Value)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_Pagination_PageWithStartAfter() {
	suite.setupMockData()

	var found []Test
	suite.Require().NoError(suite.repository.Find(context.Background(), &found, OrderBy(Descending("Value")), StartAfter(3), MaxResults(100)))
	suite.Assert().Len(found, 2)
	suite.Assert().Equal(2, found[0].Value)
	suite.Assert().Equal(1, found[1].Value)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_Pagination_PageWithStartAt() {
	suite.setupMockData()

	var found []Test
	suite.Require().NoError(suite.repository.Find(context.Background(), &found, OrderBy(Descending("Value")), StartAt(2), MaxResults(100)))
	suite.Assert().Len(found, 2)
	suite.Assert().Equal(2, found[0].Value)
	suite.Assert().Equal(1, found[1].Value)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_Pagination_PageWithSize() {
	suite.setupMockData()

	suite.Run("page=0/size=1/order=value", func() {
		var found []Test
		suite.Require().NoError(suite.repository.Find(context.Background(), &found, Offset(0), MaxResults(1), OrderBy(Descending("Value"))))
		suite.Assert().Len(found, 1)
		suite.Assert().Equal(3, found[0].Value)
	})

	suite.Run("page=1/size=1/order=value", func() {
		var found []Test
		suite.Require().NoError(suite.repository.Find(context.Background(), &found, Offset(1), MaxResults(1), OrderBy(Descending("Value"))))
		suite.Assert().Len(found, 1)
		suite.Assert().Equal(2, found[0].Value)
	})

	suite.Run("page=2/size=1/order=value", func() {
		var found []Test
		suite.Require().NoError(suite.repository.Find(context.Background(), &found, Offset(2), MaxResults(1), OrderBy(Descending("Value"))))
		suite.Assert().Len(found, 1)
		suite.Assert().Equal(1, found[0].Value)
	})

	suite.Run("page=3/size=1/order=value", func() {
		var found []Test
		suite.Require().NoError(suite.repository.Find(context.Background(), &found, Offset(3), MaxResults(1), OrderBy(Descending("Value"))))
		suite.Assert().Len(found, 0)
	})
}

func (suite *FirestoreRepositoryTestSuite) TestFind_NoOp() {
	suite.setupMockData()

	// Calling with noop should return all elements
	var expected []Test
	suite.Require().NoError(suite.repository.Find(context.Background(), &expected))

	var found []Test
	suite.Require().NoError(suite.repository.Find(context.Background(), &found, NoOp()))

	suite.Assert().Len(found, len(expected))
	suite.Assert().Equal(found, expected)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_In() {
	var found []Test

	suite.setupMockData()

	// Calling with "In" should return the values that match the given names
	suite.Require().NoError(suite.repository.Find(context.Background(), &found, In[string]("Name", []string{"test-1", "test-2"})))
	suite.Assert().Len(found, 2)
	suite.Assert().Condition(func() (success bool) {
		for _, element := range found {
			success = element.Name == "test-1" || element.Name == "test-2"
		}
		return
	})
}

func (suite *FirestoreRepositoryTestSuite) TestFind_In_EmptyValues() {

	suite.setupMockData()

	var expected []Test
	suite.Require().NoError(suite.repository.Find(context.Background(), &expected, NoOp()))

	// Calling Find with an In option that has no values should be the same as NoOp.
	var found []Test
	suite.Require().NoError(suite.repository.Find(context.Background(), &found, In[string]("Name", nil)))
	suite.Assert().Len(found, len(expected))
	suite.Assert().Equal(found, expected)
}

func (suite *FirestoreRepositoryTestSuite) TestFindOne() {
	err := suite.repository.FindOne(context.Background(), nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, errors.ErrMethodNotImplemented)
}

func (suite *FirestoreRepositoryTestSuite) TestLast() {
	err := suite.repository.Last(context.Background(), nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, errors.ErrMethodNotImplemented)
}

func (suite *FirestoreRepositoryTestSuite) TestUpdate() {
	err := suite.repository.Update(context.Background(), nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, errors.ErrMethodNotImplemented)
}

func (suite *FirestoreRepositoryTestSuite) TestDelete() {
	suite.setupMockData()

	var before []Test
	suite.Require().NoError(suite.repository.Find(context.Background(), &before, Where("Value", "==", 1)))
	suite.Require().NotZero(len(before))

	err := suite.repository.Delete(context.Background(), Where("Value", "==", 1))
	suite.Assert().NoError(err)

	var after []Test
	suite.Require().NoError(suite.repository.Find(context.Background(), &after, Where("Value", "==", 1)))
	suite.Require().Len(after, len(before)-1)
}

func (suite *FirestoreRepositoryTestSuite) TestDeleteBatch() {
	suite.setupMockData()

	var before []Test
	suite.Require().NoError(suite.repository.Find(context.Background(), &before, Where("Value", "in", []int{1, 2, 3})))
	suite.Require().NotZero(len(before))

	repo := suite.repository.(*firestoreRepository[Test])
	col := suite.fs.Collection("test")
	col.Query = col.Where("Value", "in", []int{1, 2})
	suite.Assert().NoError(repo.deleteBatch(context.Background(), col, 1))

	var after []Test
	suite.Require().NoError(suite.repository.Find(context.Background(), &after, Where("Value", "in", []int{1, 2, 3})))
	suite.Require().Len(after, len(before)-2)
}

func (suite *FirestoreRepositoryTestSuite) TestCount() {
	_, err := suite.repository.Count(context.Background())
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, errors.ErrMethodNotImplemented)
}

func (suite *FirestoreRepositoryTestSuite) setupMockData() {
	// Clear any previously existing data
	suite.clearFirestoreData()

	writer := suite.fs.BulkWriter(context.Background())

	for i := 1; i <= 3; i++ {
		ref := suite.fs.Collection("test").NewDoc()
		_, err := writer.Create(ref, Test{
			Name:  fmt.Sprintf("test-%d", i),
			Value: i,
		})
		suite.Assert().NoError(err)
	}

	writer.End()
}

type Test struct {
	Model
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func (t Test) TableName() string {
	return "test"
}
