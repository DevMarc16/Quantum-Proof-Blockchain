package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"quantum-blockchain/chain/crypto"
	"quantum-blockchain/chain/types"
)

type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   *RPCError   `json:"error"`
	ID      int         `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func callRPC(method string, params []interface{}) (*RPCResponse, error) {
	req := RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://localhost:8545", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpcResp RPCResponse
	err = json.Unmarshal(body, &rpcResp)
	if err != nil {
		return nil, err
	}

	return &rpcResp, nil
}

func main() {
	fmt.Println("🚀 Deploying QTM Token with Fixed Quantum Keys")
	fmt.Println("============================================")

	// Use fixed keys for deployment - this should be a pre-funded address
	privateKeyHex := "105c6fdb2a0c2f9b459c75d1bd78bd38f956f2669a9175efa28a7cb35121972692f0865f67c98e447eae12778abf5ef8cbd9d60055091f69d38329d882e53aade262c81dfa14569b0294a5851903d0b9171098277a5d0c3ea28b89163fc1086414024104c105a2b68d02336a24138edc889153428ac23268129090caa01023479221b591d8004199a424113968db168d433452a04486a396010921220b162a40b6891416901ab989204831e28265a4148cca322508468ea41410894431dbb02d18a4049400091b428219a949042024608041d008464a003299c8618a3646a3b425dc166c1122510c460cd2020403452210282e21130d242832888645c138110b198cd0a66c893886db188d4cb64d083702dab244a0a2881902028a008e8ab48c0c35884316441b0969cb426e09156ad9340d049871a0986c4804601cb04118a5089a240240388d1405050a430e88360c24b28413843159a0445b964864124481802192a0515916806202888ac06c00040c19390a24118991448c94860889b010e41061230786d00432cc3611c2960d48a8444c9211c2146812232840044d4ac62d19318293463221180d14868103326851a404e03871da962d1b3872933445e0c24ca1429201194d1841248a406d1493201ab94c51160601913118049018a32498226523a84809a46d8b341044c48d84b200d9406d8b228cc0048ca3360c40a4491a25501aa63153102d04204e94a8611c276d83363118080692b670e026095890410ac0215122711a480c09b5311ca7300822845a1632caa485001705a140304ac89101270ed10449e20449912089cb840c0bb341414404cc806d4b3846e2120a0081654ab81102b1448382408a800004272002958812c18c0cb92814c16c9b324853c64dd9400681242494200989a26020b240539200a01472da160891b28d5b84441b960842828d0b426c418864830405da9691139848d40660104928213605604271c1186100c440493625e3860c4c046c4134306418701bc49090801180a88d91408814b66594a0200c478804b56123820c41384063266e1aa781900232d3b668c126250c85501c444863208c52b4300019009b064e20424524024a63060993948812326821189002272d1445821b358e4c262a54226a492272c4c82954488023236060248c832651114229cb384ed946659a108c22c6251908324bb88d9f0fa1d8df0a54a1ea92a24708f8a4daead9498e2741a408247b1d750d1537f64052b2b9096b2e32a9d0de5f0e737d1e04cb8d81bf7935490b897d1b6c3dfe849773c238ab0b9e51b475cfea71642830648b32ab4e9e0e13b83fde2658eabeae48337dafa548afa71793911a4389203a50dfe72988624626b711d29b445eaf7e45ab23f9fd8574cadc981c89e49ec8e58246970502fbedf35815300ff5e82d5cd2123cb67055f209cdda30ac74b846047f7566f46bef85b38ef8e32625e33568c5e1a02237f2b61fa9f3bb79ffcec8c8d54290be88c9e28550d8faf20eba795b413f90018ec4781fa2df513aea011b60c01f129e88799d454b1f96892ae534726caf0535024cf66d38bc223f9f4970e4f7ae639d0b8984c243a7c96c71d2f9db4be5b52a778b3eaeed6fd76171e7ed20b87d749858428f3270fa920c13264f0262e07697ad124f18204909f9b1eb8845d8c7d1c5373d069222db149422912bf00bbf742f9ff9931efca0eaac471b3272a21db7bbbfade7bbfa792aabefead76072724f3b77c3a6d3bbd1b3c773ec4f482a6d0087941d5e9d0204227ece818260165d6f8dc23adfd93c1471ae31b38a2c27209d4ea05cf390f13a5cd0c9670361468fca74cc11c5a00b2dbb3ce29a92a6e96b32de49359759625bab9277a20c5e63f765466d2e378b42a03d685ff6b2e55fb8dffebad6c965391a490aae98ea2cc7568e1f0a9e4093ceea3b4fa0521b049fdcdecad3d4eac002324784bb1c31a0a75812ae6836a4265b3fc6ce87f9b9bd76d2e315ea5fa8bda463f25464ec7e12dfa24010e5514e8124a013bef91ecd8e5a3373ad24ee0a986c0d0ca20e00825c8acf4e2b1a2d83a0e9f7272c674f81672e82c170b0cbaa80069d4158f91d95277e241762cf9839b6cb4cbcfe8bbca0b0eadaf7814e26631e23c5bea6b64827d0c5a3de1b5b64c79bdeff3c31798a05a79d945c4155d18e8b493d26259c26ea3d7abe70c2c30c64de03023099d87306aef2b8b698828d877020653835ef1279c1720ffcd1b8224927bc98b8289f230e27f8572a60501815a8804f77f8158f34db55c3776d67da655db683cc4f974f048889fdc9a2e34af4957db2454463b75e6532af3646848fc65dcb563360e504c34387d3009494af3ac19fb2a6fe4b0ba93058666ff22a590e7633ada3a6497f54b6f6f2ffeee3eb7c878b807adf3697e68ccd5f629de213a5185273d0e6e9ad52bb820d3d4a46eb93e04af4cf3c76cfa9e71416bd2e357fa254ff2e00f850d1e353874542c67922aadcea06a50c7e142590b7bd12ff72cb2b5deeeee9d387dc9d8454e3ead1c998a4204f7371eb6b3b8d9b313cd617ea0eee233b85d10771bfd4784f79ad28455f1c6b1447dfa903b57611b48f64ff46db506eaf3fb7c5e243c5e7f7fb6d06cf9081fd529dd78f4453153f866a68e131646e6837928267b0e219708a255d1e588d27c9869989e1cd1de3261b776691f4b54e76a40d3cf330e6ff7b446eb2356fa6237d3ebaf7e6409b38eba1f807e2c9c3d6fa495fec64766116245a141beefff2fca1d5b7d119f30cb263a6e8be2eb6da8c97d70c17d95e12ecd468dcfdc424e7a5c24ee368548f51a421b5028d0736feaa61a88211e493b8781dc3af60abeb415d5f777fcae8e5134f6bfe577c5d053b2cabc11a6890fc52b2bc28010791c211ce21536deb4fdb068a3a30bbc92b15683171265422b8b8aebe5d4af6ec5b86744b4cd347305b71e6b73cd6b92f0ac53dcaf1b3188b6dfb083a12c797cc579da41f1bb71802a483b8214103f1c0726fe975d461b3cdb80db59be4bababf3d5bf7b9299f4a79b1a90338bde82fbc551b63a166f30e89c152e2130c30f3534d41607a5af28146e1cefda4d8331fadcf7e8dfb8cd1b519f90ba671eb76dfcd920fe3df14e04379639ea7c5c2ba994af0910c030383994207187070881c170b2c8211b2d87beee00e6ee56930a2c668444cf88e607ab52d786dc530dc6c3fe4a019544d222fe128d4de881f024356704651ae362cdbc835b22a0080503e4e86cb3d56357d9533c100bfc6f0617a889c1bccd0612d173dc07c077977c9e6aa75737e1a63cec4711e0ca27ad6878485a8ff5a9219d70ee7b41c82a8ab64ba8cbd1abe9e69a9221b356bcfd475dde0bb60281e5c34b99ea23436e21384e947ff7718a33dbc61fdde49fad34e458c58f1349055cedf92d7d2692361cafa3540faf321c704f5bb211b8cf6968ef260a846ff8796206e77d7b5d3873640e85e36388079d6357d3d4fbb9a58971b4289afb07f0a734e6dfff78ef9f2b3518c3dc3409044d49b787f44512005f4d0e09b"

	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		log.Fatal("Failed to decode private key:", err)
	}

	// Get the public key and address
	publicKeyHex := "105c6fdb2a0c2f9b459c75d1bd78bd38f956f2669a9175efa28a7cb35121972699b26b5766b7abb16cf3ee2b0884e64b3499911232072eb355f0de19aa275060e2f7ccd8cfa4db1a3aa6e4fa9c1f4ac5b46cb5dd6596ad4fce5a232853375e56b44f3aa7a96f47067ad2a93c4d0801a17ba5f72ba87c02859280da25d0bc40cc0473e85811bacfe6d041003d8be030f6f3ac789a6b153343d218c6139e0dca3fab9001a05ebad830732ab43824e5aeea5b1d1b11ee2b73e4f35799e06cb3d9ef301485b8d7375abe5d1f54da78ada16036955fc2cbe9eddf3266d7833afce8494a0fa762973fcc99b9aec287d2f2ad83ffcbdd2413a10da82403b1786cd81665eb0903cca590c0948e20d0e8b66bb48664c23de65ec9a9cb58791a74f16bcc0fa067e1f0fba5b31a20fd02e90ef393a9df62cfb95f0d4c6cfc1a87a1a7331ef1fe2fb048e476e1c205a3f828fcbfde80a8efe666afc4cac7e37c7ac1c926e4a7869822c6eac98f0a2ea99a9cc55c4e45d1571972c19327a1b0573e9c0a78b76dcde721d8eab67bc8e98e49da64e653d0d4c3f05549815fd7b64a33406911c4bf1a2fa00ff84010618ac904935cde03f86115040ccd7163e9dd06afabc9b435a80f8da5bd5af7f41e44a9e0a62d7a04602d3bddd1863e2726b6d6e03e84f011e115aeca5576ed418e5ee35cbfdcb9772406e9853ea8c94515ae53c97395f76dd98d9ebdc3e3bf7ffc6d60dd58e5aa7425c7336412e24961d8823a6f233636f0460899111e6268a711bdd3d9470ce3a129c53148002e0efe22e34aa7f5e25d5b665c98b95651cca52834f34a147b17c859c53730a4a0cc1368dc470cbc20b5cea30213c040726d6b27120ec89970c3a985f5a5f1460975720ade0402b428616bbc2c1efbdbb3e163d57bb6503f2df093bfe0b249df38e3b03f2a66d499d7b8963ce7b1f5cb7c4b79dc65a7bd5d965619216c6677c6d071719d780e3e02e94c18a68fbed9d9d66bd90459b528a16d6de3117d5ae98f34c1f6fff68a241023ddc0511b28d8444f9efc66922eacb60513bca7c7b75353dfb42ba84293113e51e9bd48c47af059148d6683ed6256b1248c55a5c268b08a2b1bf52c4d1192c0299c3cdb310d28ef7c0685681218758a06e23cce3aba937c320f8e064cecbe11fe16eba7868b99619592cbcdd6410c2526ad7d84bf183bbac317b2b4beb54598c88513bd84705878d41be06ad88c88a2074c4baaa1b268d2d4fb721464092b9e8e15f79535939400bf2a2d26da61cc7a73d36a2fce6b97a9aaef36725493dea0d92debd9fc36f623547ccfe1928879b6de7ed0a84af19086bcfd4ec1055e8cc1901fd7e8d1b868a000a773481bfb162f3bc0ea440ec505556c094e3a5a1b3583b4e94df832e956812475d4db4fad3eae52c2a5e92993e39dc1d1783f6b427e0d7808d9fa727338ba08b9e978ad57c6b2972af83946b40bdf291caa487e4adea5eed7429e63964bff6fac4c8919ea23438790af270eaaa7f0c7b7172a19a597526f945e013fbf53bbd4bcaf1b2530f49c71239a6599e48a2638b17b5ebe3e08d4c7e215e7d60ff4e502b685560bd44c75693666ecc97d4f1debe87eeb3f354bad5e1d2e48db40dd75d7e1761e8ed2de73a7c208fa97a26c73e7cbd021cfe48a46fc1a08e739787c5631cf450e4f4bbf991d9d8bd0d9596e0d3af582ca9890a431c155853fcd92b97ed38608a3c76a42d0af9458866fd150ebbafa5aa4fea325955c3c7977ca3d373af5db79c15fb2690aec6ac7e4f125fcd54959ba84aaa75f17962fc76475a738dad207593e5d736a0740e5d917997069bac8950e53883d00b8ac269f54"
	publicKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		log.Fatal("Failed to decode public key:", err)
	}

	quantumAddr := types.PublicKeyToAddress(publicKeyBytes)
	fmt.Printf("✅ Using quantum address: %s\n", quantumAddr.Hex())

	// Check balance
	resp, err := callRPC("eth_getBalance", []interface{}{quantumAddr.Hex(), "latest"})
	if err != nil {
		log.Fatal("Failed to get balance:", err)
	}
	if resp.Error != nil {
		log.Fatal("RPC error getting balance:", resp.Error.Message)
	}

	balance := resp.Result.(string)
	fmt.Printf("💰 Current balance: %s\n", balance)

	if balance == "0x0" {
		fmt.Println("❌ Account has no balance - this needs to be pre-funded")
		fmt.Printf("🔧 To fund this account, add this to genesis or initialization:\n")
		fmt.Printf("   Address: %s\n", quantumAddr.Hex())
		fmt.Printf("   Balance: 5000000000000000000 (5 ETH)\n")
		return
	}

	// Get nonce for the quantum address
	resp, err = callRPC("eth_getTransactionCount", []interface{}{quantumAddr.Hex(), "latest"})
	if err != nil {
		log.Fatal("Failed to get nonce:", err)
	}
	if resp.Error != nil {
		log.Fatal("RPC error getting nonce:", resp.Error.Message)
	}

	nonce := resp.Result.(string)
	nonceInt, err := strconv.ParseUint(nonce[2:], 16, 64)
	if err != nil {
		log.Fatal("Failed to parse nonce:", err)
	}

	fmt.Printf("📊 Current nonce: %s (%d)\n", nonce, nonceInt)

	// Contract bytecode
	qtmBytecode := "608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506040518060400160405280600d81526020017f5175616e74756d20546f6b656e000000000000000000000000000000000000008152506001908161009c9190610275565b50"

	deploymentDataBytes, err := hex.DecodeString(qtmBytecode)
	if err != nil {
		log.Fatal("Failed to decode deployment data:", err)
	}

	// Create quantum transaction
	tx := &types.QuantumTransaction{
		ChainID:  big.NewInt(8888),
		Nonce:    nonceInt,
		GasPrice: big.NewInt(1000000000), // 1 gwei
		Gas:      2000000,                // 2M gas
		To:       nil,                    // Contract creation
		Value:    big.NewInt(0),
		Data:     deploymentDataBytes,
		SigAlg:   crypto.SigAlgDilithium,
	}

	fmt.Println("🖊️  Signing transaction with Dilithium...")

	// Sign the transaction
	err = tx.SignTransaction(privateKeyBytes, crypto.SigAlgDilithium)
	if err != nil {
		log.Fatal("Failed to sign transaction:", err)
	}

	fmt.Printf("✅ Transaction signed with %d byte signature\n", len(tx.Signature))

	// Marshal to JSON
	txJSON, err := tx.MarshalJSON()
	if err != nil {
		log.Fatal("Failed to marshal transaction:", err)
	}

	// Send transaction
	hexTx := fmt.Sprintf("0x%x", txJSON)
	fmt.Printf("📤 Sending deployment transaction (%d bytes)...\n", len(hexTx))

	resp, err = callRPC("eth_sendRawTransaction", []interface{}{hexTx})
	if err != nil {
		log.Fatal("Failed to send transaction:", err)
	}

	if resp.Error != nil {
		log.Printf("❌ Deployment failed: %s\n", resp.Error.Message)
		return
	}

	txHash := resp.Result.(string)
	fmt.Printf("✅ Transaction sent! Hash: %s\n", txHash)

	// Wait for mining
	fmt.Print("⏳ Waiting for transaction to be mined")
	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second)
		fmt.Print(".")

		resp, err := callRPC("eth_getTransactionReceipt", []interface{}{txHash})
		if err != nil {
			continue
		}

		if resp.Result != nil {
			fmt.Println(" ✅ Mined!")

			receipt := resp.Result.(map[string]interface{})
			if contractAddr, ok := receipt["contractAddress"]; ok && contractAddr != nil {
				fmt.Printf("🎉 QTM Token deployed to: %s\n", contractAddr)
				return
			}
		}
	}

	fmt.Println(" ❌ Timeout waiting for transaction")
}
