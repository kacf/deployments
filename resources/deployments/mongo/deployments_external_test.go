// Copyright 2016 Mender Software AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package mongo_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/mendersoftware/go-lib-micro/identity"
	ctxstore "github.com/mendersoftware/go-lib-micro/store"
	"github.com/stretchr/testify/assert"

	"github.com/mendersoftware/deployments/resources/deployments"
	. "github.com/mendersoftware/deployments/resources/deployments/mongo"
	. "github.com/mendersoftware/deployments/utils/pointers"
)

func TestDeploymentStorageInsert(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestDeploymentStorageInsert in short mode.")
	}

	testCases := []struct {
		InputDeployment *deployments.Deployment
		InputTenant     string
		OutputError     error
	}{
		{
			InputDeployment: nil,
			OutputError:     ErrDeploymentStorageInvalidDeployment,
		},
		{
			InputDeployment: deployments.NewDeployment(),
			OutputError:     errors.New("DeploymentConstructor: non zero value required;"),
		},
		{
			InputDeployment: deployments.NewDeploymentFromConstructor(&deployments.DeploymentConstructor{
				Name:         StringToPointer("NYC Production"),
				ArtifactName: StringToPointer("App 123"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			}),
			OutputError: nil,
		},
		{
			InputDeployment: deployments.NewDeploymentFromConstructor(&deployments.DeploymentConstructor{
				Name:         StringToPointer("NYC Production"),
				ArtifactName: StringToPointer("App 123"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			}),
			InputTenant: "acme",
		},
	}

	for testCaseNumber, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %d", testCaseNumber+1), func(t *testing.T) {

			// Make sure we start test with empty database
			db.Wipe()

			session := db.Session()
			store := NewDeploymentsStorage(session)

			ctx := context.Background()
			if testCase.InputTenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: testCase.InputTenant,
				})
			}

			err := store.Insert(ctx, testCase.InputDeployment)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)

				dep := session.DB(ctxstore.DbFromContext(ctx, DatabaseName)).
					C(CollectionDeployments)
				count, err := dep.Find(nil).Count()
				assert.NoError(t, err)
				assert.Equal(t, 1, count)

				if testCase.InputTenant != "" {
					indefault, _ := session.DB(DatabaseName).
						C(CollectionDeployments).
						Find(nil).Count()
					assert.Equal(t, 0, indefault)
				}
			}

			// Need to close all sessions to be able to call wipe at next test case
			session.Close()
		})
	}
}

func TestDeploymentStorageDelete(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestDeploymentStorageDelete in short mode.")
	}

	testCases := []struct {
		InputID                    string
		InputDeploymentsCollection []interface{}
		InputTenant                string

		OutputError error
	}{
		{
			InputID:     "",
			OutputError: ErrStorageInvalidID,
		},
		{
			InputID:     "b532b01a-9313-404f-8d19-e7fcbe5cc347",
			OutputError: nil,
		},
		{
			InputID: "b532b01a-9313-404f-8d19-e7fcbe5cc347",
			InputDeploymentsCollection: []interface{}{
				deployments.Deployment{
					DeploymentConstructor: &deployments.DeploymentConstructor{
						Name:         StringToPointer("NYC Production"),
						ArtifactName: StringToPointer("App 123"),
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: StringToPointer("b532b01a-9313-404f-8d19-e7fcbe5cc347"),
				},
			},
			OutputError: nil,
		},
		{
			InputID: "b532b01a-9313-404f-8d19-e7fcbe5cc347",
			InputDeploymentsCollection: []interface{}{
				deployments.Deployment{
					DeploymentConstructor: &deployments.DeploymentConstructor{
						Name:         StringToPointer("NYC Production"),
						ArtifactName: StringToPointer("App 123"),
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: StringToPointer("b532b01a-9313-404f-8d19-e7fcbe5cc347"),
				},
			},
			InputTenant: "acme",
		},
	}

	for testCaseNumber, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %d", testCaseNumber+1), func(t *testing.T) {

			// Make sure we start test with empty database
			db.Wipe()

			session := db.Session()
			store := NewDeploymentsStorage(session)

			ctx := context.Background()
			if testCase.InputTenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: testCase.InputTenant,
				})
			}

			dep := session.DB(ctxstore.DbFromContext(ctx, DatabaseName)).
				C(CollectionDeployments)
			if testCase.InputDeploymentsCollection != nil {
				assert.NoError(t, dep.Insert(testCase.InputDeploymentsCollection...))
			}

			err := store.Delete(ctx, testCase.InputID)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)

				count, err := dep.FindId(testCase.InputID).Count()
				assert.NoError(t, err)
				assert.Equal(t, 0, count)

				if testCase.InputTenant != "" {
					indefault, _ := session.DB(DatabaseName).
						C(CollectionDeployments).
						Find(nil).Count()
					assert.Equal(t, 0, indefault)
				}
			}

			// Need to close all sessions to be able to call wipe at next test case
			session.Close()
		})
	}
}

