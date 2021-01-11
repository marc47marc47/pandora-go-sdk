package logdb

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/marc47marc47/pandora-go-sdk/base"
	"github.com/marc47marc47/pandora-go-sdk/base/config"
	"github.com/marc47marc47/pandora-go-sdk/base/models"
	"github.com/marc47marc47/pandora-go-sdk/base/reqerr"
	"github.com/marc47marc47/pandora-go-sdk/logdb"
	"github.com/stretchr/testify/assert"
)

var (
	cfg               *config.Config
	client            logdb.LogdbAPI
	region            = os.Getenv("REGION")
	endpoint          = os.Getenv("LOGDB_HOST")
	ak                = os.Getenv("ACCESS_KEY")
	sk                = os.Getenv("SECRET_KEY")
	logger            base.Logger
	defaultRepoSchema []logdb.RepoSchemaEntry
)

func init() {
	var err error
	logger = base.NewDefaultLogger()
	cfg = logdb.NewConfig().
		WithEndpoint(endpoint).
		WithAccessKeySecretKey(ak, sk).
		WithLogger(logger).
		WithLoggerLevel(base.LogDebug)

	client, err = logdb.New(cfg)
	if err != nil {
		logger.Errorf("new logdb client failed, err: %v", err)
	}

	defaultRepoSchema = []logdb.RepoSchemaEntry{
		logdb.RepoSchemaEntry{
			Key:       "f1",
			ValueType: "string",
			Analyzer:  "standard",
		},
		logdb.RepoSchemaEntry{
			Key:       "f2",
			ValueType: "float",
		},
		logdb.RepoSchemaEntry{
			Key:       "f3",
			ValueType: "date",
		},
		logdb.RepoSchemaEntry{
			Key:       "f4",
			ValueType: "long",
		},
	}
}

func TestRepo(t *testing.T) {
	repoName := "repo_sdk_test"
	createInput := &logdb.CreateRepoInput{
		RepoName:  repoName,
		Region:    region,
		Schema:    defaultRepoSchema,
		Retention: "2d",
	}

	err := client.CreateRepo(createInput)
	if err != nil {
		t.Error(err)
	}

	getOutput, err := client.GetRepo(&logdb.GetRepoInput{RepoName: repoName})
	if err != nil {
		t.Error(err)
	}
	if getOutput == nil {
		t.Error("repo ret is empty")
	}
	if region != getOutput.Region {
		t.Errorf("unexpected region: %s", region)
	}
	assert.Equal(t, defaultRepoSchema, getOutput.Schema)
	if getOutput.Retention != "2d" {
		t.Errorf("retention should be 2d but %s", getOutput.Retention)
	}

	updateInput := &logdb.UpdateRepoInput{
		RepoName:  repoName,
		Schema:    defaultRepoSchema,
		Retention: "3d",
	}

	err = client.UpdateRepo(updateInput)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Second)
	getOutput, err = client.GetRepo(&logdb.GetRepoInput{RepoName: repoName})
	if err != nil {
		t.Error(err)
	}
	if getOutput == nil {
		t.Error("schema ret is empty")
	}
	if "nb" != getOutput.Region {
		t.Error("region should be nb", getOutput.Region)
	}
	assert.Equal(t, defaultRepoSchema, getOutput.Schema)
	if getOutput.Retention != "3d" {
		t.Errorf("retention should be 3d but %s", getOutput.Retention)
	}

	listOutput, err := client.ListRepos(&logdb.ListReposInput{})
	if err != nil {
		t.Error(err)
	}
	if listOutput == nil {
		t.Error("repo list should not be empty")
	}
	if listOutput.Repos[0].RepoName != repoName {
		t.Error("repo name is different to origin name")
		t.Error(listOutput.Repos[0].RepoName)
	}

	err = client.DeleteRepo(&logdb.DeleteRepoInput{RepoName: repoName})
	if err != nil {
		t.Error(err)
	}
}

