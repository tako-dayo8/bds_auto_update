package main

import (
	"archive/zip"
	"bedrock_server_auto_update/cmd/internal/state"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

const SERVER_PROPERTIES = "server.properties"
const PERMISSIONS_JSON = "permissions.json"
const SERVER_LIST_URL = "https://net-secondary.web.minecraft-services.net/api/v1.0/download/links"

type FetchResult struct {
	Result ServerDownloadLinkList `json:"result"`
}

type ServerDownloadLinkList struct {
	Links []Links `json:"links"`
}

type Links struct {
	DownloadType string `json:"downloadType"`
	DownloadUrl  string `json:"downloadUrl"`
}

func main() {
	s, err := state.LoadState()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("state.json: ", s)

	list, err := getServerDownloadLinkList()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("downloadLinkList: ", list)

	var linux Links
	for _, link := range list.Links {
		if link.DownloadType == "serverBedrockLinux" {
			linux = link
			break
		}
	}

	fmt.Println("linuxLink", linux)

	// 最新バージョンであった場合終了
	if strings.Contains(linux.DownloadUrl, s.Version) {
		fmt.Println("最新バージョンです")
		os.Exit(0)
	}

	filename, err := getServer(linux.DownloadUrl)
	if err != nil {
		log.Fatal(err)
	}

	dirname, err := unzip(*filename)
	if err != nil {
		log.Fatal(err)
	}

	if err := swapConfigFile(dirname); err != nil {
		log.Fatal(err)
	}

	re := regexp.MustCompile(`[\d]+(?:\.[\d]+)+`)
	version := re.FindString(dirname)

	fmt.Println("get version: ", version)

	now := time.Now()
	newsate := state.State{
		Version:   version,
		InstallAt: &now,
	}

	if err := state.WriteState(newsate); err != nil {
		log.Fatal(err)
	}
}

func swapConfigFile(dirname string) error {
	// ダウンロードしてきた permissions.json, server.propertiesを削除
	if err := os.Remove(path.Join(dirname, SERVER_PROPERTIES)); err != nil {
		return err
	}
	if err := os.Remove(path.Join(dirname, PERMISSIONS_JSON)); err != nil {
		return err
	}

	// config配下の permissions.json, server.properties をコピー
	for _, file := range []string{SERVER_PROPERTIES, PERMISSIONS_JSON} {
		src, err := os.Open(path.Join("config", file))
		if err != nil {
			return err
		}

		dst, err := os.Create(path.Join(dirname, file))
		if err != nil {
			return err
		}

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}

		// ループ内のため即時クローズ
		src.Close()
		dst.Close()
	}

	return nil
}

// unzip は引数に指定されたzipファイルのpathから同じ名前のディレクトリを作成し、解凍します。
//
// 引数：zip ファイルの path
//
// 返り値: 解凍したディレクトリ（dirname） 失敗した場合 error を返します
func unzip(filename string) (string, error) {
	z, err := zip.OpenReader(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer z.Close()

	ext := path.Ext(filename)
	dirname := strings.TrimSuffix(filename, ext)

	// 前回のディレクトリが残っていたっ場合削除
	_, err = os.Stat(dirname)
	if err := os.RemoveAll(dirname); err != nil {
		return dirname, err
	}

	// ファイル名のディレクトリを作成する
	if err := os.MkdirAll(dirname, os.ModeDir); err != nil {
		log.Fatal(err)
	}

	for _, f := range z.File {
		rc, err := f.Open()
		if err != nil {
			return dirname, err
		}
		defer rc.Close()

		path := path.Join(dirname, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return dirname, err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return dirname, err
			}
		}
	}

	return dirname, nil
}

func getServer(url string) (filename *string, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// ヘッダを設定
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Accept", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	_, fn := path.Split(url)

	// 前回のzipが残っていた場合削除
	_, err = os.Stat(fn)
	if os.IsExist(err) {
		if err := os.Remove(fn); err != nil {
			return nil, err
		}
	}

	file, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	file.Write(body)

	return &fn, nil
}

func getServerDownloadLinkList() (*ServerDownloadLinkList, error) {
	resp, err := http.Get(SERVER_LIST_URL)
	if err != nil {
		return nil, err
	}

	// 処理が終了したら必ずレスポンスのボディを閉じる
	defer resp.Body.Close()

	// ステータスコードの確認
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("異常なステータスコード: %d", resp.StatusCode)
	}

	// レスポンスボディのデータをすべて読み込む
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// fmt.Println(string(body))

	var result FetchResult
	var list ServerDownloadLinkList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	} else {
		list = result.Result
	}

	// 返す前にディレイを入れる (負荷によるブロック避け)
	// time.Sleep(1 * time.Minute)

	return &list, nil
}
