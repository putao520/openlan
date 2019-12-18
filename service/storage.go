package service

import (
	"fmt"
	"github.com/lightstar-dev/openlan-go/libol"
	"github.com/lightstar-dev/openlan-go/models"
	"github.com/lightstar-dev/openlan-go/point"
	"strings"
)

type storageService struct {
	redis       *libol.RedisCli
}

var StorageService = &storageService{
}

func (s *storageService) Open(addr string, auth string, db int) *libol.RedisCli {
	if s.redis == nil {
		s.redis = libol.NewRedisCli(addr, auth, db)
		if err := s.redis.Open(); err != nil {
			libol.Error("storageService.Open: %s", err)
		}
	}

	return s.redis
}

func (s *storageService) Redis() *libol.RedisCli {
	return s.redis
}

func (s *storageService) RedisId(prefix string, table string, key string) string {
	if prefix == "" {
		prefix = "default"
	}

	index := strings.Split(prefix, ":")
	wid := index[len(index)-1]
	kid := strings.Replace(key, ":", "-", -1)
	return fmt.Sprintf("%s:%s:%s", wid, table, kid)
}

func (s *storageService) SavePoint(prefix string, m *models.Point, isAdd bool) {
	key := s.RedisId(prefix, "point", m.Client.Addr)
	value := map[string]interface{}{
		"remote":  m.Client.String(),
		"newTime": m.Client.NewTime,
		"device":  m.Device.Name(),
		"active":  isAdd,
	}

	if r := s.Redis(); r != nil {
		if err := r.HMSet(key, value); err != nil {
			libol.Error("storageService.SavePoint %s", err)
		}
	}
}

func (s *storageService) SaveLink(prefix string, link *point.Point, isAdd bool) {
	key := s.RedisId(prefix, "link", link.Addr())
	value := map[string]interface{}{
		"remote": link.Addr(),
		"upTime": link.UpTime(),
		"device": link.IfName(),
		"state":  link.State(),
		"isAddr": isAdd,
	}

	if r := s.Redis(); r != nil {
		if err := r.HMSet(key, value); err != nil {
			libol.Error("storageService.SaveLink %s", err)
		}
	}
}

func (s *storageService) SaveNeighbor(prefix string, n *models.Neighbor, isAdd bool) {
	key := s.RedisId(prefix, "neighbor", n.HwAddr.String())
	value := map[string]interface{}{
		"hwAddr":  n.HwAddr.String(),
		"ipAddr":  n.IpAddr.String(),
		"remote":  n.Client.String(),
		"newTime": n.NewTime,
		"hitTime": n.HitTime,
		"active":  isAdd,
	}

	if r := s.Redis(); r != nil {
		if err := r.HMSet(key, value); err != nil {
			libol.Error("storageService.SaveNeighbor %s", err)
		}
	}
}

func (s *storageService) SaveLine(prefix string, l *models.Line, isAdd bool) {
	key := s.RedisId(prefix, "line", l.String())
	value := map[string]interface{}{
		"ethernet":    fmt.Sprintf("0x%04x", l.EthType),
		"source":      l.IpSource.String(),
		"destination": l.IPDest.String(),
		"protocol":    fmt.Sprintf("0x%02x", l.IpProtocol),
		"port":        fmt.Sprintf("%d", l.PortDest),
	}

	if r := s.Redis(); r != nil {
		if err := r.HMSet(key, value); err != nil {
			libol.Error("storageService.SaveLine %s", err)
		}
	}
}