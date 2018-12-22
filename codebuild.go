package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/kyokomi/emoji.v1"
)

type codeBuildService struct {
	*codebuild.CodeBuild
}

func NewCodeBuildService(profile string, region string, role string) *codeBuildService {
	var config aws.Config
	if profile != "" && role == "" {
		creds := credentials.NewSharedCredentials("", profile)
		config = aws.Config{Region: aws.String(region),
			Credentials: creds,
			Endpoint:    aws.String(*argEndpoint)}
	} else if profile == "" && role != "" {
		sess := session.Must(session.NewSession())
		creds := stscreds.NewCredentials(sess, role)
		config = aws.Config{Region: aws.String(region),
			Credentials: creds,
			Endpoint:    aws.String(*argEndpoint)}
	} else if profile != "" && role != "" {
		sess := session.Must(session.NewSessionWithOptions(session.Options{Profile: profile}))
		assumeRoler := sts.New(sess)
		creds := stscreds.NewCredentialsWithClient(assumeRoler, role)
		config = aws.Config{Region: aws.String(region),
			Credentials: creds,
			Endpoint:    aws.String(*argEndpoint)}
	} else {
		config = aws.Config{Region: aws.String(region),
			Endpoint: aws.String(*argEndpoint)}
	}
	sess := session.New(&config)
	svc := codebuild.New(sess)
	sv := &codeBuildService{
		svc,
	}
	return sv
}

func (sv *codeBuildService) BiBuild(projectName string) (err error) {
	param := &codebuild.StartBuildInput{
		ProjectName: aws.String(projectName),
	}
	b, err := sv.StartBuild(param)
	if err != nil {
		emoji.Printf(":bangbang: ビルドの開始が失敗しました. Project Name: %s\n", projectName)
	} else {
		emoji.Printf(":white_check_mark: ビルドを開始しました. Project Name: %s, Build ID: %s\n", projectName, *b.Build.Id)
	}

	return err
}

func (sv *codeBuildService) GetBuildStatus(buildID string) (err error) {
	param := &codebuild.BatchGetBuildsInput{
		Ids: []*string{
			aws.String(buildID),
		},
	}
	res, err := sv.BatchGetBuilds(param)
	if err != nil {
		emoji.Printf(":bangbang: ビルドの状態取得に失敗しました. Build ID: %s\n", buildID)
	} else {
		// emoji.Printf(":white_check_mark: ビルドの状態を取得しました. Build ID: %s\n", buildID)
		var buildStatus string
		Phases := [][]string{}
		for _, r := range res.Builds {
			buildStatus = *r.BuildStatus
			for _, p := range r.Phases {
				var endTime string
				if p.EndTime == nil {
					endTime = "N/A"
				} else {
					endTime = convertDate(*p.EndTime)
				}
				var status string
				if p.PhaseStatus == nil {
					status = "N/A"
				} else {
					status = writeBuildStatus(*p.PhaseStatus)
				}
				startTime := convertDate(*p.StartTime)
				Phase := []string{
					*p.PhaseType,
					status,
					startTime,
					endTime,
				}
				Phases = append(Phases, Phase)
			}
		}
		fmt.Printf("Build ID: %s\n", buildID)
		fmt.Printf("Build Status: %s\n", buildStatus)
		outputTbl(Phases)
		os.Exit(0)
	}

	return err
}

func convertDate(d time.Time) (convertedDate string) {
	const layout = "2006-01-02 15:04:05"
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	convertedDate = d.In(jst).Format(layout)

	return convertedDate
}

func writeBuildStatus(status string) (st string) {
	if status == "SUCCEEDED" {
		st = emoji.Sprint(":white_check_mark: " + status)
	} else {
		st = emoji.Sprint(":bangbang: " + status)
	}
	return st
}

func outputTbl(data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"名前", "ステータス", "開始時刻", "終了時刻"})
	for _, value := range data {
		table.Append(value)
	}
	table.Render()
}
