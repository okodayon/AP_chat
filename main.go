package main

import(
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)


//makeはnewとは違い初期化をする,slice,map,channelのみ
var clients = make(map[*websocket.Conn]bool)//クライアントの追加,削除用
var broadcast = make(chan []byte)//実際にメッセージのやり取りを行うためのチャネル

var upgrader = websocket.Upgrader{}//HTTP通信からwebsocketにアップグレードするメソッドを持つインスタンスの作成


func HandleMessages(){
	for {
	select{
		case msg := <- broadcast:
			for client := range clients {
				//client.SetWriteDeadline(time.Now().Add(10*time.Second))
				w, err := client.NextWriter(websocket.TextMessage)
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(clients, client)
				}else{
					w.Write(msg)
					if err := w.Close(); err != nil{
						return
					}
				}
			}
		}
	}
}


//websocketを確立させて、メッセージを遅れる状態にする
func HandleConnection(w http.ResponseWriter, r *http.Request){
	log.Println(r.URL)
	//http接続からハンドシェイクをしwebsocketにupgradeしている
	//nil値は未定義ではないゼロ値
	ws, _ := upgrader.Upgrade(w,r,nil)


	//関数が戻ってきたとき(return)にdeferを使いwebsocketを閉じる
	defer func() {
		ws.Close()
		delete(clients,ws)
	}()
	//clients(メッセージの送信先)の追加
	clients[ws] = true

	for{
		
		_, message, err := ws.ReadMessage()
		if err != nil{
			log.Printf("readerr:%v",err)
			//panicにも対応できるということなのでdeferに変更してみた↑
			/*
			ws.Close()
			delete(clients,ws)
			*/
			//break
			return
		}

		broadcast <- message
	}
}


func main(){
	/*この書き方だと外部のCSS,jsファイルが読み取れなかった
	http.HandleFunc("/",func(w http.ResponseWriter, r *http.Request){
		http.ServeFile(w,r,"home.html")
	})
	*/

	//default接続時のディレクトリを設定。今回の場合index.htmlが最初に読み込まれる
	fileServer := http.FileServer(http.Dir("./public"))

	http.Handle("/",fileServer)
	//websocket接続時の動作を設定
	http.HandleFunc("/ws",HandleConnection)

	go HandleMessages()

	//localhostで立ち上げる場合
	//http.ListenAndServe("localhost:8080",nil)

	//鯖を建てる。建たなかったときにエラーを出力
	//ローカルネットワークでテストするとき用
	err := http.ListenAndServe("172.16.80.55:8080",nil)
	if err != nil{
		log.Fatal("ListenAndServe:",err)
	}

}
