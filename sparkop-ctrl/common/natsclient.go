/*
Copyright 2020

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package common


import (
    "context"
    //"strings"
    //"github.com/golang/glog"
    log "github.com/sirupsen/logrus"

    //"encoding/json"
    //"log"
    //"runtime"
    //"time"

    //stan "github.com/nats-io/go-nats-streaming"
    stan "github.com/nats-io/stan.go"

    //"github.com/shijuvar/gokit/examples/nats-streaming/pb"
)

const (
    NatsSUrl  = "nats://192.168.42.150:4222"
    NatsClusterID = "test-cluster"
    NatsChannelTraefik   = "notify-traefik"
    NatsDurableID = "sparkapp-service-durable"
)

type NatsStreamingClient struct {
    natsConn stan.Conn
}

func NewNatsStreamingClient() (*NatsStreamingClient) {
    return &NatsStreamingClient{
    }
}

func (r *NatsStreamingClient) Start(stopCh <-chan struct{}, natssurl, clietId string) {
    log.Info("nats streaming client start ... ")
    sc, err := stan.Connect(
        NatsClusterID,
        clietId,
        stan.NatsURL(natssurl),
    )

    if err != nil {
        log.Fatal(err)
    }

    r.natsConn = sc
}

func (r *NatsStreamingClient) Stop(ctx context.Context) {
    if (r.natsConn != nil) {
        r.natsConn.Close()
        r.natsConn = nil
    }
}

func (r *NatsStreamingClient) GetConn() stan.Conn {
    return r.natsConn
}

func (r *NatsStreamingClient) SendMessage(channel string, msg string) {
    ackHandler := func(ackedNuid string, err error) {
        if err != nil {
            log.Errorf("Warning: error publishing msg id %s: %v\n", ackedNuid, err.Error())
        } else {
            log.Infof("Received ack for msg id %s\n", ackedNuid)
        }
    }

    log.Infof("nats streaming client send : %s ", msg)

    // returns immediately
    nuid, err := r.natsConn.PublishAsync(channel, []byte(msg), ackHandler) 
    if err != nil {
        log.Errorf("Error publishing msg %s: %v\n", nuid, err.Error())
    }
}
