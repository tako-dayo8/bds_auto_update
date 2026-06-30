package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const STATE_FILE_PATH = "./state.json"
const SERVER_LIST_URL = "https://net-secondary.web.minecraft-services.net/api/v1.0/download/links"

type State struct {
	Version   string     `json:"version"`
	InstallAt *time.Time `json:"install_at"`
}

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
	fmt.Println("hello world")

	state, err := loadState()
	if err != nil {
		panic(err)
	}

	fmt.Println(state)

	list, err := getServerDownloadLinkList()
	if err != nil {
		panic(err)
	}

	fmt.Println(list)

	var linux Links
	for _, link := range list.Links {
		if link.DownloadType == "serverBedrockLinux" {
			linux = link
			break
		}
	}

	fmt.Println(linux)

	filename, err := getServer(linux.DownloadUrl)
	if err != nil {
		panic(err)
	}

	if err := unzip(*filename); err != nil {
		panic(err)
	}

	// now := time.Now()
	// temp := State{
	// 	Version:   "1.1.1.1.1",
	// 	InstallAt: &now,
	// }

	// if err := writeState(temp); err != nil {
	// 	panic(err)
	// }
}

func unzip(filename string) error {
	z, err := zip.OpenReader(filename)
	if err != nil {
		panic(err)
	}
	defer z.Close()

	ext := path.Ext(filename)
	dirname := strings.TrimSuffix(filename, ext)

	// 前回のディレクトリが残っていたっ場合削除
	_, err = os.Stat(dirname)
	if err := os.RemoveAll(dirname); err != nil {
		return err
	}

	// ファイル名のディレクトリを作成する
	if err := os.MkdirAll(dirname, os.ModeDir); err != nil {
		panic(err)
	}

	for _, f := range z.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := path.Join(dirname, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}

	return nil
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

func loadState() (*State, error) {
	data, err := os.ReadFile(STATE_FILE_PATH)
	if err != nil {
		defaultState := State{}
		data, err = json.Marshal(defaultState)
		if err != nil {
			return nil, err
		}
		if err = os.WriteFile(STATE_FILE_PATH, data, 0644); err != nil {
			return nil, err
		}
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

func writeState(state State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if err := os.WriteFile(STATE_FILE_PATH, []byte(data), 0644); err != nil {
		return err
	}

	return nil
}
