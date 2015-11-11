package urknall

import (
	"time"

	"github.com/megamsys/urknall/pubsub"
)

func message(key string, hostname string, taskName string) (msg *pubsub.Message) {
	return &pubsub.Message{Key: key, StartedAt: time.Now(), Hostname: hostname, TaskName: taskName}
}
