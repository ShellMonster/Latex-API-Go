package cache

// Stats 返回缓存当前关键指标，用于健康检查等场景
func (m *Manager) Stats() Stats {
	return Stats{
		LocalEntries: m.local.Len(),
		HitsLocal:    m.hitsLocal.Load(),
		HitsRedis:    m.hitsRedis.Load(),
		Misses:       m.misses.Load(),
		RedisEnabled: m.redisEnabled,
		RedisAlive:   m.redisAlive.Load(),
	}
}

func (m *Manager) markRedisAlive(alive bool) {
	if !m.redisEnabled {
		m.redisAlive.Store(false)
		return
	}

	prev := m.redisAlive.Load()
	if alive {
		m.redisAlive.Store(true)
		if !prev {
			m.logger.Info("Redis 恢复可用")
		}
		return
	}

	m.redisAlive.Store(false)
	if prev {
		m.logger.Warn("Redis 不可用，降级为仅 BigCache")
	}
}
