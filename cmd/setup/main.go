package main

import (
	"bedrock_server_auto_update/cmd/internal/state"
	"fmt"
	"log"
	"os"
	"path"
	"time"
)

const CONFIG_DIR = "config"
const SERVER_PROPERTIES = "server.properties"
const PERMISSIONS_JSON = "permissions.json"

const PERMISSIONS_JSON_DEFAULT = `{
  "allowed_modules": [
    "@minecraft/server",
    "@minecraft/server-ui",
    "@minecraft/server-net",
    "@minecraft/server-admin",
    "@minecraft/server-editor",
    "@minecraft/server-graphics",
    "@minecraft/server-gametest",
    "@minecraft/diagnostics"
  ]
}`

const SERVER_PROPERTIES_DEFAULT = `server-name=Debug Server
# サーバー名として使用されます
# 使用可能な値: セミコロン記号を含まない任意の文字列

gamemode=creative
# 新規プレイヤーのゲームモードを設定します
# 使用可能な値: "survival"（サバイバル）, "creative"（クリエイティブ）, "adventure"（アドベンチャー）

force-gamemode=false
# force-gamemode=false（またはserver.propertiesでforce-gamemodeが未定義の場合）
# ワールド作成時にサーバーが保存したゲームモード以外の値を
# クライアントに送信しないようにします（ワールド作成後にserver.propertiesで設定変更しても）
#
# force-gamemode=true にすると、ワールド作成後にserver.propertiesで
# 設定した値をクライアントに強制的に送信します

difficulty=easy
# ワールドの難易度を設定します
# 使用可能な値: "peaceful"（ピースフル）, "easy"（イージー）, "normal"（ノーマル）, "hard"（ハード）

allow-cheats=true
# trueにするとコマンドなどのチートが使用できます
# 使用可能な値: "true" または "false"

max-players=10
# サーバーでプレイできる最大プレイヤー数
# 使用可能な値: 正の整数

online-mode=true
# trueにすると接続するすべてのプレイヤーがXbox Liveで認証される必要があります
# リモート（LAN以外）サーバーに接続するクライアントはこの設定に関係なく常にXbox Live認証が必要です
# インターネットからの接続を受け付ける場合はonline-modeを有効にすることを強く推奨します
# 使用可能な値: "true" または "false"

allow-list=false
# trueにすると接続するすべてのプレイヤーが別のallowlist.jsonファイルに記載されている必要があります
# 使用可能な値: "true" または "false"

server-port=19132
# サーバーがリッスンするIPv4ポート
# 使用可能な値: [1, 65535] の範囲の整数

server-portv6=19133
# サーバーがリッスンするIPv6ポート。transport=nethernet の場合は無視され、
# server-port でデュアルスタックソケットが開かれます
# 使用可能な値: [1, 65535] の範囲の整数

# server-ip=
# 空白のままにするとすべてのインターフェースにバインドします。transport=raknet の場合は無視されます
# 使用可能な値: IPv4またはIPv6リテラル、または空白

# server-udp-ports=
# UDPクライアントトランスポートポートを設定します。カンマ区切りで複数指定でき、
# 設定を追加するために同じプロパティを複数行に記述することもできます。
# transport=raknet の場合は無視されます
# 使用可能な値:
#   * 'internal' または 'start-end' — 内部ポート（または範囲）でUDPローカルポートの割り当て範囲を制限します
#   * '[ip:]external[-external]:internal[-internal] — クライアントが到達可能な外部マッピングを追加公開します
#     両側の範囲は同じ長さである必要があります
#
# 例:
#   server-udp-ports=49152-49200 （内部ポート範囲のみ）
#   server-udp-ports=19132:32000 （外部ポート19132を内部ポート32000にマッピング）
#   server-udp-ports=203.0.113.10:19132-19140:32000-32008
#   server-udp-ports=[2001:db8::1]:19132:32000,32000

transport=raknet
# サーバーが使用するトランスポートプロトコル
# 使用可能な値: "raknet" または "nethernet"

enable-lan-visibility=true
# LAN上のサーバーを探しているクライアントへの応答を有効にします
# transport=raknet の場合、server-port と server-portv6 が非デフォルト値でも
# デフォルトポート（19132, 19133）にバインドされます。
# LAN検索が不要な場合や、同一ホストで複数サーバーを実行してポートが競合する場合はオフにしてください
# 使用可能な値: "true" または "false"

view-distance=32
# チャンク数で表した最大視野距離
# 使用可能な値: 5以上の正の整数

tick-distance=4
# プレイヤーからこのチャンク数の範囲内のワールドがティック処理されます
# 使用可能な値: [4, 12] の範囲の整数

player-idle-timeout=0
# プレイヤーがこの分数アイドル状態になるとキックされます。0に設定するとプレイヤーは無制限にアイドルできます
# 使用可能な値: 0以上の整数

# allow-player-joining=true
# falseにするとスクリプト（AsyncPlayerJoinBeforeEvent）で明示的に許可されない限りプレイヤーはサーバーに参加できません
# 使用可能な値: true, false
# デフォルト: true

max-threads=8
# サーバーが使用しようとするスレッドの最大数。0または未設定の場合は可能な限り多く使用します
# 使用可能な値: 正の整数

level-name=Bedrock level
# 使用可能な値: セミコロンやファイル名に使用できない記号を含まない任意の文字列: /\n\r\t\f?*\\<>|\":

level-seed=
# ワールドのランダム生成に使用します
# 使用可能な値: 任意の文字列

default-player-permission-level=operator
# 初めて参加する新規プレイヤーの権限レベル
# 使用可能な値: "visitor"（ビジター）, "member"（メンバー）, "operator"（オペレーター）

texturepack-required=true
# クライアントに現在のワールドのテクスチャパックの使用を強制します
# 使用可能な値: "true" または "false"

content-log-file-enabled=true
# コンテンツエラーのファイルへのログ記録を有効にします
# 使用可能な値: "true" または "false"

content-log-console-output-enabled=true
# コンテンツエラーの標準出力へのログ記録を有効にします
# 使用可能な値: "true" または "false"

content-log-level=verbose
# コンテンツログの最低レベルを設定します（errorが最高レベル）
# 使用可能な値: "error", "warning", "info", "verbose"

compression-threshold=1
# 圧縮する生のネットワークペイロードの最小サイズ
# 使用可能な値: 0-65535

compression-algorithm=zlib
# ネットワーク通信に使用する圧縮アルゴリズム
# 使用可能な値: "zlib", "snappy"

server-authoritative-movement-strict=false
# trueにするとプレイヤー位置の検証が厳格になり、クライアント情報の許容範囲が狭くなります
# クライアントへの位置補正が増えます。高レイテンシ時に動くブロック周辺のプレイヤーに影響します

server-authoritative-dismount-strict=false
# trueにするとプレイヤーの降車位置の検証が厳格になります
# 高レイテンシ時に降車位置の補正がクライアントに送られるようになります

server-authoritative-entity-interactions-strict=false
# trueにするとエンティティのインタラクションの検証が厳格になります
# 高レイテンシ時のプレイヤー同士のインタラクションに影響します

player-position-acceptance-threshold=0.5
# クライアントとサーバーのプレイヤー位置の許容誤差です。
# ダメージによるノックバックやピストンに押される場合など、サーバーとクライアントで動作開始のタイミングが
# 異なる場合に、チートでないプレイヤーへの補正が頻繁に送られないようにします
# 値が大きいほどサーバーの許容範囲が広がります。1.0を超えるとチートを許容するリスクが高まります

player-movement-action-direction-threshold=0.85
# プレイヤーの攻撃方向と視線方向がどれだけ異なってもよいかを設定します
# 使用可能な値: [0, 1] の範囲の値
# 1 は視線方向と攻撃方向が完全に一致する必要があることを意味し、
# 0 は最大90度まで異なってもよいことを意味します

server-authoritative-block-breaking-pick-range-scalar=1.5
# trueにするとサーバーがクライアントと同期してブロック採掘処理を計算し、
# クライアントがブロックを破壊できるかどうかを検証します

chat-restriction=None
# 使用可能な値: "None", "Dropped", "Disabled"
# サーバーに参加する各プレイヤーのチャット制限レベルを設定します
# "None" はデフォルトで通常のフリーチャットを意味します
# "Dropped" はチャットメッセージが破棄されどのクライアントにも送信されません。プレイヤーには機能が無効と表示されます
# "Disabled" はオペレーター以外のプレイヤーにはチャットUIが表示されません

disable-player-interaction=false
# trueにするとサーバーはクライアントにワールドとのインタラクション時に
# 他のプレイヤーを無視するよう通知します。サーバー権威ではありません

client-side-chunk-generation-enabled=true
# trueにするとサーバーはクライアントにプレイヤーのインタラクション距離外の
# 視覚的なチャンクを生成できることを通知します

block-network-ids-are-hashes=true
# trueにするとサーバーは0から始まるIDの代わりにハッシュ化されたブロックネットワークIDを送信します
# これらのIDは安定しており他のブロック変更に関係なく変わりません

disable-persona=false
# 内部使用のみ

disable-custom-skins=false
# trueにするとMinecraftストアやゲーム内アセット以外でカスタマイズされた
# プレイヤーのカスタムスキンを無効にします。不適切なカスタムスキンの対策に使用します

server-build-radius-ratio=Disabled
# 使用可能な値: "Disabled" または [0.0, 1.0] の範囲の値
# "Disabled" の場合、サーバーはプレイヤーの視野のどれだけを生成するかを動的に計算し、残りをクライアントに任せます
# それ以外の場合、クライアントのハードウェア性能に関係なくサーバーが生成する割合を上書き指定します
# client-side-chunk-generation-enabled が有効な場合のみ有効です

allow-outbound-script-debugging=true
# スクリプトデバッガの 'connect' コマンドと script-debugger-auto-attach=connect モードを許可します

allow-inbound-script-debugging=true
# スクリプトデバッガの 'listen' コマンドと script-debugger-auto-attach=listen モードを許可します

#force-inbound-debug-port=19144
# インバウンド（listen）デバッガポートを固定します。未設定の場合はデフォルトの19144が使用されます
# script-debugger-auto-attach=listen モードを使用する場合に必要です

script-debugger-auto-attach=disabled
# レベルロード時にスクリプトデバッガへの自動接続を試みます。インバウンドポートまたは接続先アドレスの設定と
# インバウンドまたはアウトバウンド接続の有効化が必要です
# "disabled": 自動接続しません
# "connect": サーバーが指定ポートでlistenモードのデバッガへの接続を試みます
# "listen": サーバーがconnectモードのデバッガからのインバウンド接続をlistenします

#script-debugger-auto-attach-connect-address=localhost:19144
# 自動接続モードが 'connect' の場合に使用するアドレス（host:port形式）
# script-debugger-auto-attach=connect モードに必要です

#script-debugger-auto-attach-timeout=0
# ワールドロード時にデバッガの接続を待機する時間

#script-debugger-passcode=
# VSCodeが接続時にユーザーにパスコードを要求します

#script-watchdog-enable=true
# ウォッチドッグを有効にします（デフォルト = true）

#script-watchdog-enable-exception-handling=true
# events.beforeWatchdogTerminate イベントによるウォッチドッグの例外処理を有効にします（デフォルト = true）

#script-watchdog-enable-shutdown=true
# ウォッチドッグの未処理例外が発生した場合のサーバーシャットダウンを有効にします（デフォルト = true）

#script-watchdog-hang-exception=true
# ハングが発生した場合にクリティカル例外をスローしスクリプト実行を中断します（デフォルト = true）

#script-watchdog-hang-threshold=10000
# 単一ティックのハングに対するウォッチドッグのしきい値を設定します（デフォルト = 10000 ms）

#script-watchdog-spike-threshold=100
# 単一ティックのスパイクに対するウォッチドッグのしきい値を設定します
# プロパティが未設定の場合は警告が無効になります

#script-watchdog-slow-threshold=10
# 複数ティックにわたる低速スクリプトに対するウォッチドッグのしきい値を設定します
# プロパティが未設定の場合は警告が無効になります

#script-watchdog-memory-warning=100
# スクリプトの合計メモリ使用量が指定したしきい値（MB単位）を超えた場合にコンテンツログ警告を出します
# 0に設定すると警告が無効になります（デフォルト = 100、最大 = 2000）

#script-watchdog-memory-limit=250
# スクリプトの合計メモリ使用量が指定したしきい値（MB単位）を超えた場合にワールドを保存してシャットダウンします
# 0に設定すると制限が無効になります（デフォルト = 250、最大 = 2000）

#diagnostics-capture-auto-start=false
# レベルロード時に診断キャプチャセッションを開始します（デフォルト = false）

#diagnostics-capture-max-files=5
# 循環前に保持する診断キャプチャファイルの最大数（デフォルト = 5）

#diagnostics-capture-max-file-size=2097152
# 新しいファイルに切り替えるまでの現在の診断キャプチャファイルの最大サイズ（バイト単位）（デフォルト = 2097152、2MB）

#disable-client-vibrant-visuals=true
# trueにするとサーバーはクライアントにVibrant Visualsの代わりに次に良いグラフィック設定を使用するよう通知します

#sentry-rate-limit-window=60
# スクリプトエラーをSentryに送信する際の制限を適用する秒数
# デフォルト: 60
# 使用可能な値: 正の整数、または0で制限を無効化

#sentry-max-events-per-window=10
# 指定した時間ウィンドウ内で許可されるイベントの最大数
# デフォルト: 10
# 使用可能な値: 正の整数、または0でSentryへのイベント送信を無効化

#enable-profiler=true
# trueにするとパフォーマンス分析用のプロファイラサポートを有効にします

#enable-editor-network-metrics=true
# trueにするとデバッガでライブ診断を表示するためのネットワークメトリクス収集を有効にします

#convert-world-to-editor-project=false
# trueかつサーバーがコマンドライン引数 Editor=true で起動された場合、
# ロード時に既存のバニラワールドをエディタープロジェクトにアップグレードします。
# Editor=true なしでは効果がありません
# デフォルト: false`

func main() {
	now := time.Now()

	// スタッツファイルの作成
	base := state.State{
		Version:   "1.1.1.1",
		InstallAt: &now,
	}

	if err := state.WriteState(base); err != nil {
		log.Fatal(err)
	}
	fmt.Println(state.STATE_FILE_PATH + "を作成しました")

	// config ディレクトリ作成
	if err := os.Mkdir(CONFIG_DIR, 0755); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(CONFIG_DIR + "ディレクトリを作成しました")
	}

	//config/ permissions.json, server.propertiesを作成
	if err := os.WriteFile(path.Join(CONFIG_DIR, PERMISSIONS_JSON), []byte(PERMISSIONS_JSON_DEFAULT), 0666); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(path.Join(CONFIG_DIR, SERVER_PROPERTIES), []byte(SERVER_PROPERTIES_DEFAULT), 0666); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s,%sを作成しました\n", SERVER_PROPERTIES, PERMISSIONS_JSON)
}
