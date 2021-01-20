package main

import(
	"log"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	//"bytes"
	"net/http"
	//"time"
	"github.com/gorilla/websocket"
	"fmt"
)

//現在、切断された際にクライアントが削除されていない。
//

type User struct{
	ID int
	Name string
}

//あまり使わないほうが良いが、とりあえず動きを確認するためにグローバルで実装
//makeはnewとは違い初期化をする,slice,map,channelのみ
var clients = make(map[*websocket.Conn]bool)//クライアントの追加,削除用
var broadcast = make(chan []byte)//実際にメッセージのやり取りを行うためのチャネル

var upgrader = websocket.Upgrader{}//HTTP通信からwebsocketにアップグレードするメソッドを持つインスタンスの作成


func HandleMessages(){
	//ticker := time.NewTicker(((60 * time.Second)*9)/10)
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
			db, err := sql.Open("mysql", "root:@/sample_db")
				if err != nil {
					panic(err.Error())
				}
				defer db.Close()

				stmtInsert, err := db.Prepare("INSERT INTO users(name) VALUES(?)")
				if err != nil {
					panic(err.Error())
				}
				defer stmtInsert.Close()

				result, err := stmtInsert.Exec(msg)
				if err != nil {
					panic(err.Error())
				}

				lastInsertID, err := result.LastInsertId()
				if err != nil{
					panic(err.Error())
				}
				fmt.Println(lastInsertID)


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
	dbsel(ws)

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


func dbsel(ws *websocket.Conn){
	db,err := sql.Open("mysql","root@/sample_db")
	if err != nil{
		panic(err.Error())
	}
	defer db.Close()

	rows,err := db.Query("SELECT * FROM users")
	if err != nil{
		panic(err.Error())
	}

	defer rows.Close()

	for rows.Next(){
		var user User
		err := rows.Scan(&user.ID,&user.Name)
		if err != nil {
			panic(err.Error())
		}
		historymsg := []byte(user.Name)
		
		
		w, err := ws.NextWriter(websocket.TextMessage)
		if err != nil {
			log.Printf("error: %v", err)
		}else{
		
			w.Write(historymsg)
			if err := w.Close(); err != nil{
				return
			}
		}
	}
	return
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

	//http.ListenAndServe("localhost:8080",nil)

	//鯖を建てる。建たなかったときにエラーを出力
	err := http.ListenAndServe("172.16.80.55:8080",nil)
	if err != nil{
		log.Fatal("ListenAndServe:",err)
	}

}
