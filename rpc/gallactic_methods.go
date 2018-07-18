package rpc

import (
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/txs"
)

const (
	GALLACTIC = "burrow."

	GET_ACCOUNTS        = GALLACTIC + "getAccounts"
	GET_ACCOUNT         = GALLACTIC + "getAccount"
	GET_STORAGE         = GALLACTIC + "getStorage"
	GET_STORAGE_AT      = GALLACTIC + "getStorageAt"
	GET_BLOCKCHAIN_INFO = GALLACTIC + "getBlockchainInfo"
	GET_LATEST_BLOCK    = GALLACTIC + "getLatestBlock"
	GET_BLOCKS          = GALLACTIC + "getBlocks"
	GET_BLOCK           = GALLACTIC + "getBlock"
	GET_CONSENSUS_STATE = GALLACTIC + "getConsensusState"
	GET_VALIDATORS      = GALLACTIC + "getValidators"
	GET_NETWORK_INFO    = GALLACTIC + "getNetworkInfo"
	GET_CHAIN_ID        = GALLACTIC + "getChainId"
	GET_PEERS           = GALLACTIC + "getPeers"
	BROADCAST_TX        = GALLACTIC + "broadcastTx"
	GET_UNCONFIRMED_TXS = GALLACTIC + "getUnconfirmedTxs"
)

func loadGallacticMethods(codec Codec, service *Service, rpcMethodMap map[string]RequestHandlerFunc) {

	accountFilterFactory := NewAccountFilterFactory()

	rpcMethodMap[BROADCAST_TX] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {

		txEnv := new(txs.Envelope)
		err := codec.DecodeBytes(txEnv, request.Params)
		if err != nil {
			return nil, INVALID_PARAMS, err
		}
		receipt, err := service.Transactor().BroadcastTx(txEnv)
		if err != nil {
			return nil, INTERNAL_ERROR, err
		}
		return receipt, 0, nil
	}

	rpcMethodMap[GET_ACCOUNTS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		param := &FilterListParam{}
		if len(request.Params) > 0 {
			err := codec.DecodeBytes(param, request.Params)
			if err != nil {
				return nil, INVALID_PARAMS, err
			}
		}
		filter, err := accountFilterFactory.NewFilter(param.Filters)
		if err != nil {
			return nil, INVALID_PARAMS, err
		}
		list, err := service.ListAccounts(func(account *account.Account) bool {
			return filter.Match(account)

		})
		if err != nil {
			return nil, INTERNAL_ERROR, err
		}
		return list, 0, nil
	}

	rpcMethodMap[GET_ACCOUNT] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		param := &AddressParam{}
		err := codec.DecodeBytes(param, request.Params)
		if err != nil {
			return nil, INVALID_PARAMS, err
		}

		acc, err := service.GetAccount(param.Address)
		if acc == nil || err != nil {
			return nil, INTERNAL_ERROR, err
		}
		return acc, 0, nil
	}

	rpcMethodMap[GET_STORAGE] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		param := &AddressParam{}
		err := codec.DecodeBytes(param, request.Params)
		if err != nil {
			return nil, INVALID_PARAMS, err
		}
		storage, err := service.DumpStorage(param.Address)
		if err != nil {
			return nil, INTERNAL_ERROR, err
		}
		return storage, 0, nil
	}

	rpcMethodMap[GET_STORAGE_AT] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		param := &StorageAtParam{}
		err := codec.DecodeBytes(param, request.Params)
		if err != nil {
			return nil, INVALID_PARAMS, err
		}

		storageItem, err := service.GetStorage(param.Address, param.Key)
		if err != nil {
			return nil, INTERNAL_ERROR, err
		}
		return storageItem, 0, nil
	}

	rpcMethodMap[GET_BLOCKCHAIN_INFO] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		resultStatus, err := service.Status()
		if err != nil {
			return nil, INTERNAL_ERROR, err
		}
		return resultStatus, 0, nil
	}

	rpcMethodMap[GET_LATEST_BLOCK] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		stat, err := service.Status()
		if err != nil {
			return nil, INTERNAL_ERROR, err
		}
		resultGetBlock, err := service.GetBlock(stat.LatestBlockHeight)
		if err != nil {
			return nil, INTERNAL_ERROR, err
		}
		return resultGetBlock, 0, nil
	}

	rpcMethodMap[GET_BLOCKS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		param := &BlocksParam{}
		err := codec.DecodeBytes(param, request.Params)
		if err != nil {
			return nil, INVALID_PARAMS, err
		}
		blocks, err := service.ListBlocks(param.MinHeight, param.MaxHeight)
		if err != nil {
			return nil, INTERNAL_ERROR, err
		}
		return blocks, 0, err
	}

	rpcMethodMap[GET_BLOCK] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		param := &HeightParam{}
		err := codec.DecodeBytes(param, request.Params)
		if err != nil {
			return nil, INVALID_PARAMS, err
		}
		block, err := service.GetBlock(param.Height)
		if err != nil {
			return nil, INTERNAL_ERROR, err
		}
		return block, 0, nil
	}

	rpcMethodMap[GET_UNCONFIRMED_TXS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		trans, err := service.ListUnconfirmedTxs(-1)
		if err != nil {
			return nil, INTERNAL_ERROR, err
		}
		return trans, 0, nil
	}

	rpcMethodMap[GET_CONSENSUS_STATE] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		resultDumpConsensusState, err := service.DumpConsensusState()
		if err != nil {
			return nil, INTERNAL_ERROR, err
		}
		return resultDumpConsensusState, 0, nil
	}

	rpcMethodMap[GET_VALIDATORS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		resultListValidators, err := service.ListValidators()
		if err != nil {
			return nil, INTERNAL_ERROR, err
		}
		return resultListValidators, 0, nil
	}

	rpcMethodMap[GET_NETWORK_INFO] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		resultNetInfo, err := service.NetInfo()
		if err != nil {
			return nil, INTERNAL_ERROR, err
		}
		return resultNetInfo, 0, nil
	}

	rpcMethodMap[GET_CHAIN_ID] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		resultChainID, err := service.ChainIdentifiers()
		if err != nil {
			return nil, INTERNAL_ERROR, err
		}
		return resultChainID, 0, nil
	}

	rpcMethodMap[GET_PEERS] = func(request *RPCRequest, requester interface{}) (interface{}, int, error) {
		resultPeers, err := service.Peers()
		if err != nil {
			return nil, INTERNAL_ERROR, err
		}
		return resultPeers, 0, nil
	}
}
