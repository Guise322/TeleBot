package broker

import (
	"context"
	"fmt"
	"time"

	"github.com/Guise322/TeleBot/server/internal/common"
	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
)

func (kr KafkaConsumer) Commit(ctx context.Context, msgUuid uuid.UUID) error {
	kr.mu.Lock()

	kr.removeOldMessages()
	kr.removeOldOffcets()
	uncomMsg, ok := kr.uncommittedMessages[msgUuid]
	if !ok {
		kr.mu.Unlock()

		return fmt.Errorf("no key %v between the processing messages", msgUuid)
	}

	kr.mu.Unlock()

	err := kr.commitMesWithRetries(ctx, uncomMsg.msg)
	if err != nil {
		return fmt.Errorf("can not commit a message in the broker: %w", err)

	}

	kr.mu.Lock()
	kr.offcets[uncomMsg.msg.Partition] = offcetWithTimeStamp{value: uncomMsg.msg.Offset, timeStamp: time.Now()}
	delete(kr.uncommittedMessages, msgUuid)
	kr.mu.Unlock()

	return nil
}

func (kr *KafkaConsumer) commitMesWithRetries(ctx context.Context, msg kafka.Message) error {
	act := func(ctx context.Context) error {
		lastOffcetWithTimeStamp, ok := kr.offcets[msg.Partition]
		if ok && lastOffcetWithTimeStamp.value > msg.Offset {
			return nil
		}

		return kr.reader.CommitMessages(ctx, msg)
	}

	return common.ExecuteWithRetries(ctx, act)
}

func (kr KafkaConsumer) removeOldMessages() {
	now := time.Now()

	for msgUuid, procMsg := range kr.uncommittedMessages {
		if procMsg.timeStamp.Add(48 * time.Hour).Before(now) {
			delete(kr.uncommittedMessages, msgUuid)
		}
	}
}

func (kr KafkaConsumer) removeOldOffcets() {
	now := time.Now()

	for part, offcet := range kr.offcets {
		if offcet.timeStamp.Add(48 * time.Hour).Before(now) {
			delete(kr.offcets, part)
		}
	}
}