func TestDeploymentStorageFindByID(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestDeploymentStorageFindByID in short mode.")
	}

	testCases := []struct {
		InputID                    string
		InputDeploymentsCollection []interface{}
		InputTenant                string

		OutputError      error
		OutputDeployment *deployments.Deployment
	}{
		{
			InputID:     "",
			OutputError: ErrStorageInvalidID,
		},
		{
			InputID:          "b532b01a-9313-404f-8d19-e7fcbe5cc347",
			OutputError:      nil,
			OutputDeployment: nil,
		},
		{
			InputID: "b532b01a-9313-404f-8d19-e7fcbe5cc347",
			InputDeploymentsCollection: []interface{}{
				&deployments.Deployment{
					DeploymentConstructor: &deployments.DeploymentConstructor{
						Name:         StringToPointer("NYC Production"),
						ArtifactName: StringToPointer("App 123"),
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
				},
				&deployments.Deployment{
					DeploymentConstructor: &deployments.DeploymentConstructor{
						Name:         StringToPointer("NYC Production"),
						ArtifactName: StringToPointer("App 123"),
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: StringToPointer("d1804903-5caa-4a73-a3ae-0efcc3205405"),
				},
			},
			OutputError:      nil,
			OutputDeployment: nil,
		},
		{
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeploymentsCollection: []interface{}{
				&deployments.Deployment{
					DeploymentConstructor: &deployments.DeploymentConstructor{
						Name:         StringToPointer("NYC Production"),
						ArtifactName: StringToPointer("App 123"),
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
					Stats: map[string]int{
						deployments.DeviceDeploymentStatusDownloading: 0,
						deployments.DeviceDeploymentStatusInstalling:  0,
						deployments.DeviceDeploymentStatusRebooting:   0,
						deployments.DeviceDeploymentStatusPending:     10,
						deployments.DeviceDeploymentStatusSuccess:     15,
						deployments.DeviceDeploymentStatusFailure:     1,
						deployments.DeviceDeploymentStatusNoArtifact:  0,
						deployments.DeviceDeploymentStatusAlreadyInst: 0,
						deployments.DeviceDeploymentStatusAborted:     0,
					},
				},
				&deployments.Deployment{
					DeploymentConstructor: &deployments.DeploymentConstructor{
						Name:         StringToPointer("NYC Production"),
						ArtifactName: StringToPointer("App 123"),
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: StringToPointer("d1804903-5caa-4a73-a3ae-0efcc3205405"),
					Stats: map[string]int{
						deployments.DeviceDeploymentStatusDownloading: 0,
						deployments.DeviceDeploymentStatusInstalling:  0,
						deployments.DeviceDeploymentStatusRebooting:   0,
						deployments.DeviceDeploymentStatusPending:     5,
						deployments.DeviceDeploymentStatusSuccess:     10,
						deployments.DeviceDeploymentStatusFailure:     3,
						deployments.DeviceDeploymentStatusNoArtifact:  0,
						deployments.DeviceDeploymentStatusAlreadyInst: 0,
						deployments.DeviceDeploymentStatusAborted:     0,
					},
				},
			},
			OutputError: nil,
			OutputDeployment: &deployments.Deployment{
				DeploymentConstructor: &deployments.DeploymentConstructor{
					Name:         StringToPointer("NYC Production"),
					ArtifactName: StringToPointer("App 123"),
					//Devices is not kept around!
				},
				Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
				Stats: map[string]int{
					deployments.DeviceDeploymentStatusDownloading: 0,
					deployments.DeviceDeploymentStatusInstalling:  0,
					deployments.DeviceDeploymentStatusRebooting:   0,
					deployments.DeviceDeploymentStatusPending:     10,
					deployments.DeviceDeploymentStatusSuccess:     15,
					deployments.DeviceDeploymentStatusFailure:     1,
					deployments.DeviceDeploymentStatusNoArtifact:  0,
					deployments.DeviceDeploymentStatusAlreadyInst: 0,
					deployments.DeviceDeploymentStatusAborted:     0,
				},
			},
		},
		{
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeploymentsCollection: []interface{}{
				&deployments.Deployment{
					DeploymentConstructor: &deployments.DeploymentConstructor{
						Name:         StringToPointer("NYC Production"),
						ArtifactName: StringToPointer("App 123"),
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id:    StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
					Stats: map[string]int{},
				},
			},
			InputTenant: "acme",

			OutputDeployment: &deployments.Deployment{
				DeploymentConstructor: &deployments.DeploymentConstructor{
					Name:         StringToPointer("NYC Production"),
					ArtifactName: StringToPointer("App 123"),
					//Devices is not kept around!
				},
				Id:    StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
				Stats: map[string]int{},
			},
		},
	}

	for testCaseNumber, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %d", testCaseNumber+1), func(t *testing.T) {

			// Make sure we start test with empty database
			db.Wipe()

			session := db.Session()
			store := NewDeploymentsStorage(session)

			ctx := context.Background()
			if testCase.InputTenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: testCase.InputTenant,
				})
			}

			dep := session.DB(ctxstore.DbFromContext(ctx, DatabaseName)).
				C(CollectionDeployments)
			if testCase.InputDeploymentsCollection != nil {
				assert.NoError(t, dep.Insert(testCase.InputDeploymentsCollection...))
			}

			deployment, err := store.FindByID(ctx, testCase.InputID)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.OutputDeployment, deployment)
			}

			// tenant is set, verify that deployment is not present in default DB
			if testCase.InputTenant != "" {
				deployment, err := store.FindByID(context.Background(),
					testCase.InputID)
				assert.Nil(t, deployment)
				assert.Nil(t, err)
			}

			// Need to close all sessions to be able to call wipe at next test case
			session.Close()
		})
	}
}

