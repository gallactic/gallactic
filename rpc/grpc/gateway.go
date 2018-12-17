package grpc

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net/http"
	"path"
	"strings"

	pb "github.com/gallactic/gallactic/rpc/grpc/proto3"
	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

//GRPC GATEWAY

const (
	grpcPort = "50051"
)

var (
	swaggerDir = flag.String("swagger_dir", "template", "path to the directory which contains swagger definitions")
)

func (s *Server) StartGateway(ctx context.Context, gatewayAddr, grpcAddr string) error {
	/*lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}*/

	getEndpoint := flag.String("get", gatewayAddr, "endpoint of Gallactic(GET)")
	postEndpoint := flag.String("post", gatewayAddr, "endpoint of Gallactic(POST)")

	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONBuiltin{}))
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := pb.RegisterBlockChainHandlerFromEndpoint(ctx, mux, *getEndpoint, opts); err != nil {
		return err
	}

	//// TODO: Please test POST requests....
	if err := pb.RegisterBlockChainHandlerFromEndpoint(ctx, mux, *postEndpoint, opts); err != nil {
		return err
	}

	go http.ListenAndServe(grpcAddr, mux) /// TODO: check error with channels

	return nil
}

//????????????????????????
/// TEST IT
func serveSwagger(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, ".swagger.json") {
		glog.Errorf("Swagger JSON not Found: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	glog.Infof("Serving %s", r.URL.Path)
	p := strings.TrimPrefix(r.URL.Path, "/swagger/")
	p = path.Join(*swaggerDir, p)
	WriteListOfEndpoints(w, r, r.URL.Path)
	http.ServeFile(w, r, p)

}

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	glog.Infof("preflight request for %s", r.URL.Path)
	fmt.Println("server-url", r.URL.Path)
	return
}

// ???????????????????
/// TEST IT
// allowCORS allows Cross Origin Resoruce Sharing from any origin.
// Don't do this without consideration in production systems. ??????????????????
func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

//???????????????????????????
/// TEST IT
// writes a list of available rpc endpoints as an html page
func WriteListOfEndpoints(w http.ResponseWriter, r *http.Request, url string) {
	fmt.Println("Writevalue", w)
	fmt.Println("httpvalue", r)
	fmt.Println("UrlPath", url)
	noArgNames := []string{}

	// for name, funcData := range funcmap {
	// 	if len(funcData.args) == 0 {
	// 		noArgNames = append(noArgNames, name)
	// 	} else {
	// 		argNames = append(argNames, name)
	// 	}
	// }
	// sort.Strings(noArgNames)
	// sort.Strings(argNames)
	buf := new(bytes.Buffer)
	buf.WriteString("<html><body>")
	buf.WriteString("<br>Available endpoints:<br>")
	buf.WriteString(fmt.Sprintf("<a href=\"%s\">%s</a></br>", r.Host, url))
	for _, name := range noArgNames {
		link := fmt.Sprintf("//%s/%s", r.Host, name)
		buf.WriteString(fmt.Sprintf("<a href=\"%s\">%s</a></br>", link, link))
	}

	// buf.WriteString("<br>Endpoints that require arguments:<br>")
	// for _, name := range argNames {
	// 	link := fmt.Sprintf("//%s/%s?", r.Host, name)
	// 	funcData := funcMap[name]
	// 	for i, argName := range funcData.argNames {
	// 		link += argName + "=_"
	// 		if i < len(funcData.argNames)-1 {
	// 			link += "&"
	// 		}
	// 	}
	// 	buf.WriteString(fmt.Sprintf("<a href=\"%s\">%s</a></br>", link, link))
	// }
	// buf.WriteString("</body></html>")
	// w.Header().Set("Content-Type", "text/html")
	// w.WriteHeader(200)
	// w.Write(buf.Bytes()) // nolint: errcheck
}
