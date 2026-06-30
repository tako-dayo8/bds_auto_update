# Bedrock Server Auto Update

Minecraft Bedrock Edition サーバーを自動で最新バージョンに更新するツールです。

## 概要

公式 API からダウンロードリンクを取得し、新バージョンが存在する場合のみサーバーをダウンロード・展開します。
更新時は `config/` に保存したカスタム設定ファイルを自動で適用します。

## 構成

```
bedrock_server_auto_update/
├── cmd/
│   ├── updater/main.go      # メイン更新処理
│   └── setup/main.go        # 初期セットアップ
├── cmd/internal/
│   └── state/state.go       # バージョン状態管理
├── config/
│   ├── server.properties    # カスタムサーバー設定
│   └── permissions.json     # スクリプト権限設定
├── state.json               # 現在のバージョン・インストール日時
└── go.mod
```

## 使い方

### 1. 初回セットアップ

`setup.exe` を実行すると以下が作成されます。

- `state.json` — バージョン管理ファイル
- `config/server.properties` — デフォルトのサーバー設定（日本語コメント付き）
- `config/permissions.json` — スクリプトモジュール権限設定

```sh
./setup.exe
```

### 2. サーバーの更新

`updater.exe` を実行すると以下の処理が行われます。

1. `state.json` から現在のバージョンを読み込む
2. 公式 API から最新のダウンロードリンクを取得
3. 最新バージョンでなければ Linux 用サーバー zip をダウンロード
4. zip を展開し、`config/` の設定ファイルを上書きコピー
5. `state.json` のバージョンとインストール日時を更新

```sh
./updater.exe
```

最新バージョンの場合はそのまま終了します。

### 3. 設定のカスタマイズ

`config/server.properties` を編集することでサーバーの動作を変更できます。
更新時にこのファイルが自動でサーバーディレクトリに適用されます。

## ビルド

```sh
# Windows 向け
go build -o setup.exe ./cmd/setup
go build -o updater.exe ./cmd/updater

# Linux 向け
GOOS=linux GOARCH=amd64 go build -o setup ./cmd/setup
GOOS=linux GOARCH=amd64 go build -o updater ./cmd/updater
```

## 動作環境

- Go 1.26 以上
- インターネット接続（公式 API へのアクセスが必要）
