package schedule_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/findbed/app/schedule"
	"github.com/findbed/app/schedule/mysqldb"
	"github.com/findbed/app/tests/helper"
	"github.com/imega/testhelpers/db"
	"github.com/imega/txwrapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RegisterLot(t *testing.T) {
	node := "xx"
	nodeID := schedule.CodeID{}
	copy(nodeID[:], node)

	txs := func(ctx context.Context, tx *sql.Tx) error {
		err := helper.CreateTimeslotTable(ctx, tx, node)
		require.NoError(t, err)

		return nil
	}

	curDB, close, err := db.Create("", txwrapper.TxFunc(txs))
	require.NoError(t, err)
	defer close()

	now := time.Now()
	scheduler := schedule.New(now, mysqldb.New(curDB))
	ctx := context.Background()

	code := schedule.CodeID{}
	copy(code[:], gofakeit.CountryAbr())

	query := schedule.Query{
		NodeID: nodeID,
		Region: code,
		From:   now.AddDate(0, 0, 1),
		To:     now.AddDate(0, 0, 3),
	}

	slots, err := scheduler.Search(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, []schedule.TimeSlot{}, slots)

	slot := schedule.TimeSlot{
		NodeID:      nodeID,
		HousingID:   schedule.LongID(gofakeit.Number(999_999, 999_999_999)),
		LotID:       schedule.LongID(gofakeit.Number(999_999, 999_999_999)),
		Region:      query.Region,
		Area:        schedule.ID(gofakeit.Number(1, 65_535)),
		Locality:    schedule.ID(gofakeit.Number(1, 65_535)),
		Sublocality: schedule.ID(gofakeit.Number(1, 65_535)),
	}
	err = scheduler.RegisterLot(ctx, slot)
	assert.NoError(t, err)

	slots, err = scheduler.Search(ctx, query)
	assert.NoError(t, err)

	slot.NodeID = [2]byte{}
	assert.Equal(t, slot, slots[0])
}

func Test_Book(t *testing.T) {
	timeslot := newTimeslot()

	txs := func(ctx context.Context, tx *sql.Tx) error {
		err := helper.CreateTimeslotTable(ctx, tx, string(timeslot.NodeID[:]))
		require.NoError(t, err)

		return nil
	}

	curDB, close, err := db.Create("", txwrapper.TxFunc(txs))
	require.NoError(t, err)
	defer close()

	now := time.Now()
	scheduler := schedule.New(now, mysqldb.New(curDB))
	ctx := context.Background()

	err = scheduler.RegisterLot(ctx, timeslot)
	assert.NoError(t, err)

	from := now.AddDate(0, 0, 1).Add(1 * time.Hour)
	to := now.AddDate(0, 0, 3).Add(5 * time.Hour)

	query := schedule.Query{
		NodeID: timeslot.NodeID,
		Region: timeslot.Region,
		From:   from,
		To:     to,
	}

	slots, err := scheduler.Search(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, len(slots), 1)

	timeslot.StartAt = from
	timeslot.EndAt = to

	err = scheduler.Book(ctx, timeslot)
	assert.NoError(t, err)

	t.Run("trying to book the same slot", func(t *testing.T) {
		slots, err = scheduler.Search(ctx, query)
		assert.NoError(t, err)
		assert.Equal(t, len(slots), 0)
	})

	t.Run(
		"unable to book a slot an hour early but ends at the same time",
		func(t *testing.T) {
			slots, err = scheduler.Search(ctx, schedule.Query{
				NodeID: timeslot.NodeID,
				Region: timeslot.Region,
				From:   from.Add(-1 * time.Hour),
				To:     to,
			})

			assert.NoError(t, err)
			assert.Equal(t, len(slots), 0)
		},
	)

	t.Run(
		"unable to book a slot an hour later but start at the same time",
		func(t *testing.T) {
			slots, err = scheduler.Search(ctx, schedule.Query{
				NodeID: timeslot.NodeID,
				Region: timeslot.Region,
				From:   from,
				To:     to.Add(time.Hour),
			})

			assert.NoError(t, err)
			assert.Equal(t, len(slots), 0)
		},
	)

	t.Run(
		"unable to book a slot an hour early and ends an hour later",
		func(t *testing.T) {
			slots, err = scheduler.Search(ctx, schedule.Query{
				NodeID: timeslot.NodeID,
				Region: timeslot.Region,
				From:   from.Add(-1 * time.Hour),
				To:     to.Add(time.Hour),
			})

			assert.NoError(t, err)
			assert.Equal(t, len(slots), 0)
		},
	)

	t.Run(
		"it's possible to book a slot an hour later and ends an two hour later",
		func(t *testing.T) {
			slots, err = scheduler.Search(ctx, schedule.Query{
				NodeID: timeslot.NodeID,
				Region: timeslot.Region,
				From:   to.Add(time.Hour),
				To:     to.Add(2 * time.Hour),
			})

			assert.NoError(t, err)
			assert.Equal(t, len(slots), 1)
		},
	)

	t.Run("to book a slot an hour early", func(t *testing.T) {
		slot := timeslot
		slot.StartAt = timeslot.StartAt.Add(-1 * time.Hour)
		slot.EndAt = timeslot.StartAt

		err = scheduler.Book(ctx, slot)
		assert.NoError(t, err)

		slots, err = scheduler.Search(ctx, schedule.Query{
			NodeID: slot.NodeID,
			Region: slot.Region,
			From:   slot.StartAt,
			To:     slot.EndAt,
		})
		assert.NoError(t, err)
		assert.Equal(t, len(slots), 0)
	})

	t.Run("to book a slot an hour later", func(t *testing.T) {
		slot := timeslot
		slot.StartAt = timeslot.EndAt
		slot.EndAt = timeslot.EndAt.Add(1 * time.Hour)

		err = scheduler.Book(ctx, slot)
		assert.NoError(t, err)

		slots, err = scheduler.Search(ctx, schedule.Query{
			NodeID: slot.NodeID,
			Region: slot.Region,
			From:   slot.StartAt,
			To:     slot.EndAt,
		})
		assert.NoError(t, err)
		assert.Equal(t, len(slots), 0)
	})
}

