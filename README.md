# bibuild [![CircleCI](https://circleci.com/gh/inokappa/bibuild.svg?style=svg)](https://circleci.com/gh/inokappa/bibuild)

## これなに

* https://inokara.hateblo.jp/entry/2018/12/22/100850

## Usage

```sh
Usage of ./pkg/bibuild_darwin_amd64:
  -bucket string
        ソースファイルをアップロードする S3 バケットを指定.
  -build
        CodeBuild プロジェクトのビルドを実行する.
  -config string
        YAML ファイルのパスを指定. (default "config.yml")
  -dir string
        ソースファイルのディレクトリを指定.
  -endpoint string
        Specify an AWS API Endpoint URL.
  -profile string
        Specify a Profile Name to the AWS Shared Credential.
  -put
        ソースファイルを S3 にアップロードする.
  -region string
        Specify an AWS Region. (default "ap-northeast-1")
  -role string
        Role ARN を指定.
  -source string
        ソースファイル名を指定. (S3 バケットには [ソース名ファイル].zip という名前でアップロードされる)
  -stat string
        確認したい CodeBuild のビルド ID 指定.
  -target string
        YAML ファイル内のターゲット名を指定. (default "default")
  -version
        バージョンを出力.
  -zip
        ソースファイルの ZIP 圧縮を行う.
```
