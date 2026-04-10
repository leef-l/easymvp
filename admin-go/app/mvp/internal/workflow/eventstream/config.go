package eventstream

const (
	defaultStreamName   = "easymvp:workflow:task-events"
	defaultConsumerName = "workflow-event-consumer"
	defaultGroupName    = "easymvp-workflow"
)

// Config 定义事件流配置。
type Config struct {
	Enabled            bool
	StreamName         string
	ConsumerGroup      string
	ConsumerName       string
	BlockMS            int
	ReclaimIdleSeconds int
	RedisRequired      bool
	ConsumerEnabled    bool
	ReadCount          int
	ReclaimCount       int
}

// Normalize 补齐默认值。
func (c Config) Normalize() Config {
	if c.StreamName == "" {
		c.StreamName = defaultStreamName
	}
	if c.ConsumerGroup == "" {
		c.ConsumerGroup = defaultGroupName
	}
	if c.ConsumerName == "" {
		c.ConsumerName = defaultConsumerName
	}
	if c.BlockMS <= 0 {
		c.BlockMS = 5000
	}
	if c.ReclaimIdleSeconds <= 0 {
		c.ReclaimIdleSeconds = 60
	}
	if c.ReadCount <= 0 {
		c.ReadCount = 20
	}
	if c.ReclaimCount <= 0 {
		c.ReclaimCount = 20
	}
	return c
}
