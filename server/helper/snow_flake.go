package helper

import (
	"sync"
	"time"
)

const (
	//SnowFlake 雪花算法
	StartTimeStamp = int64(1483228800000) //开始时间截 (2017-01-01)
	MachineIdBits  = uint(10)             //机器id所占的位数
	SequenceBits   = uint(12)             //序列所占的位数
	//MachineIdMax   = int64(-1 ^ (-1 << MachineIdBits)) //支持的最大机器id数量
	SequenceMask   = int64(-1 ^ (-1 << SequenceBits)) //
	MachineIdShift = SequenceBits                     //机器id左移位数
	TimestampShift = SequenceBits + MachineIdBits     //时间戳左移位数
)

// 创建一个雪花算法生成器(生成工厂)
// SnowFlakeMachineId: 2
// 如果本项目同时部署在多台机器，并且需要同时使用该算法，
// 请为每一台机器设置不同的ID，区间范围: [0,1023]
func CreateSnowflakeFactory() InterfaceSnowFlake {
	return &snowflake{
		timestamp: 0,
		machineId: 2,
		sequence:  0,
	}
}

type snowflake struct {
	sync.Mutex
	timestamp int64
	machineId int64
	sequence  int64
}

// 生成分布式ID
func (s *snowflake) GetId() int64 {
	s.Lock()
	defer func() {
		s.Unlock()
	}()
	now := time.Now().UnixNano() / 1e6
	if s.timestamp == now {
		s.sequence = (s.sequence + 1) & SequenceMask
		if s.sequence == 0 {
			for now <= s.timestamp {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		s.sequence = 0
	}
	s.timestamp = now
	r := (now-StartTimeStamp)<<TimestampShift | (s.machineId << MachineIdShift) | (s.sequence)
	return r
}
