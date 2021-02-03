# AP_chat
mysqlの構造  
+-------+-----------+------+-----+-------------------+-----------------------------------------------+  
| Field | Type      | Null | Key | Default           | Extra                                         |  
+-------+-----------+------+-----+-------------------+-----------------------------------------------+  
| id    | int       | NO   | PRI | NULL              | auto_increment                                |  
| CHAT  | longblob  | NO   |     | NULL              |                                               |  
| time  | timestamp | YES  |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED on update CURRENT_TIMESTAMP |  
+-------+-----------+------+-----+-------------------+-----------------------------------------------+  

go get github.com/gorilla/websocket  
go run main.go