func TestSendAndQueryLog(t *testing.T) {
	repoName := "repo_send_log"
	createInput := &logdb.CreateRepoInput{
		RepoName:  repoName,
		Region:    region,
		Schema:    defaultRepoSchema,
		Retention: "2d",
	}
	err := client.CreateRepo(createInput)
	if err != nil {
		t.Error(err)
	}

	startTime := time.Now().Unix() * 1000
	for i := 0; i < 5; i++ {
		sendLogInput := &logdb.SendLogInput{
			RepoName:       repoName,
			OmitInvalidLog: false,
			Logs: logdb.Logs{
				logdb.Log{
					"f1": "v11",
					"f2": 1.0,
					"f3": time.Now().UTC().Format(time.RFC3339),
					"f4": 1312,
				},
				logdb.Log{
					"f1": "v21",
					"f2": 1.2,
					"f3": time.Now().UTC().Format(time.RFC3339),
					"f4": 3082,
				},
				logdb.Log{
					"f1": "v31",
					"f2": 0.0,
					"f3": time.Now().UTC().Format(time.RFC3339),
					"f4": 1,
				},
				logdb.Log{
					"f1": "v41",
					"f2": 0.3,
					"f3": time.Now().UTC().Format(time.RFC3339),
					"f4": 12345671,
				},
			},
		}
		sendOutput, err := client.SendLog(sendLogInput)
		if err != nil {
			t.Error(err)
		}
		if sendOutput.Success != 4 || sendOutput.Failed != 0 || sendOutput.Total != 4 {
			t.Errorf("send log failed, success: %d, failed: %d, total: %d", sendOutput.Success, sendOutput.Failed, sendOutput.Total)
		}
		time.Sleep(10 * time.Second)
	}
	endTime := time.Now().Unix() * 1000
	time.Sleep(2 * time.Minute)

	histogramInput := &logdb.QueryHistogramLogInput{
		RepoName: repoName,
		Query:    "",
		From:     startTime,
		To:       endTime,
		Field:    "f3",
	}
	histogramOutput, err := client.QueryHistogramLog(histogramInput)
	if err != nil {
		t.Error(err)
	}
	if histogramOutput.Total != 20 {
		t.Errorf("log count should be 20, but %d", histogramOutput.Total)
	}
	if histogramOutput.PartialSuccess != false {
		t.Error("query partialSuccess should be false, but true")
	}
	if len(histogramOutput.Buckets) < 5 || len(histogramOutput.Buckets) > 20 {
		t.Errorf("histogram count should ge 5 and le 20, but %d", len(histogramOutput.Buckets))
	}

	queryInput := &logdb.QueryLogInput{
		RepoName: repoName,
		Query:    "f3:[2016-01-01 TO 2036-01-02]",
		Sort:     "f2:desc",
		From:     0,
		Size:     100,
	}
	queryOut, err := client.QueryLog(queryInput)
	if err != nil {
		t.Error(err)
	}
	if queryOut.Total != 20 {
		t.Errorf("log count should be 20, but %d", queryOut.Total)
	}
	if queryOut.PartialSuccess != false {
		t.Error("query partialSuccess should be false, but true")
	}
	if len(queryOut.Data) != 20 {
		t.Errorf("log count should be 20, but %d", len(queryOut.Data))
	}
	if len(queryOut.ScrollId) != 0 {
		t.Errorf("log scroll_id should be empty, but %v", len(queryOut.ScrollId))
	}

	queryInputWithScroll := queryInput
	queryInputWithScroll.Size = 12
	queryInputWithScroll.Scroll = "2m"

	queryOutWithScroll, err := client.QueryLog(queryInputWithScroll)
	if err != nil {
		t.Error(err)
	}
	if len(queryOutWithScroll.ScrollId) == 0 {
		t.Errorf("log scroll_id should NOT be empty, but %v", len(queryOut.ScrollId))
	}
	if len(queryOut.Data) != 12 {
		t.Errorf("log count should be 12, but %d", len(queryOut.Data))
	}

	scrollInput := &logdb.QueryScrollInput{
		RepoName: repoName,
		ScrollId: queryOutWithScroll.ScrollId,
		Scroll:   "1m",
	}
	scrollOut, err := client.QueryScroll(scrollInput)
	if err != nil {
		t.Error(err)
	}
	if len(scrollOut.Data) != 8 {
		t.Errorf("log count should be 8, but %d", len(scrollOut.Data))
	}
	if scrollOut.Total != 20 {
		t.Errorf("log total count should be 20, but %d", scrollOut.Total)
	}

	err = client.DeleteRepo(&logdb.DeleteRepoInput{RepoName: repoName})
	if err != nil {
		t.Error(err)
	}
}

