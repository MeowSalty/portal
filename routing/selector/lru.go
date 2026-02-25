package selector

import (
	"math"
	"time"

	"github.com/MeowSalty/portal/errors"
)

const (
	defaultPlatformWeight = 1.0
	defaultModelWeight    = 1.0
	defaultKeyWeight      = 1.0
)

func init() {
	Register(LRUSelector, NewLRUSelector)
}

// lruSelector 实现多维 LRU 调度策略
// 基于最近尝试时间，综合平台/模型/密钥维度计算分数。
type lruSelector struct {
	platformWeight float64
	modelWeight    float64
	keyWeight      float64
}

// NewLRUSelector 创建一个新的多维 LRU 选择器实例
func NewLRUSelector() Selector {
	return &lruSelector{
		platformWeight: defaultPlatformWeight,
		modelWeight:    defaultModelWeight,
		keyWeight:      defaultKeyWeight,
	}
}

// Select 从给定的通道列表中选择评分最高的通道
//
// 评分公式：score = wP*agePlatform + wM*ageModel + wK*ageKey
// 平局处理顺序：ageModel -> agePlatform -> ageKey -> stable ID
func (s *lruSelector) Select(channels []ChannelInfo) (string, error) {
	if len(channels) == 0 {
		return "", errors.New(errors.ErrCodeInvalidArgument, "通道列表不能为空")
	}

	now := time.Now()
	bestIndex := 0
	bestScore := math.Inf(-1)
	bestAgeModel := time.Duration(-1)
	bestAgePlatform := time.Duration(-1)
	bestAgeKey := time.Duration(-1)
	bestID := channels[0].ID

	for i, ch := range channels {
		agePlatform := now.Sub(ch.LastTryPlatform)
		ageModel := now.Sub(ch.LastTryModel)
		ageKey := now.Sub(ch.LastTryKey)

		score := s.platformWeight*agePlatform.Seconds() +
			s.modelWeight*ageModel.Seconds() +
			s.keyWeight*ageKey.Seconds()

		if score > bestScore {
			bestIndex = i
			bestScore = score
			bestAgeModel = ageModel
			bestAgePlatform = agePlatform
			bestAgeKey = ageKey
			bestID = ch.ID
			continue
		}

		if score == bestScore {
			if ageModel > bestAgeModel {
				bestIndex = i
				bestAgeModel = ageModel
				bestAgePlatform = agePlatform
				bestAgeKey = ageKey
				bestID = ch.ID
				continue
			}
			if ageModel == bestAgeModel {
				if agePlatform > bestAgePlatform {
					bestIndex = i
					bestAgePlatform = agePlatform
					bestAgeKey = ageKey
					bestID = ch.ID
					continue
				}
				if agePlatform == bestAgePlatform {
					if ageKey > bestAgeKey {
						bestIndex = i
						bestAgeKey = ageKey
						bestID = ch.ID
						continue
					}
					if ageKey == bestAgeKey && ch.ID < bestID {
						bestIndex = i
						bestID = ch.ID
					}
				}
			}
		}
	}

	return channels[bestIndex].ID, nil
}

// Name 返回选择器的名称
func (s *lruSelector) Name() string {
	return "LRU"
}
