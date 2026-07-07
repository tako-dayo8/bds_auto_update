# Bedrock Server Auto Update

Minecraft Bedrock Edition サーバーを自動で最新バージョンに更新するツールです。
Ubuntu での動作を想定しています。

## 概要

公式 API からダウンロードリンクを取得し、新バージョンが存在する場合のみサーバーをダウンロード・展開します。
更新時は `config/` に保存したカスタム設定ファイルを自動で適用します。

## 構成

```
bedrock_server_auto_update/
├── setup                    # 初期セットアップ用バイナリ
├── updater                  # 更新処理用バイナリ
├── config/
│   ├── server.properties    # カスタムサーバー設定
│   └── permissions.json     # スクリプト権限設定
├── state.json               # 現在のバージョン・インストール日時
└── bedrock-server-x.y.z/    # 展開されたサーバー本体
```

## インストール

### 1. 専用ユーザーとディレクトリの作成

```sh
sudo useradd -r -m -d /opt/bedrock -s /usr/sbin/nologin minecraft

# useradd が作成するホームディレクトリは 750 などになり、
# 一般ユーザーが cd できないためパーミッションを変更する
sudo chmod 755 /opt/bedrock
```

### 2. バイナリのダウンロード

ビルド済みの Linux バイナリは [GitHub Releases](https://github.com/tako-dayo8/bds_auto_update/releases/latest) からダウンロードできます。

```sh
cd /opt/bedrock

# setup バイナリ
sudo -u minecraft curl -LO https://github.com/tako-dayo8/bds_auto_update/releases/latest/download/setup-linux-amd64
sudo -u minecraft chmod +x setup-linux-amd64

# updater バイナリ
sudo -u minecraft curl -LO https://github.com/tako-dayo8/bds_auto_update/releases/latest/download/updater-linux-amd64
sudo -u minecraft chmod +x updater-linux-amd64
```

### 3. 初回セットアップ

`setup` を実行すると以下が作成されます。

- `state.json` — バージョン管理ファイル
- `config/server.properties` — デフォルトのサーバー設定
- `config/permissions.json` — スクリプトモジュール権限設定

```sh
cd /opt/bedrock
sudo -u minecraft ./setup-linux-amd64
```

### 4. 初回のサーバー取得

```sh
cd /opt/bedrock
sudo -u minecraft ./updater-linux-amd64
```

`updater` は以下の処理を行います。

1. `state.json` から現在のバージョンを読み込む
2. 公式 API から最新のダウンロードリンクを取得
3. 最新バージョンでなければ Linux 用サーバー zip をダウンロード
4. zip を展開し、`config/` の設定ファイルを上書きコピー
5. `state.json` のバージョンとインストール日時を更新

すでに最新バージョンの場合はそのまま終了します。

## systemd の設定

サーバー本体を service として常駐させ、timer で毎日自動更新します。

### サーバー本体の service

`/etc/systemd/system/bedrock.service` を作成します。

```ini
[Unit]
Description=Minecraft Bedrock Dedicated Server
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=minecraft
Group=minecraft
# 展開ディレクトリ名はバージョンごとに変わるため、最新のディレクトリを探して起動する
ExecStart=/bin/bash -c 'cd "$(ls -d /opt/bedrock/bedrock-server-* | sort -V | tail -1)" && LD_LIBRARY_PATH=. exec ./bedrock_server'
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

有効化して起動します。

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now bedrock.service
```

状態確認・ログ確認は次の通りです。

```sh
systemctl status bedrock.service
journalctl -u bedrock.service -f
```

### 自動更新の service と timer

`/etc/systemd/system/bedrock-updater.service` を作成します（oneshot なので常駐しません）。

```ini
[Unit]
Description=Update Minecraft Bedrock Dedicated Server
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
User=minecraft
Group=minecraft
WorkingDirectory=/opt/bedrock
ExecStart=/opt/bedrock/updater-linux-amd64
# 更新後にサーバーを再起動して新バージョンを反映する
ExecStartPost=+/usr/bin/systemctl try-restart bedrock.service
```

`/etc/systemd/system/bedrock-updater.timer` を作成します。

```ini
[Unit]
Description=Daily update check for Minecraft Bedrock Dedicated Server

[Timer]
# 毎日 AM 4:00 に実行（プレイヤーの少ない時間帯を推奨）
OnCalendar=*-*-* 04:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

timer を有効化します（service 側は enable 不要です）。

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now bedrock-updater.timer
```

動作確認は次の通りです。

```sh
# 次回実行予定の確認
systemctl list-timers bedrock-updater.timer

# 手動で即時実行
sudo systemctl start bedrock-updater.service

# 更新ログの確認
journalctl -u bedrock-updater.service
```

> **Note**
> `updater` は「更新した」「すでに最新だった」のどちらでも正常終了するため、上記の設定では実行のたびにサーバーが再起動されます。
> 再起動を避けたい場合は `ExecStartPost` の行を削除し、更新があったときに手動で `sudo systemctl restart bedrock.service` を実行してください。

## 設定のカスタマイズ

`config/server.properties` を編集することでサーバーの動作を変更できます。
更新時にこのファイルが自動でサーバーディレクトリに適用されます。
編集後は再起動で反映されます。

```sh
sudo systemctl restart bedrock.service
```

## 動作環境

- Ubuntu 22.04 以降（x86_64）
- インターネット接続（公式 API へのアクセスが必要）