func TestSendLogWithToken(t *testing.T) {
	repoName := "repo_send_log_with_token"
	createInput := &logdb.CreateRepoInput{
		RepoName:  repoName,
		Region:    region,
		Schema:    defaultRepoSchema,
		Retention: "2d",
	}
	err := client.CreateRepo(createInput)
	if err != nil {
		t.Error(err)
	}

	td := &base.TokenDesc{}
	td.Expires = time.Now().Unix() + 10
	td.Method = "POST"
	td.Url = "/v5/repos/repo_send_log_with_token/data"
	td.ContentType = "application/json"

	token, err := client.MakeToken(td)
	if err != nil {
		t.Error(err)
	}

	cfg2 := logdb.NewConfig().WithEndpoint(endpoint)

	client2, err2 := logdb.New(cfg2)
	if err2 != nil {
		logger.Error("new logdb client failed, err: %v", err2)
	}
	sendLogInput := &logdb.SendLogInput{
		RepoName:       repoName,
		OmitInvalidLog: false,
		Logs: logdb.Logs{
			logdb.Log{
				"f1": "v11",
				"f2": 1.0,
				"f3": time.Now().UTC().Format(time.RFC3339),
				"f4": 12,
			},
			logdb.Log{
				"f1": "v21",
				"f2": 1.2,
				"f3": time.Now().UTC().Format(time.RFC3339),
				"f4": 2,
			},
		},
		PandoraToken: models.PandoraToken{
			Token: token,
		},
	}
	_, err = client.SendLog(sendLogInput)
	if err != nil {
		t.Error(err)
	}

	time.Sleep(15 * time.Second)

	_, err = client2.SendLog(sendLogInput)
	if err == nil {
		t.Errorf("expired token: %s, expires: %d, now: %d", token, td.Expires, time.Now().Unix())
	}
	v, ok := err.(*reqerr.RequestError)
	if !ok {
		t.Errorf("cast err to UnauthorizedError fail, err: %v", err)
	}

	if v.ErrorType != reqerr.UnauthorizedError {
		t.Errorf("got errorType: %d, expected errorType: %d", v.ErrorType, reqerr.UnauthorizedError)
	}

	if v.StatusCode != 401 {
		t.Errorf("expires token, expires: %d, now: %d", td.Expires, time.Now().Unix())
	}

	err = client.DeleteRepo(&logdb.DeleteRepoInput{RepoName: repoName})
	if err != nil {
		t.Error(err)
	}
}

func TestQueryLogWithHighlight(t *testing.T) {
	repoName := "test_sdk_repo_send_log_with_highlight"
	createInput := &logdb.CreateRepoInput{
		RepoName:  repoName,
		Region:    region,
		Schema:    defaultRepoSchema,
		Retention: "2d",
	}
	err := client.CreateRepo(createInput)
	if err != nil {
		t.Error(err)
	}

	sendLogInput := &logdb.SendLogInput{
		RepoName:       repoName,
		OmitInvalidLog: false,
		Logs: logdb.Logs{
			logdb.Log{
				"f1": "v11",
				"f2": 1.0,
				"f3": time.Now().UTC().Format(time.RFC3339),
				"f4": 1312,
			},
		},
	}
	sendOutput, err := client.SendLog(sendLogInput)
	if err != nil {
		t.Error(err)
	}
	if sendOutput.Success != 1 || sendOutput.Failed != 0 || sendOutput.Total != 1 {
		t.Errorf("send log failed, success: %d, failed: %d, total: %d", sendOutput.Success, sendOutput.Failed, sendOutput.Total)
	}

	time.Sleep(1 * time.Second)

	queryInput := &logdb.QueryLogInput{
		RepoName: repoName,
		Query:    "f1:v11",
		Sort:     "f2:desc",
		From:     0,
		Size:     100,
		Highlight: &logdb.Highlight{
			PreTags:  []string{"<em>"},
			PostTags: []string{"</em>"},
			Fields: map[string]interface{}{
				"f1": map[string]string{},
			},
			RequireFieldMatch: false,
			FragmentSize:      1,
		},
	}
	time.Sleep(2 * time.Minute)
	queryOut, err := client.QueryLog(queryInput)
	if err != nil {
		t.Error(err)
	}
	if queryOut.Total != 1 {
		t.Errorf("log count should be 1, but %d", queryOut.Total)
	}
	if queryOut.PartialSuccess != false {
		t.Error("query partialSuccess should be false, but true")
	}
	if len(queryOut.Data) != 1 {
		t.Errorf("log count should be 1, but %d", len(queryOut.Data))
	}
	if queryOut.Data[0]["highlight"] == "" {
		t.Errorf("result don't contain highlight")
	}

	err = client.DeleteRepo(&logdb.DeleteRepoInput{RepoName: repoName})
	if err != nil {
		t.Error(err)
	}
}

