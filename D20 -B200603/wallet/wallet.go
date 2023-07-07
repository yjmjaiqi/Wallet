package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"jhblockchain/utils"
	"math/big"
	"strconv"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil/base58"
)

type Wallet struct {
	privateKey        *ecdsa.PrivateKey
	publicKey         *ecdsa.PublicKey
	blockchainAddress string
}

func NewWallet() *Wallet {

	w := new(Wallet)
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	w.privateKey = privateKey
	w.publicKey = &w.privateKey.PublicKey

	h := sha256.New()
	h.Write(w.publicKey.X.Bytes())
	h.Write(w.publicKey.Y.Bytes())
	digest := h.Sum(nil)
	address := base58.Encode(digest)
	w.blockchainAddress = address

	return w
}

func LoadWallet(privkey string) *Wallet {
	theWallet := new(Wallet)
	thepriKey := new(ecdsa.PrivateKey)

	privateKey := privkey
	privateKey_D := new(big.Int)
	privateKey_D.SetString(privateKey, 16)

	thepriKey.D = privateKey_D

	//得到 publicKey对象
	// 曲线
	curve := elliptic.P256()
	// 获取公钥
	x, y := curve.ScalarBaseMult(privateKey_D.Bytes())
	publicKey := ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}

	thepriKey.PublicKey = publicKey
	theWallet.privateKey = thepriKey
	theWallet.publicKey = &publicKey
	//计算address
	h := sha256.New()
	h.Write(publicKey.X.Bytes())
	h.Write(publicKey.Y.Bytes())

	digest := h.Sum(nil)
	// fmt.Printf("digest: %x\n", digest)
	address := base58.Encode(digest)

	theWallet.blockchainAddress = address

	return theWallet
}

func (w *Wallet) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		PrivateKey        string `json:"private_key"`
		PublicKey         string `json:"public_key"`
		BlockchainAddress string `json:"blockchain_address"`
	}{
		PrivateKey:        w.PrivateKeyStr(),
		PublicKey:         w.PublicKeyStr(),
		BlockchainAddress: w.BlockchainAddress(),
	})
}
func (w *Wallet) UnmarshalJSON(data []byte) error {
	aux := struct {
		PrivateKey        string `json:"private_key"`
		PublicKey         string `json:"public_key"`
		BlockchainAddress string `json:"blockchain_address"`
	}{}

	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	privateKeyBytes, err := hex.DecodeString(aux.PrivateKey)
	if err != nil {
		return err
	}
	privateKey := new(ecdsa.PrivateKey)
	privateKey.D = new(big.Int).SetBytes(privateKeyBytes)

	publicKeyBytes, err := hex.DecodeString(aux.PublicKey)
	if err != nil {
		return err
	}
	publicKey := new(ecdsa.PublicKey)
	publicKey.Curve = privateKey.Curve
	publicKey.X = new(big.Int).SetBytes(publicKeyBytes[:len(publicKeyBytes)/2])
	publicKey.Y = new(big.Int).SetBytes(publicKeyBytes[len(publicKeyBytes)/2:])

	w.privateKey = privateKey
	w.publicKey = publicKey
	w.blockchainAddress = aux.BlockchainAddress

	return nil
}

// 为什么要写以下返回私钥和公钥的方法
func (w *Wallet) PrivateKey() *ecdsa.PrivateKey {

	return w.privateKey
}

func (w *Wallet) PrivateKeyStr() string {
	return fmt.Sprintf("%x", w.privateKey.D.Bytes())
}

func (w *Wallet) PublicKey() *ecdsa.PublicKey {
	return w.publicKey
}

func (w *Wallet) PublicKeyStr() string {
	return fmt.Sprintf("%x%x", w.publicKey.X.Bytes(), w.publicKey.Y.Bytes())
}

func (w *Wallet) BlockchainAddress() string {
	return w.blockchainAddress
}

func FromPriKeyToPubKey(privkey string) {
	privateKey := privkey
	privateKeyInt := new(big.Int)
	privateKeyInt.SetString(privateKey, 16)
	fmt.Println("privateKeyInt:", privateKeyInt)
	// 曲线
	curve := elliptic.P256()
	// 获取公钥
	x, y := curve.ScalarBaseMult(privateKeyInt.Bytes())
	publicKey := ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}
	fmt.Println("Public Key : \n", publicKey)
	fmt.Printf("Public Key X: %x\n", publicKey.X)
	fmt.Printf("Public Key y: %x\n", publicKey.Y)

}

type Transaction struct {
	senderPrivateKey           *ecdsa.PrivateKey
	senderPublicKey            *ecdsa.PublicKey
	senderBlockchainAddress    string
	recipientBlockchainAddress string
	value                      uint64
	transactionHash            [32]byte
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string `json:"sender_blockchain_address"`
		Recipient string `json:"recipient_blockchain_address"`
		Value     uint64 `json:"value"`
		TransactionHash  string `json:"transactionHash"`
	}{
		Sender:    t.senderBlockchainAddress,
		Recipient: t.recipientBlockchainAddress,
		Value:     t.value,
		TransactionHash: fmt.Sprintf("%x",t.transactionHash),
	})
}
func Hash(sender string,receive string,value uint64) [32]byte {
	v := strconv.FormatUint(value,10)
	m, _ := json.Marshal(sender+receive+v)
	return sha256.Sum256([]byte(m))
}
func NewTransaction(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey,
	sender string, recipient string, value uint64,hash [32]byte) *Transaction {
	return &Transaction{privateKey, publicKey, sender, recipient, value,hash}
}

func (t *Transaction) GenerateSignature() *utils.Signature {
	m, _ := json.Marshal(t)
	h := sha256.Sum256([]byte(m))
	r, s, _ := ecdsa.Sign(rand.Reader, t.senderPrivateKey, h[:])
	return &utils.Signature{R: r, S: s}
}
