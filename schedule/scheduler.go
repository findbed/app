// Copyright © 2022 Dmitry Stoletov <info@imega.ru>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package schedule

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/findbed/app/schedule/mysqldb"
)

type Scheduler struct {
	connector *mysqldb.Connector
	firstDay  time.Time
}

func New(firstDay time.Time, conn *mysqldb.Connector) *Scheduler {
	return &Scheduler{connector: conn, firstDay: firstDay.Truncate(time.Hour)}
}

type LongID uint64
type ID uint16
type CodeID [2]byte

type Query struct {
	NodeID CodeID

	From time.Time
	To   time.Time

	Offset uint64
	Limit  uint64

	Region      CodeID
	Area        ID
	Locality    ID
	Sublocality ID
}

type TimeSlot struct {
	NodeID CodeID

	HousingID LongID
	LotID     LongID

	StartAt time.Time
	EndAt   time.Time

	Region      CodeID
	Area        ID
	Locality    ID
	Sublocality ID
}

func (unit *Scheduler) Search(
	ctx context.Context,
	query Query,
) ([]TimeSlot, error) {
	qry := mysqldb.Query{
		NodeID: mysqldb.CodeID(query.NodeID),
		Region: mysqldb.CodeID(query.Region),
		From:   unit.numberHoursAfterFirstDay(query.From),
		To:     unit.numberHoursAfterFirstDay(query.To),
		Offset: query.Offset,
		Limit:  query.Limit,
	}

	records, err := unit.connector.List(ctx, qry)
	if err != nil {
		return nil, fmt.Errorf("failed to get records, %w", err)
	}

	result := make([]TimeSlot, len(records))
	for idx, rec := range records {
		result[idx] = TimeSlot{
			HousingID: LongID(rec.HousingID),
			LotID:     LongID(rec.LotID),

			Region:      CodeID(rec.Region),
			Area:        ID(rec.Area),
			Locality:    ID(rec.Locality),
			Sublocality: ID(rec.Sublocality),
		}
	}

	return result, nil
}

func (unit *Scheduler) numberHoursAfterFirstDay(point time.Time) uint16 {
	cur := point.Truncate(time.Hour).Unix()
	first := unit.firstDay.Unix()
	hourInSeconds := time.Duration(time.Hour).Seconds()

	return uint16((cur - first) / int64(hourInSeconds))
}

func (unit *Scheduler) Book(ctx context.Context, slot TimeSlot) error {
	txw, err := unit.connector.Transaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to make tx, %w", err)
	}

	query := mysqldb.Query{
		NodeID: mysqldb.CodeID(slot.NodeID),

		HousingID:   uint64(slot.HousingID),
		LotID:       uint64(slot.LotID),
		Region:      mysqldb.CodeID(slot.Region),
		Area:        uint16(slot.Area),
		Locality:    uint16(slot.Locality),
		Sublocality: uint16(slot.Sublocality),
		From:        unit.numberHoursAfterFirstDay(slot.StartAt),
		To:          unit.numberHoursAfterFirstDay(slot.EndAt),
		Limit:       1,
	}

	result, err := unit.connector.List(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get the previous slot, %w", err)
	}

	if len(result) != 1 {
		return fmt.Errorf("result has not , %w", err)
	}

	prevSlot := result[0]

	slotStartAt := unit.numberHoursAfterFirstDay(slot.StartAt)
	slotEndAt := unit.numberHoursAfterFirstDay(slot.EndAt)
	rec := mysqldb.Record{
		ID:     prevSlot.ID,
		NodeID: mysqldb.CodeID(slot.NodeID),

		HousingID: uint64(slot.HousingID),
		LotID:     uint64(slot.LotID),

		Region:      mysqldb.CodeID(slot.Region),
		Area:        uint16(slot.Area),
		Locality:    uint16(slot.Locality),
		Sublocality: uint16(slot.Sublocality),

		StartAt: prevSlot.StartAt,
		EndAt:   slotStartAt,
	}

	if slotStartAt == prevSlot.StartAt {
		rec.StartAt = slotEndAt
		rec.EndAt = prevSlot.EndAt
	}

	err = mysqldb.Upd(ctx, txw, rec)
	txw.Error(err)

	if slotEndAt != prevSlot.EndAt && slotStartAt != prevSlot.StartAt {
		rec.StartAt = slotEndAt
		rec.EndAt = prevSlot.EndAt

		err = mysqldb.Add(ctx, txw, rec)
		txw.Error(err)
	}

	if err := txw.TransactionEnd(); err != nil {
		return fmt.Errorf("failed to add record, %w", err)
	}

	return nil
}

func (unit *Scheduler) Cancel(ctx context.Context, slot TimeSlot) error {
	txw, err := unit.connector.Transaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to make tx, %w", err)
	}

	defer func() error {
		if err := txw.TransactionEnd(); err != nil {
			return fmt.Errorf("failed to cancel a slot, %w", err)
		}

		return nil
	}()

	query := mysqldb.Query{
		HousingID:   uint64(slot.HousingID),
		LotID:       uint64(slot.LotID),
		NodeID:      mysqldb.CodeID(slot.NodeID),
		Region:      mysqldb.CodeID(slot.Region),
		Area:        uint16(slot.Area),
		Locality:    uint16(slot.Locality),
		Sublocality: uint16(slot.Sublocality),
		From:        unit.numberHoursAfterFirstDay(slot.StartAt),
		To:          unit.numberHoursAfterFirstDay(slot.EndAt),
	}

	rec, err := mysqldb.Get(ctx, txw, query)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to get a record, %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		err = mysqldb.Add(ctx, txw, mysqldb.Record{
			NodeID:      mysqldb.CodeID(slot.NodeID),
			HousingID:   uint64(slot.HousingID),
			LotID:       uint64(slot.LotID),
			Region:      mysqldb.CodeID(slot.Region),
			Area:        uint16(slot.Area),
			Locality:    uint16(slot.Locality),
			Sublocality: uint16(slot.Sublocality),
			StartAt:     unit.numberHoursAfterFirstDay(slot.StartAt),
			EndAt:       unit.numberHoursAfterFirstDay(slot.EndAt),
		})
		txw.Error(err)

		return nil
	}

	if rec.EndAt == query.From {
		rec.StartAt = 0
		rec.EndAt = query.To
	}

	if rec.StartAt == query.To {
		rec.StartAt = query.From
		rec.EndAt = 0
	}

	err = mysqldb.Upd(ctx, txw, *rec)
	txw.Error(err)

	return nil
}

const (
	minDay = 0
	maxDay = 65535
)

func (unit *Scheduler) RegisterLot(ctx context.Context, slot TimeSlot) error {
	rec := mysqldb.Record{
		HousingID: uint64(slot.HousingID),
		LotID:     uint64(slot.LotID),

		Region:      mysqldb.CodeID(slot.Region),
		Area:        uint16(slot.Area),
		Locality:    uint16(slot.Locality),
		Sublocality: uint16(slot.Sublocality),

		StartAt: minDay,
		EndAt:   maxDay,
	}

	err := unit.connector.Add(ctx, mysqldb.CodeID(slot.NodeID), rec)
	if err != nil {
		return fmt.Errorf("failed to add record, %w", err)
	}

	return nil
}
