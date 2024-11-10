package event

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {

	k1 := assetMsg{
		assetId: 1,
		isSell:  true,
	}
	k2 := assetMsg{
		assetId: 2,
		isSell:  false,
	}

	assetMsgCache[k1] = &assetMsgSentInfo{
		isMsgSent: true,
		price:     1000,
		sentTime:  time.Now().Add(-1 * 2 * time.Hour),
	}

	assetMsgCache[k2] = &assetMsgSentInfo{
		isMsgSent: true,
		price:     1000,
		sentTime:  time.Now().Add(-1 * 25 * time.Hour),
	}

	t.Run("Valid Cache", func(t *testing.T) {
		b := hasMsgCache(1, true, 1000)
		if !b {
			t.Error(b)
		}
	})

	t.Run("No Cache", func(t *testing.T) {
		b := hasMsgCache(1, false, 1000)
		if b {
			t.Error(b)
		}
	})

	t.Run("Invalid Cache", func(t *testing.T) {
		b := hasMsgCache(2, false, 1000)
		if b {
			t.Error(b)
		}
	})

	t.Run("Set Cache", func(t *testing.T) {
		b := hasMsgCache(3, true, 1000)
		if b {
			t.Error(b)
		}

		setMsgCache(3, true, 1000)

		b = hasMsgCache(3, true, 1000)
		if !b {
			t.Error(b)
		}
	})

	t.Run("Daily Cache", func(t *testing.T) {
		b := hasDailyCache()
		assert.Equal(t, b, false)
		t.Log(time.Now().Day())
	})
}