func TestDeploymentStorageFindUnfinishedByID(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestDeploymentStorageFindUnfinishedByID in short mode.")
	}
	now := time.Now()

	testCases := map[string]struct {
		InputID                    string
		InputDeploymentsCollection []interface{}
		InputTenant                string

		OutputError      error
		OutputDeployment *deployments.Deployment
	}{
		"empty ID": {
			InputID:     "",
			OutputError: ErrStorageInvalidID,
		},
		"empty database": {
			InputID:          "b532b01a-9313-404f-8d19-e7fcbe5cc347",
			OutputError:      nil,
			OutputDeployment: nil,
		},
		"no deployments with given ID": {
			InputID: "b532b01a-9313-404f-8d19-e7fcbe5cc347",
			InputDeploymentsCollection: []interface{}{
				&deployments.Deployment{
					DeploymentConstructor: &deployments.DeploymentConstructor{
						Name:         StringToPointer("NYC Production"),
						ArtifactName: StringToPointer("App 123"),
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
				},
				&deployments.Deployment{
					DeploymentConstructor: &deployments.DeploymentConstructor{
						Name:         StringToPointer("NYC Production"),
						ArtifactName: StringToPointer("App 123"),
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: StringToPointer("d1804903-5caa-4a73-a3ae-0efcc3205405"),
				},
			},
			OutputError:      nil,
			OutputDeployment: nil,
		},
		"all correct": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeploymentsCollection: []interface{}{
				&deployments.Deployment{
					DeploymentConstructor: &deployments.DeploymentConstructor{
						Name:         StringToPointer("NYC Production"),
						ArtifactName: StringToPointer("App 123"),
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
					Stats: map[string]int{
						deployments.DeviceDeploymentStatusDownloading: 0,
						deployments.DeviceDeploymentStatusInstalling:  0,
						deployments.DeviceDeploymentStatusRebooting:   0,
						deployments.DeviceDeploymentStatusPending:     10,
						deployments.DeviceDeploymentStatusSuccess:     15,
						deployments.DeviceDeploymentStatusFailure:     1,
						deployments.DeviceDeploymentStatusNoArtifact:  0,
						deployments.DeviceDeploymentStatusAlreadyInst: 0,
						deployments.DeviceDeploymentStatusAborted:     0,
					},
				},
				&deployments.Deployment{
					DeploymentConstructor: &deployments.DeploymentConstructor{
						Name:         StringToPointer("NYC Production"),
						ArtifactName: StringToPointer("App 123"),
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: StringToPointer("d1804903-5caa-4a73-a3ae-0efcc3205405"),
					Stats: map[string]int{
						deployments.DeviceDeploymentStatusDownloading: 0,
						deployments.DeviceDeploymentStatusInstalling:  0,
						deployments.DeviceDeploymentStatusRebooting:   0,
						deployments.DeviceDeploymentStatusPending:     5,
						deployments.DeviceDeploymentStatusSuccess:     10,
						deployments.DeviceDeploymentStatusFailure:     3,
						deployments.DeviceDeploymentStatusNoArtifact:  0,
						deployments.DeviceDeploymentStatusAlreadyInst: 0,
						deployments.DeviceDeploymentStatusAborted:     0,
					},
				},
			},
			OutputError: nil,
			OutputDeployment: &deployments.Deployment{
				DeploymentConstructor: &deployments.DeploymentConstructor{
					Name:         StringToPointer("NYC Production"),
					ArtifactName: StringToPointer("App 123"),
					//Devices is not kept around!
				},
				Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
				Stats: map[string]int{
					deployments.DeviceDeploymentStatusDownloading: 0,
					deployments.DeviceDeploymentStatusInstalling:  0,
					deployments.DeviceDeploymentStatusRebooting:   0,
					deployments.DeviceDeploymentStatusPending:     10,
					deployments.DeviceDeploymentStatusSuccess:     15,
					deployments.DeviceDeploymentStatusFailure:     1,
					deployments.DeviceDeploymentStatusNoArtifact:  0,
					deployments.DeviceDeploymentStatusAlreadyInst: 0,
					deployments.DeviceDeploymentStatusAborted:     0,
				},
			},
		},
		"deployment already finished": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeploymentsCollection: []interface{}{
				&deployments.Deployment{
					DeploymentConstructor: &deployments.DeploymentConstructor{
						Name:         StringToPointer("NYC Production"),
						ArtifactName: StringToPointer("App 123"),
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id:       StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
					Finished: &now,
				},
				&deployments.Deployment{
					DeploymentConstructor: &deployments.DeploymentConstructor{
						Name:         StringToPointer("NYC Production"),
						ArtifactName: StringToPointer("App 123"),
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: StringToPointer("d1804903-5caa-4a73-a3ae-0efcc3205405"),
				},
			},
			OutputError:      nil,
			OutputDeployment: nil,
		},
		"multi tenant, deployment found": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeploymentsCollection: []interface{}{
				&deployments.Deployment{
					DeploymentConstructor: &deployments.DeploymentConstructor{
						Name:         StringToPointer("NYC Production"),
						ArtifactName: StringToPointer("App 123"),
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
					Stats: map[string]int{
						deployments.DeviceDeploymentStatusPending: 10,
						deployments.DeviceDeploymentStatusSuccess: 15,
						deployments.DeviceDeploymentStatusFailure: 1,
					},
				},
			},
			InputTenant: "acme",
			OutputError: nil,
			OutputDeployment: &deployments.Deployment{
				DeploymentConstructor: &deployments.DeploymentConstructor{
					Name:         StringToPointer("NYC Production"),
					ArtifactName: StringToPointer("App 123"),
					//Devices is not kept around!
				},
				Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
				Stats: map[string]int{
					deployments.DeviceDeploymentStatusPending: 10,
					deployments.DeviceDeploymentStatusSuccess: 15,
					deployments.DeviceDeploymentStatusFailure: 1,
				},
			},
		},
	}

	for testCaseName, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %s", testCaseName), func(t *testing.T) {

			// Make sure we start test with empty database
			db.Wipe()

			session := db.Session()
			store := NewDeploymentsStorage(session)

			ctx := context.Background()
			if testCase.InputTenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: testCase.InputTenant,
				})
			}

			dep := session.DB(ctxstore.DbFromContext(ctx, DatabaseName)).
				C(CollectionDeployments)
			if testCase.InputDeploymentsCollection != nil {
				assert.NoError(t, dep.Insert(testCase.InputDeploymentsCollection...))
			}

			deployment, err := store.FindUnfinishedByID(ctx, testCase.InputID)

			if testCase.OutputError != nil {
				assert.EqualError(t, err, testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.OutputDeployment, deployment)
			}

			// tenant is set, verify that deployment is not present in default DB
			if testCase.InputTenant != "" {
				deployment, err := store.FindUnfinishedByID(context.Background(),
					testCase.InputID)
				assert.Nil(t, deployment)
				assert.Nil(t, err)
			}

			// Need to close all sessions to be able to call wipe at next test case
			session.Close()
		})
	}
}

