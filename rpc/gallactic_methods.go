package rpc

import (
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/txs"
)

const (
	GALLACTIC = "gallactic."

	GET_ACCOUNTS        = GALLACTIC + "getAccounts"
	GET_ACCOUNT         = GALLACTIC + "getAccount"
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
)

func loadGallacticMethods(codec Codec, service *Service, rpcServiceMap map[string]RequestHandlerFunc) {

	accountFilterFactory := NewAccountFilterFactory()

	rpcServiceMap[BROADCAST_TX] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {

		txEnv := new(txs.Envelope)
		err := codec.DecodeBytes(txEnv, request.Params)
		if err != nil {
			return nil, RPCErrorInvalidParams, err
		}
		receipt, err := service.Transactor().BroadcastTx(txEnv)
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return receipt, 0, nil
	}

	rpcServiceMap[GET_ACCOUNTS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &FilterListInput{}
		if len(request.Params) > 0 {
			err := codec.DecodeBytes(input, request.Params)
			if err != nil {
				return nil, RPCErrorInvalidParams, err
			}
		}
		filter, err := accountFilterFactory.NewFilter(input.Filters)
		if err != nil {
			return nil, RPCErrorInvalidParams, err
		}
		list, err := service.ListAccounts(func(account *account.Account) bool {
			return filter.Match(account)

		})
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return list, 0, nil
	}

	rpcServiceMap[GET_ACCOUNT] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &AddressInput{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, RPCErrorInvalidParams, err
		}

		acc, err := service.GetAccount(input.Address)
		if acc == nil || err != nil {
			return nil, RPCErrorInternalError, err
		}
		return acc, 0, nil
	}

	rpcServiceMap[GET_STORAGE] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &AddressInput{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, RPCErrorInvalidParams, err
		}
		storage, err := service.DumpStorage(input.Address)
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return storage, 0, nil
	}

	rpcServiceMap[GET_STORAGE_AT] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &StorageAtInput{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, RPCErrorInvalidParams, err
		}

		storageItem, err := service.GetStorage(input.Address, input.Key)
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return storageItem, 0, nil
	}

	rpcServiceMap[GET_STATUS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		status, err := service.Status()
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return status, 0, nil
	}

	rpcServiceMap[GET_LATEST_BLOCK] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		status, err := service.Status()
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		resultGetBlock, err := service.GetBlock(status.LatestBlockHeight)
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return resultGetBlock, 0, nil
	}

	rpcServiceMap[GET_BLOCKS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &BlocksInput{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, RPCErrorInvalidParams, err
		}
		blocks, err := service.ListBlocks(input.MinHeight, input.MaxHeight)
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return blocks, 0, err
	}

	rpcServiceMap[GET_BLOCK] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &BlockInput{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, RPCErrorInvalidParams, err
		}
		block, err := service.GetBlock(input.Height)
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return block, 0, nil
	}

	rpcServiceMap[GET_UNCONFIRMED_TXS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		transactions, err := service.ListUnconfirmedTxs(-1)
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return transactions, 0, nil
	}

	rpcServiceMap[GET_BLOCK_TXS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		input := &BlockInput{}
		err := codec.DecodeBytes(input, request.Params)
		if err != nil {
			return nil, RPCErrorInvalidParams, err
		}
		transactions, err := service.ListBlockTxs(input.Height)
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return transactions, 0, nil
	}

	rpcServiceMap[GET_CONSENSUS_STATE] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		consensusState, err := service.DumpConsensusState()
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return consensusState, 0, nil
	}

	rpcServiceMap[GET_VALIDATORS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		validators, err := service.ListValidators()
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return validators, 0, nil
	}

	rpcServiceMap[GET_NETWORK_INFO] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		info, err := service.NetInfo()
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return info, 0, nil
	}

	rpcServiceMap[GET_CHAIN_ID] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		chainId, err := service.ChainIdentifiers()
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return chainId, 0, nil
	}

	rpcServiceMap[GET_PEERS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		peers, err := service.Peers()
		if err != nil {
			return nil, RPCErrorInternalError, err
		}
		return peers, 0, nil
	}

	rpcServiceMap[GET_GENESIS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		genesis := service.Genesis()
		return genesis, 0, nil
	}
}
