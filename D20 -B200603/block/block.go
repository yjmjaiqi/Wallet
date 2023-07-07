package block

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"jhblockchain/utils"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
	"os"
	"encoding/hex"

	
	"github.com/fatih/color"
)

var MINING_DIFFICULT = 0x800
const MINING_ACCOUNT_ADDRESS = "YIJIAMING BLOCKCHAIN"
const MINING_REWARD = 9
const MINING_TIMER_SEC = 10

type Block struct {
	nonce        int
	number 		 int
	previousHash [32]byte
	hash 		 [32]byte
	timestamp    int64
	transactions []*Transaction
}

func NewBlock(nonce int,number int, previousHash [32]byte, txs []*Transaction) *Block {
	b := new(Block)
	b.timestamp = time.Now().UnixNano()
	b.nonce = nonce
	b.number = number
	b.hash = b.Hash()
	b.previousHash = previousHash
	b.transactions = txs
	return b
}
func (b *Block) Print() {
	log.Printf("%-15v:%30d\n", "timestamp", b.timestamp)
	//fmt.Printf("timestamp       %d\n", b.timestamp)
	log.Printf("%-15v:%30d\n", "nonce", b.nonce)
	log.Printf("%-15v:%30x\n", "previous_hash", b.previousHash)
	//log.Printf("%-15v:%30s\n", "transactions", b.transactions)
	for _, t := range b.transactions {
		t.Print()
	}
}

type Blockchain struct {
	transactionPool   []*Transaction
	chain             []*Block
	blockchainAddress string
	port              uint16
	mux               sync.Mutex
}

// 新建一条链的第一个区块
// NewBlockchain(blockchainAddress string) *Blockchain
// 函数定义了一个创建区块链的方法，它接收一个字符串类型的参数 blockchainAddress，
// 它返回一个区块链类型的指针。在函数内部，它创建一个区块链对象并为其设置地址，
// 然后创建一个创世块并将其添加到区块链中，最后返回区块链对象。
func NewBlockchain(blockchainAddress string, port uint16) *Blockchain {
	bc := new(Blockchain)
	b := &Block{}
	// bc.CreateBlock(0,1,b.Hash()) //创世纪块
	bc.blockchainAddress = blockchainAddress
	bc.port = port
	blocks, _ := LoadBlocks()
	if blocks != nil{
		bc.chain = blocks
	}else{
		bc.CreateBlock(0,1,b.Hash()) //创世纪块
	}
	return bc
}
func (bc *Blockchain) TransactionPool() []*Transaction {
	return bc.transactionPool
}

func (bc *Blockchain) HistoryTransactionPool() []*Transaction {
	historyTransactions := make([]*Transaction, 0)
	for _, block := range bc.chain {
		historyTransactions = append(historyTransactions,block.transactions...)
	}
	return historyTransactions
}

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Blocks []*Block `json:"chains"`
	}{
		Blocks: bc.chain,
	})
}

// (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block
//  函数是在区块链上创建新的区块，它接收两个参数：一个int类型的nonce和一个字节数组类型的 previousHash，
//  返回一个区块类型的指针。在函数内部，它使用传入的参数来创建一个新的区块，
//  然后将该区块添加到区块链的链上，并清空交易池。

func (bc *Blockchain) CreateBlock(nonce int,number int, previousHash [32]byte) *Block {
	b := NewBlock(nonce,number, previousHash, bc.transactionPool)
	bc.chain = append(bc.chain, b)
	bc.transactionPool = []*Transaction{}
	file, err := os.OpenFile("yijiaming.txt", os.O_CREATE | os.O_APPEND | os.O_RDWR,0744)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	err = enc.Encode(b)
	if err != nil{
		return nil
	}
	return b
}
// 从文件读取区块信息
func LoadBlocks() ([]*Block, error){
	file, err := os.Open("yijiaming.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	blocks := make([]*Block, 0)
	dec := json.NewDecoder(file)
	for dec.More() {
		var bl Block
		if err := dec.Decode(&bl); err != nil {
			return nil, err
		}
		blocks = append(blocks, &bl)
	}

	return blocks, nil
}

func (bc *Blockchain) Print() {
	for i, block := range bc.chain {
		color.Green("%s BLOCK %d %s\n", strings.Repeat("=", 25), i, strings.Repeat("=", 25))
		block.Print()
	}
	color.Yellow("%s\n\n\n", strings.Repeat("*", 50))
}

func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	return sha256.Sum256([]byte(m))
}

