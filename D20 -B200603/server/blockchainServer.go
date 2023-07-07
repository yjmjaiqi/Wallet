package main

import (
	"encoding/json"
	"io"
	"jhblockchain/block"
	"jhblockchain/utils"
	"jhblockchain/wallet"
	"log"
	"net/http"
	"strconv"
	"os"
	"github.com/fatih/color"
)

var cache map[string]*block.Blockchain = make(map[string]*block.Blockchain)

type BlockchainServer struct {
	port uint16
}

func NewBlockchainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{port}
}

func (bcs *BlockchainServer) Port() uint16 {
	return bcs.port
}
// 从文件读取钱包信息
func LoadWallet() (*wallet.Wallet, error){
	file, err := os.Open("wallet.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	var wallet *wallet.Wallet
	if err := dec.Decode(&wallet); err != nil {
		return nil, err
	}

	return wallet, nil
}
func (bcs *BlockchainServer) GetBlockchain() *block.Blockchain {
	bc, ok := cache["blockchain"]
	// 通过判断 cache["blockchain"]有没有值，来确保不重复执行NewBlockchain()
	if !ok {
		minersWallet,err := LoadWallet()
		if err != nil{
			minersWallet = wallet.NewWallet()
			// 写入挖矿奖励地址
			file, err := os.OpenFile("wallet.txt", os.O_CREATE | os.O_APPEND | os.O_RDWR,0744)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
	
			enc := json.NewEncoder(file)
			err = enc.Encode(minersWallet)
			if err != nil{
				return nil
			}
		}
		// NewBlockchain与以前的方法不一样,增加了地址和端口2个参数,是为了区别不同的节点
		bc = block.NewBlockchain(minersWallet.BlockchainAddress(), bcs.Port())
		cache["blockchain"] = bc
		color.Magenta("===矿工帐号信息====\n")
		color.Magenta("矿工private_key\n %v\n", minersWallet.PrivateKeyStr())
		color.Magenta("矿工publick_key\n %v\n", minersWallet.PublicKeyStr())
		color.Magenta("矿工blockchain_address\n %s\n", minersWallet.BlockchainAddress())
		color.Magenta("===============\n")
	}
	return bc
}

func (bcs *BlockchainServer) Transactions(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		{
			// Get:显示交易池的内容，Mine成功后清空交易池
			w.Header().Add("Content-Type", "application/json")
			bc := bcs.GetBlockchain()

			transactions := bc.TransactionPool()
			m, _ := json.Marshal(struct {
				Transactions []*block.Transaction `json:"transactions"`
				Length       int                  `json:"length"`
			}{
				Transactions: transactions,
				Length:       len(transactions),
			})
			io.WriteString(w, string(m[:]))
		}
	case http.MethodPost:
		{
			log.Printf("\n\n\n")
			log.Println("接受到wallet发送的交易")
			decoder := json.NewDecoder(req.Body)
			var t block.TransactionRequest
			err := decoder.Decode(&t)
			if err != nil {
				log.Printf("ERROR: %v", err)
				io.WriteString(w, string(utils.JsonStatus("Decode Transaction失败")))
				return
			}

			log.Println("发送人公钥SenderPublicKey:", *t.SenderPublicKey)
			log.Println("发送人地址SenderPrivateKey:", *t.SenderBlockchainAddress)
			log.Println("接收人地址RecipientBlockchainAddress:", *t.RecipientBlockchainAddress)
			log.Println("金额Value:", *t.Value)
			log.Println("交易Signature:", *t.Signature)

			if !t.Validate() {
				log.Println("ERROR: missing field(s)")
				io.WriteString(w, string(utils.JsonStatus("fail")))
				return
			}

			publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
			signature := utils.SignatureFromString(*t.Signature)
			bc := bcs.GetBlockchain()

			isCreated := bc.AddTransaction(*t.SenderBlockchainAddress,
				*t.RecipientBlockchainAddress, int64(*t.Value), publicKey, signature)

			w.Header().Add("Content-Type", "application/json")
			var m []byte
			if !isCreated {
				w.WriteHeader(http.StatusBadRequest)
				m = utils.JsonStatus("fail[from:blockchainServer]")
			} else {
				w.WriteHeader(http.StatusCreated)
				m = utils.JsonStatus("success[from:blockchainServer]")
			}
			io.WriteString(w, string(m))

		}

	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Mine(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		isMined := bc.Mining()

		var m []byte
		if !isMined {
			w.WriteHeader(http.StatusBadRequest)
			m = utils.JsonStatus("挖矿失败[from:Mine]")
		} else {
			m = utils.JsonStatus("挖矿成功[from:Mine]")
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m))
	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) GetChain(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		bc := bcs.GetBlockchain()
		m, _ := bc.MarshalJSON()
		io.WriteString(w, string(m[:]))
	default:
		log.Printf("ERROR: Invalid HTTP Method")

	}
}

