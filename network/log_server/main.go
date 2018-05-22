package main

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//
func main() {
	router := gin.Default()
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	db := session.DB("net_log")

	postReport(router, db, "/report")
	getBlocks(router, db, "/blocks")
	getTxs(router, db, "/txs")
	getNodes(router, db, "/nodes")

	router.Run(":30333")
}
type LogBlock struct {
	From string `json:"from"`
	Time string `json:"time"`
	SubType       string `json:"sub_type"`
	BlockHeadHash string `json:"block_head_hash"`// base64
	BlockNum      uint64 `json:"block_num"`
}
type LogTx struct {
	From string `json:"from"`
	Time string `json:"time"`
	SubType   string `json:"sub_type"`
	TxHash    string `json:"tx_hash"`
	Publisher string `json:"publisher"`
	Nonce     int64 `json:"nonce"`
}
type LogNode struct {
	From string `json:"from"`
	Time string `json:"time"`
	SubType string `json:"sub_type"`
	Log     string `json:"log"`
}
func postReport(router *gin.Engine, db *mgo.Database, path string)  {
	router.POST(path, func(c *gin.Context) {
		typ := c.PostFormArray("type")
		var err error
		switch typ[0] {
		case "Block":
			num, err := strconv.Atoi(c.PostForm("block_number"))
			if err != nil {
				fmt.Errorf("failed to conv string to int :%v", err)
				break
			}
			err = db.C("blocks").Insert(&LogBlock{
				From:c.PostForm("from"),
				Time:c.PostForm("time"),
				SubType:typ[1],
				BlockHeadHash:c.PostForm("block_head_hash"),
				BlockNum:uint64(num),
			})

		case "Tx":
			nonce, err := strconv.Atoi(c.PostForm("nonce"))
			if err != nil {
				fmt.Errorf("failed to conv string to int :%v", err)
				break
			}
			err = db.C("txs").Insert(&LogTx{
				From:c.PostForm("from"),
				Time:c.PostForm("time"),
				SubType:typ[1],
				TxHash:c.PostForm("hash"),
				Publisher:c.PostForm("publisher"),
				Nonce:int64(nonce),
			})
		case "Node":
			err = db.C("nodes").Insert(&LogNode{
				From:c.PostForm("from"),
				Time:c.PostForm("time"),
				SubType:typ[1],
				Log:c.PostForm("log"),
			})
		}
		c.JSON(200, gin.H{"err_msg": err})
	})
}

func getBlocks(router *gin.Engine, db *mgo.Database, path string) {
	router.GET(path, func(c *gin.Context) {
		filter := bson.M{}
		if c.Query("from") != "" {
			filter["from"] = c.Query("from")
		}
		if c.Query("time") != "" {
			filter["time"] = c.Query("time")
		}
		if c.Query("sub_type") != "" {
			filter["sub_type"] = c.Query("sub_type")
		}
		if c.Query("block_head_hash") != "" {
			filter["block_head_hash"] = c.Query("block_head_hash")
		}
		if c.Query("block_num") != "" {
			num, err := strconv.Atoi(c.Query("block_number"))
		if err != nil {
			fmt.Errorf("failed to conv string to int :%v", err)
		}
			filter["block_num"] = num
		}

		var result []LogBlock
		err := db.C("blocks").Find(filter).All(&result)
		if err !=nil {
			fmt.Errorf("%v", err)
			c.JSON(200, err)
		}

		c.JSON(200, result)
	})
}

func getTxs(router *gin.Engine, db *mgo.Database, path string) {
//todo
}

func getNodes(router *gin.Engine, db *mgo.Database, path string) {

//	todo
}
