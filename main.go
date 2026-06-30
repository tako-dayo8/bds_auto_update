package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

	// now := time.Now()
	// temp := State{
	// 	Version:   "1.1.1.1.1",
	// 	InstallAt: &now,
	// }

	// if err := writeState(temp); err != nil {
	// 	panic(err)
	// }
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
	time.Sleep(1 * time.Minute)

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