func (bcs *BlockchainServer) StartMine(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		bc := bcs.GetBlockchain()
		bc.StartMining()

		m := utils.JsonStatus("success")
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m))
	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) Amount(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodPost:

		var data map[string]interface{}
		// 解析JSON数据

		err := json.NewDecoder(req.Body).Decode(&data)
		if err != nil {
			http.Error(w, "无法解析JSON数据", http.StatusBadRequest)
			return
		}
		// 获取JSON字段的值
		blockchainAddress := data["blockchain_address"].(string)

		color.Green("查询账户: %s 余额请求", blockchainAddress)

		amount := bcs.GetBlockchain().CalculateTotalAmount(blockchainAddress)

		ar := &block.AmountResponse{Amount: amount}
		m, _ := ar.MarshalJSON()

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m[:]))

	default:
		log.Printf("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}
func (bcs *BlockchainServer) BlockNumber(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		color.Cyan("####################BlockNumber###############")
		var number map[string]interface{}
		// 解析JSON数据

		err := json.NewDecoder(req.Body).Decode(&number)
		if err != nil {
			http.Error(w, "无法解析JSON数据", http.StatusBadRequest)
			return
		}
		// 获取JSON字段的值
		blockNumber := number["block_number"].(float64)

		color.Green("区块信息: %s", blockNumber)

		bc := bcs.GetBlockchain()
		block,_ := bc.GetBlockByNumber(int(blockNumber))
		m, _ := block.MarshalJSON()

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m[:]))

	default:
		log.Printf("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}
func (bcs *BlockchainServer) BlockHash(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		color.Cyan("####################BlockHash###############")
		var number map[string]interface{}
		// 解析JSON数据

		err := json.NewDecoder(req.Body).Decode(&number)
		if err != nil {
			http.Error(w, "无法解析JSON数据", http.StatusBadRequest)
			return
		}
		// 获取JSON字段的值
		blockHash := number["block_hash"].(string)

		color.Green("区块信息: %s", blockHash)

		bc := bcs.GetBlockchain()
		block,_ := bc.GetBlockByHash(blockHash)
		m, _ := block.MarshalJSON()

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m[:]))

	default:
		log.Printf("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}
func (bcs *BlockchainServer) GetTransactionByHash(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		color.Cyan("####################GetTransactionByHash###############")
		var transaction map[string]interface{}
		// 解析JSON数据

		err := json.NewDecoder(req.Body).Decode(&transaction)
		if err != nil {
			http.Error(w, "无法解析JSON数据", http.StatusBadRequest)
			return
		}
		// 获取JSON字段的值
		transactionHash := transaction["transaction_hash"].(string)

		color.Green("交易信息: %s", transactionHash)

		bc := bcs.GetBlockchain()
		Transaction := bc.GetTransactionByHash(transactionHash)
		m, _ := Transaction.MarshalJSON()

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m[:]))

	default:
		log.Printf("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockchainServer) GetHistoryTransaction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//设置允许的方法
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	switch req.Method {
	case http.MethodGet:
		{
			// Get:显示历史交易内容
			w.Header().Add("Content-Type", "application/json")
			bc := bcs.GetBlockchain()

			transactions := bc.HistoryTransactionPool()
			m, _ := json.Marshal(struct {
				Transactions []*block.Transaction `json:"transactions"`
				Length       int                  `json:"length"`
			}{
				Transactions: transactions,
				Length:       len(transactions),
			})
			io.WriteString(w, string(m[:]))
		}
	
	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}
func (bcs *BlockchainServer) Run() {

	// http.HandleFunc("/", handle_HelloWorld)
	http.HandleFunc("/", bcs.GetChain)

	http.HandleFunc("/transactions", bcs.Transactions) //GET 方式和  POST方式
	http.HandleFunc("/mine", bcs.Mine)
	http.HandleFunc("/mine/start", bcs.StartMine)
	http.HandleFunc("/amount", bcs.Amount)
	http.HandleFunc("/blockNumber", bcs.BlockNumber)
	http.HandleFunc("/blockHash", bcs.BlockHash)
	http.HandleFunc("/getTransactionHash", bcs.GetTransactionByHash)
	http.HandleFunc("/historyTransaction", bcs.GetHistoryTransaction)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(int(bcs.Port())), nil))

}
