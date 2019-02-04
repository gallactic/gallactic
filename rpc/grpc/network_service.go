package grpc

import (
	"context"
	"github.com/gallactic/gallactic/core/blockchain"
	"github.com/gallactic/gallactic/core/consensus/tendermint/p2p"
	"github.com/gallactic/gallactic/core/consensus/tendermint/query"
	pb "github.com/gallactic/gallactic/rpc/grpc/proto3"
	net "github.com/tendermint/tendermint/p2p"
)

type networkService struct {
	nodeview   *query.NodeView
	blockchain *blockchain.Blockchain
}

var _ pb.NetworkServer = &networkService{}

func NewNetworkService(blockchain *blockchain.Blockchain, nView *query.NodeView) *networkService {
	return &networkService{
		blockchain: blockchain,
		nodeview:   nView,
	}
}

//Network info
func (s *networkService) GetNetworkInfo(context.Context, *pb.Empty1) (*pb.NetInfoResponse, error) {
	var contexts context.Context
	peers, err := s.GetPeers(contexts, nil)
	if err != nil {
		return nil, err
	}
	return &pb.NetInfoResponse{
		Peers: peers.Peer,
	}, nil

}

//Get list of Peers
func (ns *networkService) GetPeers(context.Context, *pb.Empty1) (*pb.PeerResponse, error) {
	peers := make([]*pb.Peer, ns.nodeview.Peers().Size())
	for i, peer := range ns.nodeview.Peers().List() {
		ni := new(p2p.GNodeInfo)
		tmni, _ := peer.NodeInfo().(*net.DefaultNodeInfo)
		ni.ID_ = tmni.ID_
		ni.Network = tmni.Network
		ni.ProtocolVersion = tmni.ProtocolVersion
		ni.Version = tmni.Version
		ni.Channels = tmni.Channels
		ni.ListenAddr = tmni.ListenAddr
		ni.Moniker = tmni.Moniker
		peers[i] = &pb.Peer{
			NodeInfo:   *ni,
			IsOutbound: peer.IsOutbound(),
		}
	}
	return &pb.PeerResponse{
		Peer: peers,
	}, nil
}
