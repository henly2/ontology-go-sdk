package ont_api

const (
	gasPrice = 500
	gasLimit = 20000

	gasToAddress = "AFmseVrdL9f9oyCzZefL9tG6UbviEH9ugK"
)

const (
	TxStateSuccess = 1

	ExpDescSUCCESS = "SUCCESS"

	OperationTransfer = "transfer"
)

type (
	TokenInfo struct {
		Name          string
		Symbol        string
		Precision     int
		TotalSupply   uint64
		ScriptHashHex string
	}

	OperationInfo struct {
		*TokenInfo

		FuncName string
		From     string
		To       string
		Amount   uint64
		IsFee    bool
	}

	TransactionInfo struct {
		TxHash      string
		State       byte
		GasConsumed uint64

		Operations []OperationInfo
	}

	BlockInfo struct {
		Version uint32
		//PrevBlockHash    common.Uint256
		//TransactionsRoot common.Uint256
		//BlockRoot        common.Uint256
		Timestamp uint32
		Height    uint32
		//ConsensusData    uint64
		//ConsensusPayload []byte
		//NextBookkeeper   common.Address

		//Program *program.Program
		//Bookkeepers []keypair.PublicKey
		//SigData     [][]byte

		//hash *common.Uint256
		Hash string

		Txs []string
	}

	///////////////////////////////////////
	// explorer api
	/*
			{
		    "Action":"QueryTransaction",
		    "Version":"1.0",
		    "Error":0,
		    "Desc":"SUCCESS",
		    "Result":{
		        "TxnHash":"9762458cd30612509f7cda589b7e1f7e59cb35d200eeb0a010ccc7b347057eb5",
		        "TxnType":209,
		        "TxnTime":1522207312,
		        "Height":1212,
		        "ConfirmFlag":1,
		        "BlockIndex":1,
		        "Fee":"0.010000000",
		        "Description":"transfer",
		        "Detail":{
					}
				}
			}
	*/
	ExpTransactionInfo struct {
		Action  string
		Version string
		Error   int
		Desc    string
		Result  struct {
			TxnHash     string
			TxnType     int
			TxnTime     int
			Height      int
			ConfirmFlag int
			BlockIndex  int
			Fee         string
			Description string
		}
	}
)
