package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/cox96de/runner/githubapp/ghclient"

	"github.com/cox96de/runner/githubapp/db"
	mockserver "github.com/cox96de/runner/mock/server"
	"github.com/google/go-github/v64/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
)

func TestApp_handleCheckSuite(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		installTokenMatch := mock.WithRequestMatch(mock.PostAppInstallationsAccessTokensByInstallationId, &github.InstallationToken{
			Token: lo.ToPtr("install_token"),
		})
		repoContentMatch := mock.WithRequestMatch(mock.GetReposContentsByOwnerByRepoByPath, &github.RepositoryContent{
			Content: lo.ToPtr(`
jobs:
  job1:
     name: "job_name"
     runs-on: 
       container-image: "debian"
     steps:
       - name: "step1"
         run: 
           - "echo hello"`),
		})
		client := mock.NewMockedHTTPClient(
			installTokenMatch,
			repoContentMatch,
			mock.WithRequestMatch(mock.PostReposCheckRunsByOwnerByRepo, &github.CheckRun{ID: lo.ToPtr(int64(1))},
				&github.CheckRun{ID: lo.ToPtr(int64(1))}, &github.CheckRun{ID: lo.ToPtr(int64(1))}),
		)
		runnerServer := mockserver.NewMockServer(t)
		dbConn := db.NewMockDB(t)
		githubClient := github.NewClient(client)
		ghClient := ghclient.NewClient(githubClient)
		ghClient.SetDefaultHTTPClient(client)
		app := NewApp(ghClient, "http://base.url", dbConn, "git clone", nil)
		app.SetRunnerServer(runnerServer.App)
		err := app.handleCheckSuite(context.Background(), &github.CheckSuiteEvent{
			Action: lo.ToPtr("requested"),
			Repo: &github.Repository{
				Name: lo.ToPtr("runner"),
				Owner: &github.User{
					Login: lo.ToPtr("cox96de"),
				},
			},
		})
		assert.NilError(t, err)
		// TODO: check created data.
	})
	t.Run("no_ci_file", func(t *testing.T) {
		installTokenMatch := mock.WithRequestMatch(mock.PostAppInstallationsAccessTokensByInstallationId, &github.InstallationToken{
			Token: lo.ToPtr("install_token"),
		})
		repoContentMatch := mock.WithRequestMatchHandler(mock.GetReposContentsByOwnerByRepoByPath, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			mock.WriteError(writer, http.StatusNotFound, "Not Found")
		}))
		client := mock.NewMockedHTTPClient(
			installTokenMatch,
			repoContentMatch,
		)
		runnerServer := mockserver.NewMockServer(t)
		dbConn := db.NewMockDB(t)
		githubClient := github.NewClient(client)
		ghClient := ghclient.NewClient(githubClient)
		ghClient.SetDefaultHTTPClient(client)
		app := NewApp(ghClient, "http://base.url", dbConn, "git clone", nil)
		app.SetRunnerServer(runnerServer.App)
		err := app.handleCheckSuite(context.Background(), &github.CheckSuiteEvent{
			Action: lo.ToPtr("requested"),
			Repo: &github.Repository{
				Name: lo.ToPtr("runner"),
				Owner: &github.User{
					Login: lo.ToPtr("cox96de"),
				},
			},
		})
		assert.NilError(t, err)
		// TODO: check created data.
	})
}
