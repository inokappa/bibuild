package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/kyokomi/emoji.v1"
	"gopkg.in/yaml.v2"
)

const (
	appVersion = "0.0.2"
)

var (
	argVersion    = flag.Bool("version", false, "バージョンを出力.")
	argProfile    = flag.String("profile", "", "Specify a Profile Name to the AWS Shared Credential.")
	argRole       = flag.String("role", "", "Role ARN を指定.")
	argRegion     = flag.String("region", "ap-northeast-1", "Specify an AWS Region.")
	argEndpoint   = flag.String("endpoint", "", "Specify an AWS API Endpoint URL.")
	argConfigFile = flag.String("config", "config.yml", "YAML ファイルのパスを指定.")
	argTarget     = flag.String("target", "default", "YAML ファイル内のターゲット名を指定.")
	argDirectory  = flag.String("dir", "", "ソースファイルのディレクトリを指定.")
	argBucket     = flag.String("bucket", "", "ソースファイルをアップロードする S3 バケットを指定.")
	argSourceFile = flag.String("source", "", "ソースファイル名を指定. (S3 バケットには [ソース名ファイル].zip という名前でアップロードされる)")
	argZip        = flag.Bool("zip", false, "ソースファイルの ZIP 圧縮を行う.")
	argPut        = flag.Bool("put", false, "ソースファイルを S3 にアップロードする.")
	argBuild      = flag.Bool("build", false, "CodeBuild プロジェクトのビルドを実行する.")
	argBuildStat  = flag.String("stat", "", "確認したい CodeBuild のビルド ID 指定.")
)

func readConfig(config_yaml string, target string) (config map[interface{}]interface{}) {
	yml, err := ioutil.ReadFile(config_yaml)
	if err != nil {
		fmt.Println(err)
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(yml), &m)
	if err != nil {
		fmt.Println(err)
	}

	if m[target] == nil {
		emoji.Printf(":bangbang: 指定したターゲットは定義されていません.\n")
		os.Exit(1)
	}

	config = make(map[interface{}]interface{})

	// 苦肉のエラー処理
	for _, v := range []string{"project_name",
		"source_bucket",
		"source_key",
		"directory"} {
		if m[target].(map[interface{}]interface{})[v] == nil {
			emoji.Printf(":warning: YAML ファイルのキー: %s が未定義です.\n", v)
			config[v] = ""
		} else {
			config[v] = m[target].(map[interface{}]interface{})[v].(string)
		}
	}

	return config
}

func zipIt(source, target string) (err error) {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		// Windows 環境で圧縮して Linux 環境で展開すると `\` がファイル名として展開される問題を解消
		header.Name = strings.Replace(header.Name, "\\", "/", -1)
		path = strings.Replace(path, "\\", "/", -1)

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		emoji.Println(":bangbang: ソースファイルの zip 圧縮に失敗しました.")
	} else {
		emoji.Println(":white_check_mark: ソースファイルの zip 圧縮に成功しました.")
	}

	return err
}

func main() {
	flag.Parse()

	if *argVersion {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	_, err := os.Stat(*argConfigFile)
	if err != nil {
		emoji.Println(":bangbang: YAML ファイルが存在していません. -config オプションで YAML ファイルを指定して下さい.")
		os.Exit(1)
	}

	// YAML ファイルから設定を読み込む
	appConfig := readConfig(*argConfigFile, *argTarget)
	var projectName string
	projectName = appConfig["project_name"].(string)

	// CodeBuild 用のソースファイルを指定
	var zipfile string
	if *argSourceFile != "" {
		zipfile = *argSourceFile + ".zip"
	} else {
		zipfile = "source.zip"
	}

	// CodeBuild 用の S3 バケットを指定
	var bucket string
	if *argBucket != "" {
		bucket = *argBucket
	} else {
		bucket = appConfig["source_bucket"].(string)
	}
	if bucket == "" {
		emoji.Println(":bangbang: CodePipeline 又は CodeBuild 用のソースバケットが指定されていません.")
		os.Exit(1)
	}

	// CodeBuild のビルドステータスを取得する
	if *argBuildStat != "" {
		sv := NewCodeBuildService(*argProfile, *argRegion, *argRole)
		err := sv.GetBuildStatus(*argBuildStat)
		if err != nil {
			fmt.Printf("\x1b[31;1mError : %s\x1b[0m\n", err)
			os.Exit(1)
		}
	}

	// ディレクトリを指定
	var directory string
	if *argDirectory != "" {
		directory = *argDirectory
	} else {
		directory = appConfig["directory"].(string)
	}
	_, err = os.Stat(directory)
	if err != nil {
		emoji.Printf(":bangbang: ソース用のディレクトリ %s が存在していません.\n", directory)
		os.Exit(1)
	}

	// -zip オプションを指定した場合
	if *argZip {
		// カレントディレクトリを取得
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// directory に移動
		os.Chdir(directory)
		err = zipIt("./", "../"+zipfile)
		if err != nil {
			fmt.Printf("\x1b[31;1mError : %s\x1b[0m\n", err)
			os.Exit(1)
		}
		// 既存のディレクトリに戻る
		os.Chdir(pwd)
	} else {
		emoji.Printf(":beer: ソースファイルの圧縮をスキップします.\n")
	}

	// -put オプションを指定した場合
	if *argPut {
		sv := NewS3Service(*argProfile, *argRegion, *argRole)
		err := sv.S3PutObject(zipfile, "application/zip", bucket, "", zipfile)
		if err != nil {
			fmt.Printf("\x1b[31;1mError : %s\x1b[0m\n", err)
			os.Exit(1)
		}
	} else {
		emoji.Printf(":beer: S3 へのアップロードをスキップします.\n")
	}

	if *argBuild {
		sv := NewCodeBuildService(*argProfile, *argRegion, *argRole)
		err := sv.BiBuild(projectName)
		if err != nil {
			fmt.Printf("\x1b[31;1mError : %s\x1b[0m\n", err)
			os.Exit(1)
		}
	} else {
		emoji.Printf(":beer: CodeBuild のビルドをスキップします.\n")
	}
}
