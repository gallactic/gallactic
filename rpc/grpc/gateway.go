package grpc

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"

	pb "github.com/gallactic/gallactic/rpc/grpc/proto3"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

func (s *Server) StartGateway(ctx context.Context, grpcAddr, gatewayAddr string) error {

	getEndpoint := flag.String("get", grpcAddr, "endpoint of Gallactic(GET)")

	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONBuiltin{}))
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := pb.RegisterBlockChainHandlerFromEndpoint(ctx, mux, *getEndpoint, opts); err != nil {
		return err
	}

	if err := pb.RegisterNetworkHandlerFromEndpoint(ctx, mux, *getEndpoint, opts); err != nil {
		return err
	}

	if err := pb.RegisterTransactionHandlerFromEndpoint(ctx, mux, *getEndpoint, opts); err != nil {
		return err
	}

	s.handleEntryPoint(mux, gatewayAddr)

	go http.ListenAndServe(gatewayAddr, mux) /// TODO: check error with channels

	return nil
}

func (s *Server) handleEntryPoint(mux *runtime.ServeMux, addr string) {
	entryPoint := runtime.MustPattern(runtime.NewPattern(1, []int{2, 0}, []string{""}, ""))
	// grpc endpoints
	mux.Handle("GET", entryPoint, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)
		buf := new(bytes.Buffer)
		buf.WriteString("<html><body>")

		for k, v := range s.GetServiceInfo() {
			buf.WriteString(fmt.Sprintf("<br>%s endpoints:<br>", k))

			for _, m := range v.Methods {
				/// Show only get methods
				if strings.HasPrefix(m.Name, "Get") {
					link := fmt.Sprintf("//%s/%s", addr, m.Name[3:])
					buf.WriteString(fmt.Sprintf("<a href=\"%s\">%s</a></br>", link, link))
				}
			}
		}
		buf.WriteString("</body></html>")

		w.Write(buf.Bytes())
	})
}