func TestDeploymentStorageUpdateStats(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDeploymentStorageUpdateStats in short mode.")
	}

	testCases := map[string]struct {
		InputID         string
		InputDeployment *deployments.Deployment
		InputTenant     string

		InputStateFrom string
		InputStateTo   string

		OutputError error
		OutputStats map[string]int
	}{
		"pending -> finished": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &deployments.Deployment{
				Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
				Stats: map[string]int{
					deployments.DeviceDeploymentStatusDownloading: 1,
					deployments.DeviceDeploymentStatusInstalling:  2,
					deployments.DeviceDeploymentStatusRebooting:   3,
					deployments.DeviceDeploymentStatusPending:     10,
					deployments.DeviceDeploymentStatusSuccess:     15,
					deployments.DeviceDeploymentStatusFailure:     4,
					deployments.DeviceDeploymentStatusNoArtifact:  5,
					deployments.DeviceDeploymentStatusAlreadyInst: 0,
					deployments.DeviceDeploymentStatusAborted:     0,
				},
			},
			InputStateFrom: deployments.DeviceDeploymentStatusPending,
			InputStateTo:   deployments.DeviceDeploymentStatusSuccess,

			OutputError: nil,
			OutputStats: map[string]int{
				deployments.DeviceDeploymentStatusDownloading: 1,
				deployments.DeviceDeploymentStatusInstalling:  2,
				deployments.DeviceDeploymentStatusRebooting:   3,
				deployments.DeviceDeploymentStatusPending:     9,
				deployments.DeviceDeploymentStatusSuccess:     16,
				deployments.DeviceDeploymentStatusFailure:     4,
				deployments.DeviceDeploymentStatusNoArtifact:  5,
				deployments.DeviceDeploymentStatusAlreadyInst: 0,
				deployments.DeviceDeploymentStatusAborted:     0,
			},
		},
		"rebooting -> failed": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &deployments.Deployment{
				Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
				Stats: map[string]int{
					deployments.DeviceDeploymentStatusDownloading: 1,
					deployments.DeviceDeploymentStatusInstalling:  2,
					deployments.DeviceDeploymentStatusRebooting:   3,
					deployments.DeviceDeploymentStatusPending:     10,
					deployments.DeviceDeploymentStatusSuccess:     15,
					deployments.DeviceDeploymentStatusFailure:     4,
					deployments.DeviceDeploymentStatusNoArtifact:  5,
					deployments.DeviceDeploymentStatusAlreadyInst: 0,
					deployments.DeviceDeploymentStatusAborted:     0,
				},
			},
			InputStateFrom: deployments.DeviceDeploymentStatusRebooting,
			InputStateTo:   deployments.DeviceDeploymentStatusFailure,

			OutputError: nil,
			OutputStats: map[string]int{
				deployments.DeviceDeploymentStatusDownloading: 1,
				deployments.DeviceDeploymentStatusInstalling:  2,
				deployments.DeviceDeploymentStatusRebooting:   2,
				deployments.DeviceDeploymentStatusPending:     10,
				deployments.DeviceDeploymentStatusSuccess:     15,
				deployments.DeviceDeploymentStatusFailure:     5,
				deployments.DeviceDeploymentStatusNoArtifact:  5,
				deployments.DeviceDeploymentStatusAlreadyInst: 0,
				deployments.DeviceDeploymentStatusAborted:     0,
			},
		},
		"invalid deployment id": {
			InputID:         "",
			InputDeployment: nil,
			InputStateFrom:  deployments.DeviceDeploymentStatusRebooting,
			InputStateTo:    deployments.DeviceDeploymentStatusFailure,

			OutputError: ErrStorageInvalidID,
			OutputStats: nil,
		},
		"wrong deployment id": {
			InputID:         "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: nil,
			InputStateFrom:  deployments.DeviceDeploymentStatusRebooting,
			InputStateTo:    deployments.DeviceDeploymentStatusFailure,

			OutputError: ErrStorageInvalidID,
			OutputStats: nil,
		},
		"no old state": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &deployments.Deployment{
				Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
				Stats: map[string]int{
					deployments.DeviceDeploymentStatusDownloading: 1,
					deployments.DeviceDeploymentStatusInstalling:  2,
					deployments.DeviceDeploymentStatusRebooting:   3,
					deployments.DeviceDeploymentStatusPending:     10,
					deployments.DeviceDeploymentStatusSuccess:     15,
					deployments.DeviceDeploymentStatusFailure:     4,
					deployments.DeviceDeploymentStatusNoArtifact:  5,
					deployments.DeviceDeploymentStatusAlreadyInst: 0,
					deployments.DeviceDeploymentStatusAborted:     0,
				},
			},
			InputStateFrom: "",
			InputStateTo:   deployments.DeviceDeploymentStatusFailure,

			OutputError: ErrStorageInvalidInput,
			OutputStats: nil,
		},
		"install install": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &deployments.Deployment{
				Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
				Stats: map[string]int{
					deployments.DeviceDeploymentStatusDownloading: 1,
					deployments.DeviceDeploymentStatusInstalling:  2,
					deployments.DeviceDeploymentStatusRebooting:   3,
					deployments.DeviceDeploymentStatusPending:     10,
					deployments.DeviceDeploymentStatusSuccess:     15,
					deployments.DeviceDeploymentStatusFailure:     4,
					deployments.DeviceDeploymentStatusNoArtifact:  5,
					deployments.DeviceDeploymentStatusAlreadyInst: 0,
					deployments.DeviceDeploymentStatusAborted:     0,
				},
			},
			InputStateFrom: deployments.DeviceDeploymentStatusInstalling,
			InputStateTo:   deployments.DeviceDeploymentStatusInstalling,

			OutputError: nil,
			OutputStats: map[string]int{
				deployments.DeviceDeploymentStatusDownloading: 1,
				deployments.DeviceDeploymentStatusInstalling:  2,
				deployments.DeviceDeploymentStatusRebooting:   3,
				deployments.DeviceDeploymentStatusPending:     10,
				deployments.DeviceDeploymentStatusSuccess:     15,
				deployments.DeviceDeploymentStatusFailure:     4,
				deployments.DeviceDeploymentStatusNoArtifact:  5,
				deployments.DeviceDeploymentStatusAlreadyInst: 0,
				deployments.DeviceDeploymentStatusAborted:     0,
			},
		},
		"tenant, pending -> finished": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &deployments.Deployment{
				Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
				Stats: map[string]int{
					deployments.DeviceDeploymentStatusDownloading: 1,
					deployments.DeviceDeploymentStatusInstalling:  2,
					deployments.DeviceDeploymentStatusRebooting:   3,
					deployments.DeviceDeploymentStatusPending:     10,
					deployments.DeviceDeploymentStatusSuccess:     15,
					deployments.DeviceDeploymentStatusFailure:     4,
					deployments.DeviceDeploymentStatusNoArtifact:  5,
					deployments.DeviceDeploymentStatusAlreadyInst: 0,
					deployments.DeviceDeploymentStatusAborted:     0,
				},
			},
			InputTenant: "acme",

			InputStateFrom: deployments.DeviceDeploymentStatusPending,
			InputStateTo:   deployments.DeviceDeploymentStatusSuccess,

			OutputError: nil,
			OutputStats: map[string]int{
				deployments.DeviceDeploymentStatusDownloading: 1,
				deployments.DeviceDeploymentStatusInstalling:  2,
				deployments.DeviceDeploymentStatusRebooting:   3,
				deployments.DeviceDeploymentStatusPending:     9,
				deployments.DeviceDeploymentStatusSuccess:     16,
				deployments.DeviceDeploymentStatusFailure:     4,
				deployments.DeviceDeploymentStatusNoArtifact:  5,
				deployments.DeviceDeploymentStatusAlreadyInst: 0,
				deployments.DeviceDeploymentStatusAborted:     0,
			},
		},
	}

	for testCaseName, tc := range testCases {
		t.Run(fmt.Sprintf("test case %s", testCaseName), func(t *testing.T) {

			db.Wipe()

			session := db.Session()
			store := NewDeploymentsStorage(session)

			ctx := context.Background()
			if tc.InputTenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.InputTenant,
				})
			}

			if tc.InputDeployment != nil {
				err := session.DB(DatabaseName).
					C(CollectionDeployments).
					Insert(tc.InputDeployment)
				assert.NoError(t, err)
				// multi tenant test only makes sense if there
				// is a deployment to input, if there's one
				// we'll add it to tenant's DB
				if tc.InputTenant != "" {
					err = session.DB(ctxstore.DbFromContext(ctx, DatabaseName)).
						C(CollectionDeployments).
						Insert(tc.InputDeployment)
					assert.NoError(t, err)
				}
			}

			err := store.UpdateStats(ctx,
				tc.InputID, tc.InputStateFrom, tc.InputStateTo)

			if tc.OutputError != nil {
				assert.EqualError(t, err, tc.OutputError.Error())
			} else {
				var deployment *deployments.Deployment
				err := session.DB(ctxstore.DbFromContext(ctx, DatabaseName)).
					C(CollectionDeployments).
					FindId(tc.InputID).One(&deployment)
				assert.NoError(t, err)
				assert.Equal(t, tc.OutputStats, deployment.Stats)

				// if there's a tenant, verify that deployment
				// in default DB remains unchanged, again only
				// makes sense if there's an input deployment
				if tc.InputTenant != "" && tc.InputDeployment != nil {
					var defDeployment *deployments.Deployment
					err := session.DB(DatabaseName).
						C(CollectionDeployments).
						FindId(tc.InputID).One(&defDeployment)
					assert.NoError(t, err)
					assert.Equal(t, defDeployment.Stats, tc.InputDeployment.Stats)
				}

			}

			// Need to close all sessions to be able to call wipe at next test case
			session.Close()
		})
	}
}

