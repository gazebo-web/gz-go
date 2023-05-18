package repository

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go/v4"
	"fmt"
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
	repository      Repository
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

	suite.repository = NewFirestoreRepository[TestNoSQL](suite.fs)
}

func (suite *FirestoreRepositoryTestSuite) TearDownTest() {
	suite.tearDownFirebaseEmulatorDatabase()
}

func (suite *FirestoreRepositoryTestSuite) tearDownFirebaseEmulatorDatabase() {
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

func (suite *FirestoreRepositoryTestSuite) TestFind() {
	var all []TestNoSQL

	suite.setupMockData()

	suite.Require().NoError(suite.repository.Find(&all))
	suite.Assert().Len(all, 3)
}

func (suite *FirestoreRepositoryTestSuite) setupMockData() {
	_, _, err := suite.fs.Collection("test").Add(context.Background(), TestNoSQL{
		Name:  "test-1",
		Value: 1,
	})
	suite.Require().NoError(err)

	_, _, err = suite.fs.Collection("test").Add(context.Background(), TestNoSQL{
		Name:  "test-2",
		Value: 2,
	})
	suite.Require().NoError(err)

	_, _, err = suite.fs.Collection("test").Add(context.Background(), TestNoSQL{
		Name:  "test-3",
		Value: 3,
	})
	suite.Require().NoError(err)
}

type TestNoSQL struct {
	ModelNoSQL
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func (t TestNoSQL) TableName() string {
	return "test"
}
