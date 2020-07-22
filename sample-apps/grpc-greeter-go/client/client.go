/*
 *
 * Original work Copyright 2015 gRPC authors.
 * Modified work Copyright 2020 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*
 * Changes:
 * 2019-06-24: Log the SayHello response header called hostname
 */

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
)

const (
	defaultName = "world"
	timeout     = 5 * time.Second
)

func main() {
	address := flag.String("address", "localhost:50051", "host:port of gRPC server")
	cert := flag.String("cert", "/data/cert.pem", "path to TLS certificate")
	repeat := flag.Int("repeat", 9, "number of unary gRPC requests to send")
	insecure := flag.Bool("insecure", false, "connect without TLS")
	flag.Parse()

	// Set up a connection to the server.
	resolver.SetDefaultScheme("dns")
	var conn *grpc.ClientConn
	var err error
	if *insecure {
		conn, err = grpc.Dial(*address, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
	} else {
		tc, err := credentials.NewClientTLSFromFile(*cert, "")
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		conn, err = grpc.Dial(*address, grpc.WithTransportCredentials(tc), grpc.WithBalancerName(roundrobin.Name))
	}
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Determine name to send to server.
	name := defaultName
	nonFlagArgs := make([]string, 0)
	for _, arg := range os.Args {
		if !strings.HasPrefix(arg, "--") {
			nonFlagArgs = append(nonFlagArgs, arg)
		}
	}
	if len(nonFlagArgs) > 1 {
		name = nonFlagArgs[1]
	}

	for i := 0; i < *repeat; i++ {
		var header metadata.MD
		r, err := c.SayHello(context.Background(), &pb.HelloRequest{Name: name}, grpc.Header(&header))
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		hostname := "unknown"
		// [START istio_sample_apps_grpc_greeter_go_client_hostname]
		if len(header["hostname"]) > 0 {
			hostname = header["hostname"][0]
		}
		log.Printf("%s from %s", r.Message, hostname)
		time.Sleep(time.Duration(1) * time.Second)
		// [END istio_sample_apps_grpc_greeter_go_client_hostname]
	}
}
