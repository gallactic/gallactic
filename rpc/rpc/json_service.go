package rpcc

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gallactic/gallactic/core/config"
	pb "github.com/gallactic/gallactic/rpc/grpc/proto3"
	"github.com/hyperledger/burrow/logging"
	"google.golang.org/grpc"
)

const (
	GALLACTIC = "gallactic."

	GET_ACCOUNTS        = GALLACTIC + "getAccounts"
	GET_ACCOUNT         = GALLACTIC + "getAccount"
	GET_VALIDATOR       = GALLACTIC + "getValidator"
	GET_STORAGE         = GALLACTIC + "getStorage"
	GET_STORAGE_AT      = GALLACTIC + "getStorageAt"
	GET_STATUS          = GALLACTIC + "getStatus"
	GET_LATEST_BLOCK    = GALLACTIC + "getLatestBlock"
	GET_BLOCKS          = GALLACTIC + "getBlocks"
	GET_BLOCK           = GALLACTIC + "getBlock"
	GET_CONSENSUS_STATE = GALLACTIC + "getConsensusState"
	GET_VALIDATORS      = GALLACTIC + "getValidators"
	GET_NETWORK_INFO    = GALLACTIC + "getNetworkInfo"
	GET_CHAIN_ID        = GALLACTIC + "getChainId"
	GET_PEERS           = GALLACTIC + "getPeers"
	GET_GENESIS         = GALLACTIC + "getGenesis"
	BROADCAST_TX        = GALLACTIC + "broadcastTx"
	GET_UNCONFIRMED_TXS = GALLACTIC + "getUnconfirmedTxs"
	GET_BLOCK_TXS       = GALLACTIC + "getBlockTxs"
	GET_LastBlock_Info  = GALLACTIC + "getLastBlockInfo"
)

// Create a new JSON-RPC 2.0 service for gallactic
func NewJSONService(conf *config.Config, codec Codec, logger *logging.Logger) HttpService {

	httpService := &JSONService{
		codec:  codec,
		logger: logger.WithScope("NewJSONService"),
	}
	bcSvc, ntwSvc, txSvc := newGrpcClient(*conf)
	dhMap := getMethods(codec, bcSvc, ntwSvc, txSvc)
	httpService.defaultHandlers = dhMap
	return httpService
}

// Process a request.
func (js *JSONService) Process(r *http.Request, w http.ResponseWriter) {

	// Create new request object and unmarshal.
	req := &RPCRequest{}
	decoder := json.NewDecoder(r.Body)
	errU := decoder.Decode(req)

	// Error when decoding.
	if errU != nil {
		js.writeError("Failed to parse request: "+errU.Error(), "",
			RPCErrorParseError, w)
		return
	}

	// Wrong protocol version.
	if req.JSONRPC != "2.0" {
		js.writeError("Wrong protocol version: "+req.JSONRPC, req.Id,
			RPCErrorInvalidRequest, w)
		return
	}

	mName := req.Method

	if handler, ok := js.defaultHandlers[mName]; ok {
		js.logger.TraceMsg("Request received",
			"id", req.Id,
			"method", req.Method)
		resp, errCode, err := handler(req, w)
		if err != nil {
			js.writeError(err.Error(), req.Id, errCode, w)
		} else {
			js.writeResponse(req.Id, resp, w)
		}
	} else {
		js.writeError("Method not found: "+mName, req.Id, RPCErrorMethodNotFound, w)
	}
}

func newGrpcClient(conf config.Config) (pb.BlockChainClient, pb.NetworkClient, pb.TransactionClient) {
	addr := conf.GRPC.ListenAddress
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return pb.NewBlockChainClient(conn), pb.NewNetworkClient(conn), pb.NewTransactionClient(conn)
}