func Test_Book_with_two_lots(t *testing.T) {
	timeslot := newTimeslot()
	timeslotSecond := timeslot
	timeslotSecond.LotID += timeslotSecond.LotID

	txs := func(ctx context.Context, tx *sql.Tx) error {
		err := helper.CreateTimeslotTable(ctx, tx, string(timeslot.NodeID[:]))
		require.NoError(t, err)

		return nil
	}

	curDB, close, err := db.Create("", txwrapper.TxFunc(txs))
	require.NoError(t, err)
	defer close()

	now := time.Now()
	scheduler := schedule.New(now, mysqldb.New(curDB))
	ctx := context.Background()

	err = scheduler.RegisterLot(ctx, timeslot)
	assert.NoError(t, err)

	err = scheduler.RegisterLot(ctx, timeslotSecond)
	assert.NoError(t, err)

	from := now.AddDate(0, 0, 1)
	to := now.AddDate(0, 0, 3)

	query := schedule.Query{
		NodeID: timeslot.NodeID,
		Region: timeslot.Region,
		From:   from,
		To:     to,
	}

	slots, err := scheduler.Search(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, len(slots), 2)

	timeslot.StartAt = from
	timeslot.EndAt = to

	err = scheduler.Book(ctx, timeslot)
	assert.NoError(t, err)

	slots, err = scheduler.Search(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, len(slots), 1)
}

func Test_Cancel(t *testing.T) {
	timeslot := newTimeslot()

	txs := func(ctx context.Context, tx *sql.Tx) error {
		err := helper.CreateTimeslotTable(ctx, tx, string(timeslot.NodeID[:]))
		require.NoError(t, err)

		return nil
	}

	curDB, close, err := db.Create("", txwrapper.TxFunc(txs))
	require.NoError(t, err)
	defer close()

	now := time.Now()
	scheduler := schedule.New(now, mysqldb.New(curDB))
	ctx := context.Background()

	err = scheduler.RegisterLot(ctx, timeslot)
	assert.NoError(t, err)

	from := now.AddDate(0, 0, 1).Add(1 * time.Hour)
	to := now.AddDate(0, 0, 3).Add(5 * time.Hour)

	query := schedule.Query{
		NodeID: timeslot.NodeID,
		Region: timeslot.Region,
		From:   from,
		To:     to,
	}

	slots, err := scheduler.Search(ctx, query)
	assert.NoError(t, err)
	assert.Equal(t, len(slots), 1)

	timeslot.StartAt = from
	timeslot.EndAt = to

	err = scheduler.Book(ctx, timeslot)
	assert.NoError(t, err)

	t.Run("to cancel a slot an hour early", func(t *testing.T) {
		slot := timeslot
		slot.StartAt = timeslot.StartAt
		slot.EndAt = timeslot.StartAt.Add(time.Hour)

		err = scheduler.Cancel(ctx, slot)
		assert.NoError(t, err)

		slots, err = scheduler.Search(ctx, schedule.Query{
			NodeID: slot.NodeID,
			Region: slot.Region,
			From:   slot.StartAt,
			To:     slot.EndAt,
		})
		assert.NoError(t, err)
		assert.Equal(t, len(slots), 1)
	})

	t.Run("to cancel a slot an hour later", func(t *testing.T) {
		slot := timeslot
		slot.StartAt = timeslot.EndAt.Add(-time.Hour)
		slot.EndAt = timeslot.EndAt

		err = scheduler.Cancel(ctx, slot)
		assert.NoError(t, err)

		slots, err = scheduler.Search(ctx, schedule.Query{
			NodeID: slot.NodeID,
			Region: slot.Region,
			From:   slot.StartAt,
			To:     slot.EndAt,
		})
		assert.NoError(t, err)
		assert.Equal(t, len(slots), 1)
	})

	t.Run("to cancel the middle slot", func(t *testing.T) {
		slot := timeslot
		slot.StartAt = timeslot.StartAt.Add(5 * time.Hour)
		slot.EndAt = timeslot.EndAt.Add(-5 * time.Hour)

		err = scheduler.Cancel(ctx, slot)
		assert.NoError(t, err)

		slots, err = scheduler.Search(ctx, schedule.Query{
			NodeID: slot.NodeID,
			Region: slot.Region,
			From:   slot.StartAt,
			To:     slot.EndAt,
		})
		assert.NoError(t, err)
		assert.Equal(t, len(slots), 1)
	})
}

func newTimeslot() schedule.TimeSlot {
	code := schedule.CodeID{}
	copy(code[:], gofakeit.CountryAbr())

	return schedule.TimeSlot{
		NodeID:      code,
		HousingID:   schedule.LongID(gofakeit.Number(999_999, 999_999_999)),
		LotID:       schedule.LongID(gofakeit.Number(999_999, 999_999_999)),
		Region:      code,
		Area:        schedule.ID(gofakeit.Number(1, 65_535)),
		Locality:    schedule.ID(gofakeit.Number(1, 65_535)),
		Sublocality: schedule.ID(gofakeit.Number(1, 65_535)),
	}
}
