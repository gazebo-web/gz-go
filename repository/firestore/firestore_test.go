package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go/v4"
	"fmt"
	"github.com/gazebo-web/gz-go/v7/errors"
	"github.com/gazebo-web/gz-go/v7/repository"
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

	suite.repository = NewFirestoreRepository[Test](suite.fs, "test")
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
	err := suite.repository.FirstOrCreate(nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, errors.ErrMethodNotImplemented)
}

func (suite *FirestoreRepositoryTestSuite) TestCreate() {
	_, err := suite.repository.Create(nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, errors.ErrMethodNotImplemented)
}

func (suite *FirestoreRepositoryTestSuite) TestCreateBulk() {
	_, err := suite.repository.CreateBulk(nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, errors.ErrMethodNotImplemented)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_All() {
	var found []Test

	suite.setupMockData()

	suite.Require().NoError(suite.repository.Find(&found))
	suite.Assert().Len(found, 3)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_MaxResults() {
	var found []Test

	suite.setupMockData()

	// Calling with max results should return the same amount of elements
	suite.Require().NoError(suite.repository.Find(&found, MaxResults(1)))
	suite.Assert().Len(found, 1)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_Offset() {
	var found []Test

	suite.setupMockData()

	// Calling with offset should return total - offset elements.
	suite.Require().NoError(suite.repository.Find(&found, Offset(1)))
	suite.Assert().Len(found, 2)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_OrderBy_Ascending() {
	var found []Test

	suite.setupMockData()

	suite.Require().NoError(suite.repository.Find(&found, OrderBy(
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

	suite.Require().NoError(suite.repository.Find(&found, OrderBy(
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

	suite.Require().NoError(suite.repository.Find(&found, Where("Value", "==", 1)))
	suite.Assert().Len(found, 1)
	suite.Assert().Equal(1, found[0].Value)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_Pagination_PageWithCursor() {
	suite.setupMockData()

	var found []Test
	suite.Require().NoError(suite.repository.Find(&found, OrderBy(Descending("Value")), StartAfter(3), MaxResults(100)))
	suite.Assert().Len(found, 2)
	suite.Assert().Equal(2, found[0].Value)
	suite.Assert().Equal(1, found[1].Value)
}

func (suite *FirestoreRepositoryTestSuite) TestFind_Pagination_PageWithSize() {
	suite.setupMockData()

	suite.Run("page=0/size=1/order=value", func() {
		var found []Test
		suite.Require().NoError(suite.repository.Find(&found, Offset(0), MaxResults(1), OrderBy(Descending("Value"))))
		suite.Assert().Len(found, 1)
		suite.Assert().Equal(3, found[0].Value)
	})

	suite.Run("page=1/size=1/order=value", func() {
		var found []Test
		suite.Require().NoError(suite.repository.Find(&found, Offset(1), MaxResults(1), OrderBy(Descending("Value"))))
		suite.Assert().Len(found, 1)
		suite.Assert().Equal(2, found[0].Value)
	})

	suite.Run("page=2/size=1/order=value", func() {
		var found []Test
		suite.Require().NoError(suite.repository.Find(&found, Offset(2), MaxResults(1), OrderBy(Descending("Value"))))
		suite.Assert().Len(found, 1)
		suite.Assert().Equal(1, found[0].Value)
	})

	suite.Run("page=3/size=1/order=value", func() {
		var found []Test
		suite.Require().NoError(suite.repository.Find(&found, Offset(3), MaxResults(1), OrderBy(Descending("Value"))))
		suite.Assert().Len(found, 0)
	})
}

func (suite *FirestoreRepositoryTestSuite) TestFindOne() {
	err := suite.repository.FindOne(nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, errors.ErrMethodNotImplemented)
}

func (suite *FirestoreRepositoryTestSuite) TestLast() {
	err := suite.repository.Last(nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, errors.ErrMethodNotImplemented)
}

func (suite *FirestoreRepositoryTestSuite) TestUpdate() {
	err := suite.repository.Update(nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, errors.ErrMethodNotImplemented)
}

func (suite *FirestoreRepositoryTestSuite) TestDelete() {
	err := suite.repository.Delete()
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, errors.ErrMethodNotImplemented)
}

func (suite *FirestoreRepositoryTestSuite) TestCount() {
	_, err := suite.repository.Count()
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, errors.ErrMethodNotImplemented)
}

func (suite *FirestoreRepositoryTestSuite) setupMockData() {
	// Clear any previously existing data
	suite.clearFirestoreData()

	_, _, err := suite.fs.Collection("test").Add(context.Background(), Test{
		Name:  "test-1",
		Value: 1,
	})
	suite.Require().NoError(err)

	_, _, err = suite.fs.Collection("test").Add(context.Background(), Test{
		Name:  "test-2",
		Value: 2,
	})
	suite.Require().NoError(err)

	_, _, err = suite.fs.Collection("test").Add(context.Background(), Test{
		Name:  "test-3",
		Value: 3,
	})
	suite.Require().NoError(err)
}

type Test struct {
	Model
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func (t Test) TableName() string {
	return "test"
}
