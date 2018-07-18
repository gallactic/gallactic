package rpc

// Used to handle requests. interface{} param is a wildcard used for example with socket events.
type RequestHandlerFunc func(request *RPCRequest, requester interface{}) (interface{}, int, error)

func GetMethods(codec Codec, service *Service) map[string]RequestHandlerFunc {

	rpcMap := make(map[string]RequestHandlerFunc)
	loadGallacticMethods(codec, service, rpcMap)

	return rpcMap

}