func TestDeploymentStorageUpdateStatsAndFinishDeployment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDeploymentStorageUpdateStatsAndFinishDeployment in short mode.")
	}

	testCases := map[string]struct {
		InputID         string
		InputDeployment *deployments.Deployment
		InputStats      map[string]int
		InputTenant     string

		OutputError error
	}{
		"all correct": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &deployments.Deployment{
				Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
				Stats: map[string]int{
					deployments.DeviceDeploymentStatusDownloading: 1,
					deployments.DeviceDeploymentStatusInstalling:  2,
					deployments.DeviceDeploymentStatusRebooting:   3,
					deployments.DeviceDeploymentStatusPending:     3,
					deployments.DeviceDeploymentStatusSuccess:     6,
					deployments.DeviceDeploymentStatusFailure:     8,
					deployments.DeviceDeploymentStatusNoArtifact:  4,
					deployments.DeviceDeploymentStatusAlreadyInst: 2,
					deployments.DeviceDeploymentStatusAborted:     5,
				},
			},
			InputStats: map[string]int{
				deployments.DeviceDeploymentStatusDownloading: 1,
				deployments.DeviceDeploymentStatusInstalling:  2,
				deployments.DeviceDeploymentStatusRebooting:   3,
				deployments.DeviceDeploymentStatusPending:     10,
				deployments.DeviceDeploymentStatusSuccess:     15,
				deployments.DeviceDeploymentStatusFailure:     4,
				deployments.DeviceDeploymentStatusNoArtifact:  5,
				deployments.DeviceDeploymentStatusAlreadyInst: 0,
				deployments.DeviceDeploymentStatusAborted:     5,
			},

			OutputError: nil,
		},
		"invalid deployment id": {
			InputID:         "",
			InputDeployment: nil,
			InputStats:      nil,

			OutputError: ErrStorageInvalidID,
		},
		"wrong deployment id": {
			InputID:         "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: nil,
			InputStats:      nil,

			OutputError: ErrStorageInvalidID,
		},
		"tenant, all correct": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &deployments.Deployment{
				Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
				Stats: newTestStats(deployments.Stats{
					deployments.DeviceDeploymentStatusRebooting: 3,
				}),
			},
			InputStats: newTestStats(deployments.Stats{
				deployments.DeviceDeploymentStatusRebooting: 3,
			}),
			InputTenant: "acme",

			OutputError: nil,
		},
	}

	for testCaseName, tc := range testCases {
		t.Run(fmt.Sprintf("test case %s", testCaseName), func(t *testing.T) {

			db.Wipe()

			session := db.Session()
			store := NewDeploymentsStorage(session)

			ctx := context.Background()
			if tc.InputTenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.InputTenant,
				})
			}

			dep := session.DB(ctxstore.DbFromContext(ctx, DatabaseName)).
				C(CollectionDeployments)
			if tc.InputDeployment != nil {
				assert.NoError(t, dep.Insert(tc.InputDeployment))
			}

			err := store.UpdateStatsAndFinishDeployment(ctx,
				tc.InputID, tc.InputStats)

			if tc.OutputError != nil {
				assert.EqualError(t, err, tc.OutputError.Error())
			} else {
				var deployment *deployments.Deployment
				err := session.DB(ctxstore.DbFromContext(ctx, DatabaseName)).
					C(CollectionDeployments).
					FindId(tc.InputID).One(&deployment)
				assert.NoError(t, err)
				assert.Equal(t, tc.InputStats, deployment.Stats)
			}

			if tc.InputTenant != "" && tc.InputDeployment != nil {
				// tenant is configured, so deployments that are
				// part of test input were added to tenant's DB,
				// trying to update them in default DB will
				// raise an error
				err := store.UpdateStatsAndFinishDeployment(context.Background(),
					tc.InputID, tc.InputStats)
				assert.EqualError(t, err, ErrStorageInvalidID.Error())
			}

			// Need to close all sessions to be able to call wipe at next test case
			session.Close()
		})
	}
}