func (b *Block) MarshalJSON() ([]byte, error) {

	return json.Marshal(struct {
		Timestamp    int64          `json:"timestamp"`
		Nonce        int            `json:"nonce"`
		Number        int            `json:"number"`
		PreviousHash string         `json:"previous_hash"`
		Hash 		string         `json:"hash"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Timestamp:    b.timestamp,
		Nonce:        b.nonce,
		Number:		  b.number,
		PreviousHash: fmt.Sprintf("%x", b.previousHash),
		Hash:    	  fmt.Sprintf("%x",b.hash),
		Transactions: b.transactions,
	})
}

func (b *Block) UnmarshalJSON(data []byte) error {
	var previousHash string
	var hash string
	v := &struct {
		Timestamp    *int64          `json:"timestamp"`
		Nonce        *int            `json:"nonce"`
		Number        *int            `json:"number"`
		PreviousHash *string         `json:"previous_hash"`
		Hash 		 *string         `json:"hash"`
		Transactions *[]*Transaction `json:"transactions"`
	}{
		Timestamp:    &b.timestamp,
		Nonce:        &b.nonce,
		Number:		  &b.number,
		PreviousHash: &previousHash,
		Hash:    	  &hash,
		Transactions: &b.transactions,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ph, _ := hex.DecodeString(*v.PreviousHash)
	copy(b.previousHash[:], ph[:32])
	ha, _ := hex.DecodeString(*v.Hash)
	copy(b.hash[:], ha[:32])
	return nil
}
func (bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) CopyTransactionPool() []*Transaction {
	transactions := make([]*Transaction, 0)
	for _, t := range bc.transactionPool {
		transactions = append(transactions,
			NewTransaction(t.senderAddress,
				t.receiveAddress,
				t.value))
	}
	return transactions
}
func (bc *Blockchain) ValidProof(nonce int,
	previousHash [32]byte,
	transactions []*Transaction,
	difficulty int,
) bool {
	bigi_2 := big.NewInt(2)
	bigi_256 := big.NewInt(256)
	big_diff := big.NewInt(int64(difficulty))

	target := new(big.Int).Exp(bigi_2,bigi_256,nil)
	target = new(big.Int).Div(target,big_diff)
	// zeros := strings.Repeat("0", difficulty)
	tmpBlock := Block{nonce: nonce, previousHash: previousHash, transactions: transactions, timestamp: time.Now().UnixNano()}
	// tmpHashStr := fmt.Sprintf("%x", tmpBlock.Hash())
	result := bytesToBigInt(tmpBlock.Hash())
	//log.Println("guessHashStr", tmpHashStr)
	return target.Cmp(result) > 0
}
func bytesToBigInt(b [32]byte) *big.Int{
	bytes := b[:]
	result := new(big.Int).SetBytes(bytes)
	return result
}

func (bc *Blockchain) ProofOfWork() int {
	transactions := bc.CopyTransactionPool() //选择交易？控制交易数量？
	previousHash := bc.LastBlock().Hash()
	nonce := 0
	if bc.getTime(len(bc.chain)-1) < 3e+9{
		MINING_DIFFICULT += 32
	}else{
		if MINING_DIFFICULT >= 130000{
			MINING_DIFFICULT -= 32
		}
	}
	begin := time.Now()
	for !bc.ValidProof(nonce, previousHash, transactions, MINING_DIFFICULT) {
		nonce += 1
	}
	end := time.Now()

	log.Printf("POW spend Time:%f Second", end.Sub(begin).Seconds())
	log.Printf("POW spend Time:%s", end.Sub(begin))

	return nonce
}
func (bc *Blockchain) getTime(blockNum int) int64{
	if blockNum == 0{
		return 0
	}
	return int64(bc.chain[blockNum].timestamp-bc.chain[blockNum-1].timestamp)
}
// 将交易池的交易打包
func (bc *Blockchain) Mining() bool {
	bc.mux.Lock()

	defer bc.mux.Unlock()

	// 此处判断交易池是否有交易，你可以不判断，打包无交易区块
	if len(bc.transactionPool) == 0 {
		return false
	}

	bc.AddTransaction(MINING_ACCOUNT_ADDRESS, bc.blockchainAddress, MINING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash()
	bc.CreateBlock(nonce, bc.LastBlock().number+1,previousHash)
	log.Println("action=mining, status=success")
	return true
}

func (bc *Blockchain) CalculateTotalAmount(accountAddress string) uint64 {
	var totalAmount uint64 = 0
	for _, _chain := range bc.chain {
		for _, _tx := range _chain.transactions {
			if accountAddress == _tx.receiveAddress {
				totalAmount = totalAmount + uint64(_tx.value)
			}
			if accountAddress == _tx.senderAddress {
				totalAmount = totalAmount - uint64(_tx.value)
			}
		}
	}
	return totalAmount
}

func (bc *Blockchain) StartMining() {
	bc.Mining()
	// 使用time.AfterFunc函数创建了一个定时器，它在指定的时间间隔后执行bc.StartMining函数（自己调用自己）。
	_ = time.AfterFunc(time.Second*MINING_TIMER_SEC, bc.StartMining)
	color.Yellow("minetime: %v\n", time.Now())
}

type AmountResponse struct {
	Amount uint64 `json:"amount"`
}

func (ar *AmountResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Amount uint64 `json:"amount"`
	}{
		Amount: ar.Amount,
	})
}

type Transaction struct {
	senderAddress  string
	receiveAddress string
	value          int64
	transactionHash [32]byte
}

func NewTransaction(sender string, receive string, value int64) *Transaction {
	t := new(Transaction)
	t.senderAddress = sender
	t.receiveAddress = receive
	t.value = value
	t.transactionHash = Hash(sender,receive,uint64(value))
	return t
}
func Hash(sender string,receive string,value uint64) [32]byte {
	v := strconv.FormatUint(value,10)
	m, _ := json.Marshal(sender+receive+v)
	return sha256.Sum256([]byte(m))
}


func (bc *Blockchain) AddTransaction(
	sender string,
	recipient string,
	value int64,
	senderPublicKey *ecdsa.PublicKey,
	s *utils.Signature) bool {
	t := NewTransaction(sender, recipient, value)

	//如果是挖矿得到的奖励交易，不验证
	if sender == MINING_ACCOUNT_ADDRESS {
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}

	// 判断有没有足够的余额
	// log.Printf("transaction.go sender:%s  account=%d", sender, bc.CalculateTotalAmount(sender))
	// if bc.CalculateTotalAmount(sender) <= uint64(value) {
	// 	log.Printf("ERROR: %s ，你的钱包里没有足够的钱", sender)
	// 	return false
	// }

	if bc.VerifyTransactionSignature(senderPublicKey, s, t) {

		bc.transactionPool = append(bc.transactionPool, t)
		return true
	} else {
		log.Println("ERROR: 验证交易")
	}
	return false

}

func (bc *Blockchain) VerifyTransactionSignature(
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature, t *Transaction) bool {
	m, _ := json.Marshal(t)
	h := sha256.Sum256([]byte(m))
	return ecdsa.Verify(senderPublicKey, h[:], s.R, s.S)
}

func (t *Transaction) Print() {
	color.Red("%s\n", strings.Repeat("~", 30))
	color.Cyan("发送地址             %s\n", t.senderAddress)
	color.Cyan("接受地址             %s\n", t.receiveAddress)
	color.Cyan("金额                 %d\n", t.value)

}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string `json:"sender_blockchain_address"`
		Recipient string `json:"recipient_blockchain_address"`
		Value     int64  `json:"value"`
		TransactionHash  string `json:"transactionHash"`
	}{
		Sender:    t.senderAddress,
		Recipient: t.receiveAddress,
		Value:     t.value,
		TransactionHash: fmt.Sprintf("%x",t.transactionHash),
	})
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	var transactionHash string
	v := &struct {
		Sender    *string `json:"sender_blockchain_address"`
		Recipient *string `json:"recipient_blockchain_address"`
		Value     *int64  `json:"value"`
		TransactionHash  *string `json:"transactionHash"`
	}{
		Sender:    &t.senderAddress,
		Recipient: &t.receiveAddress,
		Value:     &t.value,
		TransactionHash: &transactionHash,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	tran, _ := hex.DecodeString(*v.TransactionHash)
	copy(t.transactionHash[:], tran[:32])
	return nil
}
type TransactionRequest struct {
	SenderBlockchainAddress    *string `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string `json:"recipient_blockchain_address"`
	SenderPublicKey            *string `json:"sender_public_key"`
	Value                      *uint64 `json:"value"`
	Signature                  *string `json:"signature"`
}

func (tr *TransactionRequest) Validate() bool {
	if tr.SenderBlockchainAddress == nil ||
		tr.RecipientBlockchainAddress == nil ||
		tr.SenderPublicKey == nil ||
		tr.Value == nil ||
		tr.Signature == nil {
		return false
	}
	return true
}
func (bc *Block) GetBlockTransaction() []*Transaction {
	transactions := make([]*Transaction, 0)
	for _, t := range bc.transactions {
		transactions = append(transactions,
			NewTransaction(t.senderAddress,
				t.receiveAddress,
				t.value))
	}
	return transactions
}
// 根据区块ID输出该结构体内容
func (blockchain *Blockchain) GetBlockByNumber(blockid int) (*Block,error){
	for _,block := range blockchain.chain{
		if block.number == blockid{
			log.Printf("%-15v:%30d\n", "nonce", block.nonce)
			log.Printf("%-15v:%30d\n", "timestamp", block.timestamp)
			log.Printf("%-15v:%30d\n", "number", block.number)
			log.Printf("%-15v:%30x\n", "previousHash", block.previousHash)
			log.Printf("%-15v:%30x\n", "hash", block.hash)
			// log.Printf("%-15v:%s\n", "transactions", block.transactions)
			return block,nil
		}
	}
	log.Printf("%-15v:%30s\n", "error", "没找到对应区块ID结构体内容")
	return nil,errors.New("没找到对应区块信息")
	// return nil,errors.New("该区块不存在")
}
//根据区块哈希输出该结构体内容
func (blockchain *Blockchain) GetBlockByHash(hash string) (*Block,error){
	for _,block := range blockchain.chain{
		if hash == fmt.Sprintf("%x",block.hash){
			log.Printf("%-15v:%30d\n", "nonce", block.nonce)
			log.Printf("%-15v:%30d\n", "timestamp", block.timestamp)
			log.Printf("%-15v:%30d\n", "number", block.number)
			log.Printf("%-15v:%30x\n", "previousHash", block.previousHash)
			log.Printf("%-15v:%30x\n", "hash", block.hash)
			// log.Printf("%-15v:%30s\n", "transactions", block.transactions)
			return block,nil
		}
	}
	log.Printf("%-15v:%30s\n", "error", "没找到对应区块哈希的结构体内容")
	return nil,errors.New("没找到对应区块哈希的结构体内容")
}
//根据交易哈希输出该结构体内容
func (bc *Blockchain) GetTransactionByHash(hash string) *Transaction{
	for i,block := range bc.chain{
		for _,transaction := range block.transactions{
			if fmt.Sprintf("%x",transaction.transactionHash) == hash{
				log.Printf("%-15v:%30s\n", "该交易发送方", transaction.senderAddress)
				log.Printf("%-15v:%30s\n", "该交易接收方", transaction.receiveAddress)
				log.Printf("%-15v:%30d\n", "该交易价值", transaction.value)
				log.Printf("%-15v:%30d\n", "该交易hash", transaction.transactionHash)
				log.Printf("%-15v:%30d\n", "该交易所属区块为", i)
				return transaction
			}
		}
	}
	log.Printf("%-15v\n", "交易不存在")
	return nil
}
