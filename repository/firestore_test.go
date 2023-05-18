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

	suite.repository = NewFirestoreRepository(suite.fs, &TestNoSQL{})
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

func (suite *FirestoreRepositoryTestSuite) TestFind() {}

type TestNoSQL struct {
	ModelNoSQL
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func (t TestNoSQL) TableName() string {
	return "test"
}
