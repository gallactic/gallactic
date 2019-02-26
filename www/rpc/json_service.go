package rpc

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gallactic/gallactic/core/config"
	pb "github.com/gallactic/gallactic/www/grpc/proto3"
	log "github.com/inconshreveable/log15"
	"google.golang.org/grpc"
)

const (
	rpcErrorServerError    = -32000
	rpcErrorInvalidRequest = -32600
	rpcErrorMethodNotFound = -32601
	rpcErrorInvalidParams  = -32602
	rpcErrorInternalError  = -32603
	rpcErrorParseError     = -32700
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
func NewJSONService(conf *config.Config, codec Codec) HttpService {

	httpService := &JSONService{
		codec: codec,
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
			rpcErrorParseError, w)
		return
	}

	// Wrong protocol version.
	if req.JSONRPC != "2.0" {
		js.writeError("Wrong protocol version: "+req.JSONRPC, req.Id,
			rpcErrorInvalidRequest, w)
		return
	}

	mName := req.Method

	if handler, ok := js.defaultHandlers[mName]; ok {
		log.Info(
			"Request received",
			"id", req.Id,
			"method", req.Method,
		)
		resp, errCode, err := handler(req, w)
		if err != nil {
			js.writeError(err.Error(), req.Id, errCode, w)
		} else {
			js.writeResponse(req.Id, resp, w)
		}
	} else {
		js.writeError("Method not found: "+mName, req.Id, rpcErrorMethodNotFound, w)
	}
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
		js.writeError("Internal error: "+err.Error(), id, rpcErrorInternalError, w)
		return
	}
	w.WriteHeader(200)
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

	rpm[BROADCAST_TX] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		txEnv := &pb.TransactRequest{
			Envelope: string(request.Params),
		}
		receipt, err := txSvc.BroadcastTxSync(context.Background(), txEnv)
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return receipt, 0, nil
	}

	rpm[GET_ACCOUNT] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &pb.AddressRequest{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, rpcErrorInvalidParams, err
		}
		acc, err := bcSvc.GetAccount(context.Background(), input)
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return acc, 0, nil
	}

	rpm[GET_ACCOUNTS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		list, err := bcSvc.GetAccounts(context.Background(), &pb.Empty{})
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return list, 0, nil
	}

	rpm[GET_BLOCK] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &pb.BlockRequest{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, rpcErrorInvalidParams, err
		}
		block, err := bcSvc.GetBlock(context.Background(), input)
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return block, 0, nil
	}

	rpm[GET_BLOCK_TXS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &pb.BlockRequest{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, rpcErrorInvalidParams, err
		}
		transactions, err := bcSvc.GetBlockTxs(context.Background(), input)
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return transactions, 0, nil
	}

	rpm[GET_CONSENSUS_STATE] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		consensusState, err := bcSvc.GetConsensusState(context.Background(), &pb.Empty{})
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return consensusState, 0, nil
	}

	rpm[GET_CHAIN_ID] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		chainID, err := bcSvc.GetChainID(context.Background(), &pb.Empty{})
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return chainID, 0, nil
	}

	rpm[GET_GENESIS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		genesis, err := bcSvc.GetGenesis(context.Background(), &pb.Empty{})
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return genesis, 0, nil
	}

	rpm[GET_LATEST_BLOCK] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		resultGetBlock, err := bcSvc.GetLatestBlock(context.Background(), &pb.Empty{})
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return resultGetBlock, 0, nil
	}

	rpm[GET_NETWORK_INFO] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		info, err := ntwSvc.GetNetworkInfo(context.Background(), &pb.Empty1{})
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return info, 0, nil
	}

	rpm[GET_PEERS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		peers, err := ntwSvc.GetNetworkInfo(context.Background(), &pb.Empty1{})
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return peers, 0, nil
	}

	rpm[GET_STATUS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		status, err := bcSvc.GetStatus(context.Background(), &pb.Empty{})
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return status, 0, nil
	}

	rpm[GET_STORAGE] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &pb.StorageRequest{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, rpcErrorInvalidParams, err
		}
		storage, err := bcSvc.GetStorage(context.Background(), input)
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return storage, 0, nil
	}

	rpm[GET_STORAGE_AT] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &pb.StorageAtRequest{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, rpcErrorInvalidParams, err
		}

		storageItem, err := bcSvc.GetStorageAt(context.Background(), input)
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return storageItem, 0, nil
	}

	rpm[GET_UNCONFIRMED_TXS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		transactions, err := txSvc.GetUnconfirmedTxs(context.Background(), &pb.Empty2{})
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return transactions, 0, nil
	}

	rpm[GET_VALIDATOR] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &pb.AddressRequest{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, rpcErrorInvalidParams, err
		}
		val, err := bcSvc.GetValidator(context.Background(), input)
		if val == nil || err != nil {
			return nil, rpcErrorInternalError, err
		}
		return val, 0, nil
	}

	rpm[GET_VALIDATORS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		list, err := bcSvc.GetValidators(context.Background(), &pb.Empty{})
		if err != nil {
			return nil, rpcErrorInternalError, err
		}
		return list, 0, nil
	}

	return rpm
}