func TestPartialQuery(t *testing.T) {
	repoName := "repo_send_log"
	createInput := &logdb.CreateRepoInput{
		RepoName:  repoName,
		Region:    region,
		Schema:    defaultRepoSchema,
		Retention: "2d",
	}
	err := client.CreateRepo(createInput)
	if err != nil {
		t.Error(err)
	}

	startTime := time.Now().Unix() * 1000
	for i := 0; i < 5; i++ {
		sendLogInput := &logdb.SendLogInput{
			RepoName:       repoName,
			OmitInvalidLog: false,
			Logs: logdb.Logs{
				logdb.Log{
					"f1": "v11",
					"f2": 1.0,
					"f3": time.Now().UTC().Format(time.RFC3339),
					"f4": 1312,
				},
				logdb.Log{
					"f1": "v21",
					"f2": 1.2,
					"f3": time.Now().UTC().Format(time.RFC3339),
					"f4": 3082,
				},
				logdb.Log{
					"f1": "v31",
					"f2": 0.0,
					"f3": time.Now().UTC().Format(time.RFC3339),
					"f4": 1,
				},
				logdb.Log{
					"f1": "v41",
					"f2": 0.3,
					"f3": time.Now().UTC().Format(time.RFC3339),
					"f4": 12345671,
				},
			},
		}
		sendOutput, err := client.SendLog(sendLogInput)
		if err != nil {
			t.Error(err)
		}
		if sendOutput.Success != 4 || sendOutput.Failed != 0 || sendOutput.Total != 4 {
			t.Errorf("send log failed, success: %d, failed: %d, total: %d", sendOutput.Success, sendOutput.Failed, sendOutput.Total)
		}
	}
	endTime := time.Now().Unix() * 1000
	time.Sleep(3 * time.Minute)

	queryInput := &logdb.PartialQueryInput{
		RepoName:    repoName,
		StartTime:   startTime,
		EndTime:     endTime,
		QueryString: "f1:v11",
		Size:        1,
		Sort:        "f3",
		SearchType:  logdb.PartialQuerySearchTypeA,
	}
	queryInput.Highlight.PostTag = "@test@"
	queryInput.Highlight.PreTag = "@test/@"
	queryOut, err := client.PartialQuery(queryInput)
	if err != nil {
		t.Error(err)
	}
	bodystring, err := json.Marshal(queryOut)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(bodystring))
	for queryOut.PartialSuccess == true {
		queryOut, err = client.PartialQuery(queryInput)
		if err != nil {
			t.Error(err)
		}
		bodystring, err = json.Marshal(queryOut)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(string(bodystring))
	}
	if queryOut.Total != 5 {
		t.Errorf("log count should be 5, but %d", queryOut.Total)
	}
	if queryOut.PartialSuccess != false {
		t.Error("query partialSuccess should be false, but true")
	}
	if len(queryOut.Hits) != 1 {
		t.Errorf("log count should be 1, but %d", len(queryOut.Hits))
	}
	err = client.DeleteRepo(&logdb.DeleteRepoInput{RepoName: repoName})
	if err != nil {
		t.Error(err)
	}
}