func newTestStats(stats deployments.Stats) deployments.Stats {
	st := deployments.NewDeviceDeploymentStats()
	for k, v := range stats {
		st[k] = v
	}
	return st
}

func TestDeploymentStorageFindBy(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping TestDeploymentStorageFindBy in short mode.")
	}

	someDeployments := []*deployments.Deployment{
		{
			DeploymentConstructor: &deployments.DeploymentConstructor{
				Name:         StringToPointer("NYC Production Inc."),
				ArtifactName: StringToPointer("App 123"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
			Stats: newTestStats(deployments.Stats{
				deployments.DeviceDeploymentStatusNoArtifact: 1,
			}),
		},
		{
			DeploymentConstructor: &deployments.DeploymentConstructor{
				Name:         StringToPointer("NYC Production Inc."),
				ArtifactName: StringToPointer("App 123"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: StringToPointer("d1804903-5caa-4a73-a3ae-0efcc3205405"),
			Stats: newTestStats(deployments.Stats{
				deployments.DeviceDeploymentStatusNoArtifact: 1,
			}),
		},
		{
			DeploymentConstructor: &deployments.DeploymentConstructor{
				Name:         StringToPointer("foo"),
				ArtifactName: StringToPointer("bar"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: StringToPointer("e8c32ff6-7c1b-43c7-aa31-2e4fc3a3c130"),
			Stats: newTestStats(deployments.Stats{
				deployments.DeviceDeploymentStatusFailure: 2,
			}),
		},
		{
			DeploymentConstructor: &deployments.DeploymentConstructor{
				Name:         StringToPointer("foo"),
				ArtifactName: StringToPointer("bar"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: StringToPointer("3fe15222-0a41-401f-8f5e-582aba2a002c"),
			Stats: newTestStats(deployments.Stats{
				deployments.DeviceDeploymentStatusNoArtifact: 1,
			}),
		},
		{
			DeploymentConstructor: &deployments.DeploymentConstructor{
				Name:         StringToPointer("foo"),
				ArtifactName: StringToPointer("bar"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: StringToPointer("3fe15222-0a41-401f-8f5e-582aba2a002d"),
			Stats: newTestStats(deployments.Stats{
				deployments.DeviceDeploymentStatusDownloading: 1,
			}),
		},
		{
			DeploymentConstructor: &deployments.DeploymentConstructor{
				Name:         StringToPointer("zed"),
				ArtifactName: StringToPointer("daz"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: StringToPointer("3fe15222-1234-401f-8f5e-582aba2a002e"),
			Stats: newTestStats(deployments.Stats{
				deployments.DeviceDeploymentStatusDownloading: 1,
				deployments.DeviceDeploymentStatusPending:     1,
			}),
		},
		{
			DeploymentConstructor: &deployments.DeploymentConstructor{
				Name:         StringToPointer("zed"),
				ArtifactName: StringToPointer("daz"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: StringToPointer("3fe15222-1234-401f-8f5e-582aba2a002f"),
			Stats: newTestStats(deployments.Stats{
				deployments.DeviceDeploymentStatusPending: 1,
			}),
		},
		{
			DeploymentConstructor: &deployments.DeploymentConstructor{
				Name:         StringToPointer("zed"),
				ArtifactName: StringToPointer("daz"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: StringToPointer("44dd8822-eeb1-44db-a18e-f4f5acc43796"),
			Stats: newTestStats(deployments.Stats{
				deployments.DeviceDeploymentStatusNoArtifact: 1,
				deployments.DeviceDeploymentStatusSuccess:    1,
			}),
		},
		{
			DeploymentConstructor: &deployments.DeploymentConstructor{
				Name:         StringToPointer("123"),
				ArtifactName: StringToPointer("dfs"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc34a"},
			},
			Id: StringToPointer("3fe15222-1234-401f-8f5e-582aba2a002a"),
			Stats: newTestStats(deployments.Stats{
				deployments.DeviceDeploymentStatusAborted: 1,
			}),
		},

		//in progress deployment, with only pending and already-installed counters > 0
		{
			DeploymentConstructor: &deployments.DeploymentConstructor{
				Name:         StringToPointer("baz"),
				ArtifactName: StringToPointer("asdf"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: StringToPointer("12345678-0a41-401f-8f5e-582aba2a002d"),
			Stats: newTestStats(deployments.Stats{
				deployments.DeviceDeploymentStatusPending:     1,
				deployments.DeviceDeploymentStatusAlreadyInst: 1,
			}),
		},
		//in progress deployment, with only pending and success counters > 0
		{
			DeploymentConstructor: &deployments.DeploymentConstructor{
				Name:         StringToPointer("baz"),
				ArtifactName: StringToPointer("asdf"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: StringToPointer("22345678-0a41-401f-8f5e-582aba2a002d"),
			Stats: newTestStats(deployments.Stats{
				deployments.DeviceDeploymentStatusPending: 1,
				deployments.DeviceDeploymentStatusSuccess: 1,
			}),
		},
		//in progress deployment, with only pending and failure counters > 0
		{
			DeploymentConstructor: &deployments.DeploymentConstructor{
				Name:         StringToPointer("baz"),
				ArtifactName: StringToPointer("asdf"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: StringToPointer("32345678-0a41-401f-8f5e-582aba2a002d"),
			Stats: newTestStats(deployments.Stats{
				deployments.DeviceDeploymentStatusPending: 1,
				deployments.DeviceDeploymentStatusFailure: 1,
			}),
		},
		//in progress deployment, with only pending and noartifact counters > 0
		{
			DeploymentConstructor: &deployments.DeploymentConstructor{
				Name:         StringToPointer("baz"),
				ArtifactName: StringToPointer("asdf"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: StringToPointer("42345678-0a41-401f-8f5e-582aba2a002d"),
			Stats: newTestStats(deployments.Stats{
				deployments.DeviceDeploymentStatusPending:    1,
				deployments.DeviceDeploymentStatusNoArtifact: 1,
			}),
		},
		//finished deployment, with only already installed counter > 0
		{
			DeploymentConstructor: &deployments.DeploymentConstructor{
				Name:         StringToPointer("baz"),
				ArtifactName: StringToPointer("asdf"),
				Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
			},
			Id: StringToPointer("52345678-0a41-401f-8f5e-582aba2a002d"),
			Stats: newTestStats(deployments.Stats{
				deployments.DeviceDeploymentStatusAlreadyInst: 1,
			}),
		},
	}

	testCases := []struct {
		InputModelQuery            deployments.Query
		InputDeploymentsCollection []*deployments.Deployment
		InputTenant                string

		OutputError error
		OutputID    []string
	}{
		{
			InputModelQuery: deployments.Query{
				SearchText: "foobar-empty-db",
			},
			OutputError: ErrDeploymentStorageCannotExecQuery,
		},
		{
			InputModelQuery: deployments.Query{
				SearchText: "foobar-no-match",
			},
			InputDeploymentsCollection: []*deployments.Deployment{
				{
					DeploymentConstructor: &deployments.DeploymentConstructor{
						Name:         StringToPointer("NYC Production"),
						ArtifactName: StringToPointer("App 123"),
						Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc347"},
					},
					Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
				},
			},
		},
		{
			InputModelQuery: deployments.Query{
				SearchText: "NYC",
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				"d1804903-5caa-4a73-a3ae-0efcc3205405",
			},
		},
		{
			InputModelQuery: deployments.Query{
				SearchText: "NYC foo",
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				"d1804903-5caa-4a73-a3ae-0efcc3205405",
				"e8c32ff6-7c1b-43c7-aa31-2e4fc3a3c130",
				"3fe15222-0a41-401f-8f5e-582aba2a002c",
				"3fe15222-0a41-401f-8f5e-582aba2a002d",
			},
		},
		{
			InputModelQuery: deployments.Query{
				SearchText: "bar",
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"e8c32ff6-7c1b-43c7-aa31-2e4fc3a3c130",
				"3fe15222-0a41-401f-8f5e-582aba2a002c",
				"3fe15222-0a41-401f-8f5e-582aba2a002d",
			},
		},
		{
			InputModelQuery: deployments.Query{
				SearchText: "bar",
				Status:     deployments.StatusQueryInProgress,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"3fe15222-0a41-401f-8f5e-582aba2a002d",
			},
		},
		{
			InputModelQuery: deployments.Query{
				SearchText: "bar",
				Status:     deployments.StatusQueryFinished,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"e8c32ff6-7c1b-43c7-aa31-2e4fc3a3c130",
				"3fe15222-0a41-401f-8f5e-582aba2a002c",
			},
		},
		{
			InputModelQuery: deployments.Query{
				Status: deployments.StatusQueryInProgress,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"3fe15222-0a41-401f-8f5e-582aba2a002d",
				"3fe15222-1234-401f-8f5e-582aba2a002e",
				"12345678-0a41-401f-8f5e-582aba2a002d",
				"22345678-0a41-401f-8f5e-582aba2a002d",
				"32345678-0a41-401f-8f5e-582aba2a002d",
				"42345678-0a41-401f-8f5e-582aba2a002d",
			},
		},
		{
			InputModelQuery: deployments.Query{
				Status: deployments.StatusQueryPending,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"3fe15222-1234-401f-8f5e-582aba2a002f",
			},
		},
		{
			InputModelQuery: deployments.Query{
				Status: deployments.StatusQueryFinished,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				"d1804903-5caa-4a73-a3ae-0efcc3205405",
				"e8c32ff6-7c1b-43c7-aa31-2e4fc3a3c130",
				"3fe15222-0a41-401f-8f5e-582aba2a002c",
				"44dd8822-eeb1-44db-a18e-f4f5acc43796",
				"3fe15222-1234-401f-8f5e-582aba2a002a",
				"52345678-0a41-401f-8f5e-582aba2a002d",
			},
		},
		{
			InputModelQuery: deployments.Query{
				// whatever name
				SearchText: "",
				// any status
				Status: deployments.StatusQueryAny,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				"d1804903-5caa-4a73-a3ae-0efcc3205405",
				"e8c32ff6-7c1b-43c7-aa31-2e4fc3a3c130",
				"3fe15222-0a41-401f-8f5e-582aba2a002c",
				"3fe15222-0a41-401f-8f5e-582aba2a002d",
				"3fe15222-1234-401f-8f5e-582aba2a002e",
				"3fe15222-1234-401f-8f5e-582aba2a002f",
				"44dd8822-eeb1-44db-a18e-f4f5acc43796",
				"3fe15222-1234-401f-8f5e-582aba2a002a",
				"12345678-0a41-401f-8f5e-582aba2a002d",
				"22345678-0a41-401f-8f5e-582aba2a002d",
				"32345678-0a41-401f-8f5e-582aba2a002d",
				"42345678-0a41-401f-8f5e-582aba2a002d",
				"52345678-0a41-401f-8f5e-582aba2a002d",
			},
		},
		{
			InputModelQuery: deployments.Query{
				// whatever name
				SearchText: "",
				// any status
				Status: deployments.StatusQueryAny,
				Limit:  2,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"12345678-0a41-401f-8f5e-582aba2a002d",
				"22345678-0a41-401f-8f5e-582aba2a002d",
			},
		},
		{
			InputModelQuery: deployments.Query{
				// whatever name
				SearchText: "",
				// any status
				Status: deployments.StatusQueryAny,
				Limit:  2,
				Skip:   2,
			},
			InputDeploymentsCollection: someDeployments,
			OutputError:                nil,
			OutputID: []string{
				"32345678-0a41-401f-8f5e-582aba2a002d",
				"3fe15222-0a41-401f-8f5e-582aba2a002c",
			},
		},
		{
			InputModelQuery: deployments.Query{
				SearchText: "NYC",
			},
			InputDeploymentsCollection: someDeployments,
			InputTenant:                "acme",
			OutputError:                nil,
			OutputID: []string{
				"a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
				"d1804903-5caa-4a73-a3ae-0efcc3205405",
			},
		},
	}

	for testCaseNumber, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %d", testCaseNumber+1), func(t *testing.T) {
			t.Logf("testing search: '%s'", testCase.InputModelQuery.SearchText)
			t.Logf("        status: %v", testCase.InputModelQuery.Status)

			// Make sure we start test with empty database
			db.Wipe()

			session := db.Session()
			store := NewDeploymentsStorage(session)

			ctx := context.Background()
			if testCase.InputTenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: testCase.InputTenant,
				})
			}

			for _, d := range testCase.InputDeploymentsCollection {
				if d.Created == nil {
					now := time.Now()
					d.Created = &now
				}
				assert.NoError(t, store.Insert(ctx, d))
			}

			deps, err := store.Find(ctx,
				testCase.InputModelQuery)

			if testCase.OutputError != nil {
				assert.EqualError(t, err,
					testCase.OutputError.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, deps, len(testCase.OutputID))
				for _, dep := range deps {
					assert.Contains(t, testCase.OutputID,
						*dep.Id,
						"got unexpected deployment %s",
						*dep.Id)
				}
			}

			if testCase.InputTenant != "" {
				// have to add a deployment, otherwise, it won't
				// be possible to run find queries
				err := store.Insert(context.Background(),
					&deployments.Deployment{
						DeploymentConstructor: &deployments.DeploymentConstructor{
							Name:         StringToPointer("foo-" + testCase.InputTenant),
							ArtifactName: StringToPointer("bar-" + testCase.InputTenant),
							Devices:      []string{"b532b01a-9313-404f-8d19-e7fcbe5cc399"},
						},
						Id:      StringToPointer("e8c32ff6-7c1b-43c7-aa31-2e4fc3a3c199"),
						Stats:   newTestStats(deployments.Stats{}),
						Created: TimeToPointer(time.Now().UTC()),
					})
				assert.NoError(t, err)

				// tenant is set, so only tenant's DB was set
				// up, verify that we cannot find anything in
				// default DB
				deps, err := store.Find(context.Background(),
					testCase.InputModelQuery)
				assert.Len(t, deps, 0)
				assert.NoError(t, err)
			}

			// Need to close all sessions to be able to call wipe at next test case
			session.Close()
		})
	}
}

func TestDeploymentFinish(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDeploymentFinish in short mode.")
	}

	testCases := map[string]struct {
		InputID         string
		InputDeployment *deployments.Deployment
		InputTenant     string

		OutputError error
	}{
		"finished": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &deployments.Deployment{
				Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
			},
			OutputError: nil,
		},
		"nonexistent": {
			InputID:     "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			OutputError: errors.New("Invalid id"),
		},
		"tenant, finished": {
			InputID: "a108ae14-bb4e-455f-9b40-2ef4bab97bb7",
			InputDeployment: &deployments.Deployment{
				Id: StringToPointer("a108ae14-bb4e-455f-9b40-2ef4bab97bb7"),
			},
			InputTenant: "acme",
			OutputError: nil,
		},
	}

	for testCaseName, tc := range testCases {
		t.Run(fmt.Sprintf("test case %s", testCaseName), func(t *testing.T) {

			db.Wipe()

			session := db.Session()
			store := NewDeploymentsStorage(session)

			ctx := context.Background()
			if tc.InputTenant != "" {
				ctx = identity.WithContext(ctx, &identity.Identity{
					Tenant: tc.InputTenant,
				})
			}

			if tc.InputDeployment != nil {
				dep := session.DB(ctxstore.DbFromContext(ctx, DatabaseName)).
					C(CollectionDeployments)
				assert.NoError(t, dep.Insert(tc.InputDeployment))
			}

			now := time.Now()
			err := store.Finish(ctx, tc.InputID, now)

			if tc.OutputError != nil {
				assert.EqualError(t, err, tc.OutputError.Error())
			} else {
				var deployment *deployments.Deployment
				err := session.DB(ctxstore.DbFromContext(ctx, DatabaseName)).
					C(CollectionDeployments).
					FindId(tc.InputID).One(&deployment)
				assert.NoError(t, err)

				if assert.NotNil(t, deployment.Finished) {
					// mongo might have trimmed our time a
					// bit, let's check that we are within a
					// 1s range
					assert.WithinDuration(t, now, *deployment.Finished, time.Second)
				}
			}

			if tc.InputTenant != "" {
				// deployment was added to tenant's DB, so this
				// should fail with default DB
				err := store.Finish(context.Background(), tc.InputID, now)
				assert.EqualError(t, err, ErrStorageInvalidID.Error())
			}
			// Need to close all sessions to be able to call wipe at next test case
			session.Close()
		})
	}
}
