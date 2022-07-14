/*
* Copyright 2022-present Open Networking Foundation

* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at

* http://www.apache.org/licenses/LICENSE-2.0

* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"github.com/opencord/voltha-protos/v5/go/voltha"
)

const (
	volthaEventsTopic    = "voltha.events"
	kafkaBackoffInterval = time.Second * 10
)

//Used to listen for events coming from VOLTHA
type KafkaConsumer struct {
	address           string
	config            *sarama.Config
	client            sarama.Client
	consumer          sarama.Consumer
	partitionConsumer sarama.PartitionConsumer
	highwater         int64
}

//Creates a sarama client with the specified address
func NewKafkaConsumer(clusterAddress string) *KafkaConsumer {
	c := KafkaConsumer{address: clusterAddress}
	c.config = sarama.NewConfig()
	c.config.ClientID = "bbf-adapter-consumer"
	c.config.Consumer.Return.Errors = true
	c.config.Consumer.Offsets.Initial = sarama.OffsetNewest
	c.config.Version = sarama.V1_0_0_0

	return &c
}

//Starts consuming new messages on the voltha events topic, executing the provided callback on each event
func (c *KafkaConsumer) Start(ctx context.Context, eventCallback func(context.Context, *voltha.Event)) error {
	var err error

	for {
		if c.client, err = sarama.NewClient([]string{c.address}, c.config); err == nil {
			logger.Debug(ctx, "kafka-client-created")
			break
		} else {
			logger.Warnw(ctx, "kafka-not-reachable", log.Fields{
				"err": err,
			})
		}

		//Wait a bit before trying again
		select {
		case <-ctx.Done():
			return fmt.Errorf("kafka-client-creation-stopped-due-to-context-done")
		case <-time.After(kafkaBackoffInterval):
			continue
		}
	}

	c.consumer, err = sarama.NewConsumerFromClient(c.client)
	if err != nil {
		return err
	}

	partitions, _ := c.consumer.Partitions(volthaEventsTopic)

	// TODO: Add support for multiple partitions
	if len(partitions) > 1 {
		logger.Warnw(ctx, "only-listening-one-partition", log.Fields{
			"topic":         volthaEventsTopic,
			"partitionsNum": len(partitions),
		})
	}

	hw, err := c.client.GetOffset(volthaEventsTopic, partitions[0], sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("cannot-get-highwater: %v", err)
	}
	c.highwater = hw

	c.partitionConsumer, err = c.consumer.ConsumePartition(volthaEventsTopic, partitions[0], sarama.OffsetOldest)
	if nil != err {
		return fmt.Errorf("Error in consume(): Topic %v Partitions: %v", volthaEventsTopic, partitions)
	}

	//Start consuming the event topic in a goroutine
	logger.Debugw(ctx, "start-consuming-kafka-topic", log.Fields{"topic": volthaEventsTopic})
	go func(topic string, pConsumer sarama.PartitionConsumer) {
		for {
			select {
			case <-ctx.Done():
				logger.Info(ctx, "stopped-listening-for-events-due-to-context-done")
				return
			case err := <-pConsumer.Errors():
				logger.Errorw(ctx, "kafka-consumer-error", log.Fields{
					"err":       err.Error(),
					"topic":     err.Topic,
					"partition": err.Partition,
				})
			case msg := <-pConsumer.Messages():
				if msg.Offset <= c.highwater {
					continue
				}

				//Unmarshal the content of the message to a voltha Event protobuf message
				event := &voltha.Event{}
				if err := proto.Unmarshal(msg.Value, event); err != nil {
					logger.Errorw(ctx, "error-unmarshalling-kafka-event", log.Fields{"err": err})
					continue
				}

				eventCallback(ctx, event)
			}
		}
	}(volthaEventsTopic, c.partitionConsumer)

	return nil
}

//Closes the sarama client and all consumers
func (c *KafkaConsumer) Stop() error {
	if err := c.partitionConsumer.Close(); err != nil {
		return err
	}

	if err := c.consumer.Close(); err != nil {
		return err
	}

	if err := c.client.Close(); err != nil {
		return err
	}

	return nil
}
