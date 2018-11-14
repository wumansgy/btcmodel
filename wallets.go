package main

import (
	"fmt"
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"crypto/elliptic"
	"sort"
)

const walletFileName = "wallet.dat"

//1. 创建一个保存所有wallet的集合，交wallets
//2. address -> wallet // map[string]*wallet

type Wallets struct {
	//key : 这个私钥对对应的地址，
	//value : 这个私钥对
	WalletsMap map[string]*Wallet
}

//创建wallets实例,供后续迭代操作
func NewWallets() *Wallets {
	var wallets Wallets

	wallets.WalletsMap = make(map[string]*Wallet)
	//1. 从本地文件将已经存储Wallets结构加载到内存
	wallets.LoadFromFile()

	//2. 将wallets返回
	return &wallets
}

//创建一个钱包
func (ws *Wallets) CreateWallet() string {
	fmt.Printf("CreateWallet...\n")

	//1. 创建一个wallet
	wallet := NewWallet()

	address := wallet.GetAddress()

	//2. 添加到map中
	ws.WalletsMap[address] = wallet
	//3. 保存到本地
	if !ws.SaveToFile() {
		return ""
	}

	return address
}

func (ws *Wallets) GetAddresses() []string {

	var addresses []string

	for address := range ws.WalletsMap {
		addresses = append(addresses, address)
	}

	// Strings sorts a slice of strings in increasing order.递增
	sort.Strings(addresses)
	return addresses
}

func (ws *Wallets) LoadFromFile() bool {
	fmt.Printf("LoadFromFile...\n")
	//1. 判断文件是否存在
	//2. 读取文件
	//3. 解码
	//4. 填充ws.map

	if !IsFileExist(walletFileName) {
		return false
	}

	data, err := ioutil.ReadFile(walletFileName)
	if err != nil {
		return false
	}

	//注册接口数据类型
	gob.Register(elliptic.P256())

	decoder := gob.NewDecoder(bytes.NewReader(data))

	var wsLocal Wallets
	err = decoder.Decode(&wsLocal)

	if err != nil {
		fmt.Println(err)
		fmt.Printf("解码失败!\n")
		return false
	}

	ws.WalletsMap = wsLocal.WalletsMap

	return true
}

func (ws *Wallets) SaveToFile() bool {
	fmt.Printf("SaveToFile...\n")
	//1. 把wallets里面的数据使用gob编码

	var buffer bytes.Buffer

	//gob: type not registered for interface: elliptic.p256Curve
	//1. gob编码的结构里面如果涉及到了interface类型的数据，要对gob进行注册

	//注册接口数据类型
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(ws)

	if err != nil {
		fmt.Println(err)
		return false
	}

	//2. 写入文件
	//func WriteFile(filename string, data []byte, perm os.FileMode) error {
	err = ioutil.WriteFile(walletFileName, buffer.Bytes(), 0600)

	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}