func getMethods(codec Codec, bcSvc pb.BlockChainClient, ntwSvc pb.NetworkClient,
	txSvc pb.TransactionClient) map[string]RequestHandlerFunc {
	// rpm stands for RPC Service Map
	rpm := make(map[string]RequestHandlerFunc)

	rpm[GET_ACCOUNT] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &pb.AddressRequest{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, RPCErrorInvalidParams, err
		}
		acc, err := bcSvc.GetAccount(context.Background(), input)
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return acc, 0, nil
	}

	rpm[GET_ACCOUNTS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		list, err := bcSvc.GetAccounts(context.Background(), &pb.Empty{})
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return list, 0, nil
	}

	rpm[GET_BLOCK] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &pb.BlockRequest{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, RPCErrorInvalidParams, err
		}
		block, err := bcSvc.GetBlock(context.Background(), input)
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return block, 0, nil
	}

	rpm[GET_BLOCK_TXS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &pb.BlockRequest{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, RPCErrorInvalidParams, err
		}
		transactions, err := bcSvc.GetBlockTxs(context.Background(), input)
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return transactions, 0, nil
	}

	rpm[GET_CONSENSUS_STATE] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		consensusState, err := bcSvc.GetConsensusState(context.Background(), &pb.Empty{})
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return consensusState, 0, nil
	}

	rpm[GET_CHAIN_ID] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		chainID, err := bcSvc.GetChainID(context.Background(), &pb.Empty{})
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return chainID, 0, nil
	}

	rpm[GET_GENESIS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		genesis, err := bcSvc.GetGenesis(context.Background(), &pb.Empty{})
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return genesis, 0, nil
	}

	rpm[GET_LATEST_BLOCK] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		resultGetBlock, err := bcSvc.GetLatestBlock(context.Background(), &pb.Empty{})
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return resultGetBlock, 0, nil
	}

	rpm[GET_NETWORK_INFO] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		info, err := ntwSvc.GetNetworkInfo(context.Background(), &pb.Empty1{})
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return info, 0, nil
	}

	rpm[GET_PEERS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		peers, err := ntwSvc.GetNetworkInfo(context.Background(), &pb.Empty1{})
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return peers, 0, nil
	}

	rpm[GET_STATUS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		status, err := bcSvc.GetStatus(context.Background(), &pb.Empty{})
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return status, 0, nil
	}

	rpm[GET_STORAGE] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &pb.StorageRequest{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, RPCErrorInvalidParams, err
		}
		storage, err := bcSvc.GetStorage(context.Background(), input)
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return storage, 0, nil
	}

	rpm[GET_STORAGE_AT] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &pb.StorageAtRequest{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, RPCErrorInvalidParams, err
		}

		storageItem, err := bcSvc.GetStorageAt(context.Background(), input)
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return storageItem, 0, nil
	}

	// rpm[GET_UNCONFIRMED_TXS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
	// 	transactions, err := bcSvc.Get
	// 	if err != nil {
	// 		return nil, RPCErrorInternalError, err
	// 	}
	// 	return transactions, 0, nil
	// }

	rpm[GET_VALIDATOR] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &pb.AddressRequest{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, RPCErrorInvalidParams, err
		}
		val, err := bcSvc.GetValidator(context.Background(), input)
		if val == nil || err != nil {
			return nil, RPCErrorInternalError, err
		}
		return val, 0, nil
	}

	rpm[GET_VALIDATORS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		list, err := bcSvc.GetAccounts(context.Background(), &pb.Empty{})
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return list, 0, nil
	}

	return rpm
}

// Helper for writing error responses.
func (js *JSONService) writeError(msg, id string, code int, w http.ResponseWriter) {
	response := NewRPCErrorResponse(id, code, msg)
	err := js.codec.Encode(response, w)
	// If there's an error here all bets are off.
	if err != nil {
		http.Error(w, "Failed to marshal standard error response: "+err.Error(), 500)
		return
	}
	w.WriteHeader(200)
}

// Helper for writing responses.
func (js *JSONService) writeResponse(id string, result interface{}, w http.ResponseWriter) {
	response := NewRPCResponse(id, result)
	err := js.codec.Encode(response, w)
	if err != nil {
		js.writeError("Internal error: "+err.Error(), id, RPCErrorInternalError, w)
		return
	}
	w.WriteHeader(200)
}
