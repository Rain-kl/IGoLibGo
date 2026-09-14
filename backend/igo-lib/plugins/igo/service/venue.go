// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"context"
	"net/http"

	"Wavelet/igo-lib/plugins/igo/consts"
)

// ListLibraries loads venues for the authorized account.
func (s *Service) ListLibraries(ctx context.Context, userID uint64) ([]do.LibrarySummary, error) {
	cookie, _, err := s.cookie(ctx, userID)
	if err != nil {
		return nil, err
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	libs, err := s.client.ListLibraries(ctx, tpl, cookie)
	return libs, wrapTrace(err)
}

// GetBoundLibrary returns the locked venue.
func (s *Service) GetBoundLibrary(ctx context.Context, userID uint64) (*do.BoundLibraryResponse, error) {
	v, err := dao.GetVenue(ctx, userID)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return &do.BoundLibraryResponse{Bound: false}, nil
	}
	return &do.BoundLibraryResponse{Bound: true, Library: toVenueSummary(v)}, nil
}

// RefreshBoundLibrary reloads the locked venue layout.
func (s *Service) RefreshBoundLibrary(ctx context.Context, userID uint64) (*do.BoundLibraryResponse, error) {
	v, err := dao.GetVenue(ctx, userID)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, consts.NewError(http.StatusConflict, consts.CodeConflict, "尚未锁定场馆")
	}
	return s.BindLibrary(ctx, userID, v.LibraryID)
}

// GetLibrary returns one venue summary from the live list.
func (s *Service) GetLibrary(ctx context.Context, userID uint64, libraryID int) (*do.LibrarySummary, error) {
	libs, err := s.ListLibraries(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range libs {
		if libs[i].LibraryID == libraryID {
			return &libs[i], nil
		}
	}
	return nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "场馆不存在")
}

// GetLibraryLayout returns the seat map.
func (s *Service) GetLibraryLayout(ctx context.Context, userID uint64, libraryID int) (*do.LibraryLayoutResponse, error) {
	cookie, _, err := s.cookie(ctx, userID)
	if err != nil {
		return nil, err
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	layout, err := s.client.GetLayout(ctx, tpl, cookie, libraryID)
	return layout, wrapTrace(err)
}

// GetLibraryRule returns opening/booking rules.
func (s *Service) GetLibraryRule(ctx context.Context, userID uint64, libraryID int) (*do.LibraryRuleResponse, error) {
	cookie, _, err := s.cookie(ctx, userID)
	if err != nil {
		return nil, err
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil, err
	}
	rule, err := s.client.GetRule(ctx, tpl, cookie, libraryID)
	return rule, wrapTrace(err)
}

// BindLibrary locks a venue.
func (s *Service) BindLibrary(ctx context.Context, userID uint64, libraryID int) (*do.BoundLibraryResponse, error) {
	layout, err := s.GetLibraryLayout(ctx, userID, libraryID)
	if err != nil {
		return nil, err
	}
	row := &entity.Venue{
		UserID:      userID,
		LibraryID:   layout.LibraryID,
		Name:        layout.Name,
		Floor:       layout.Floor,
		IsOpen:      layout.IsOpen,
		TotalSeats:  layout.TotalSeats,
		UsedSeats:   layout.UsedSeats,
		BookedSeats: layout.BookedSeats,
	}
	if err := dao.UpsertVenue(ctx, row); err != nil {
		return nil, err
	}
	return &do.BoundLibraryResponse{Bound: true, Library: toVenueSummary(row), Layout: layout}, nil
}

// PreviewLibrary loads a venue without locking it.
func (s *Service) PreviewLibrary(ctx context.Context, userID uint64, libraryID int) (*do.LibraryLayoutResponse, error) {
	return s.GetLibraryLayout(ctx, userID, libraryID)
}

// GetFavorites returns favorite seats for a venue.
func (s *Service) GetFavorites(ctx context.Context, userID uint64, libraryID int) ([]do.SeatRef, error) {
	rows, err := dao.ListFavorites(ctx, userID, libraryID)
	if err != nil {
		return nil, err
	}
	out := make([]do.SeatRef, 0, len(rows))
	for _, r := range rows {
		out = append(out, do.SeatRef{SeatKey: r.SeatKey, SeatName: r.SeatName})
	}
	return out, nil
}

// SaveFavorites replaces favorite seats for a venue.
func (s *Service) SaveFavorites(ctx context.Context, userID uint64, libraryID int, seats []do.SeatRef) error {
	items := make([]entity.Favorite, 0, len(seats))
	for _, seat := range seats {
		items = append(items, entity.Favorite{SeatKey: seat.SeatKey, SeatName: seat.SeatName})
	}
	return dao.ReplaceFavorites(ctx, userID, libraryID, items)
}

// SetSeatLabels writes labels onto selected seats.
func (s *Service) SetSeatLabels(ctx context.Context, userID uint64, libraryID int, req do.SetSeatLabelsRequest) ([]do.SeatLabel, error) {
	items := make([]entity.SeatLabel, 0, len(req.Seats))
	for _, seat := range req.Seats {
		items = append(items, entity.SeatLabel{SeatKey: seat.SeatKey, SeatName: seat.SeatName, LabelText: req.Text})
	}
	if err := dao.UpsertSeatLabels(ctx, userID, libraryID, items); err != nil {
		return nil, err
	}
	return s.listLabels(ctx, userID, libraryID)
}

// DeleteSeatLabels removes labels by seat key.
func (s *Service) DeleteSeatLabels(ctx context.Context, userID uint64, libraryID int, seatKeys []string) error {
	return dao.DeleteSeatLabels(ctx, userID, libraryID, seatKeys)
}

func (s *Service) listLabels(ctx context.Context, userID uint64, libraryID int) ([]do.SeatLabel, error) {
	rows, err := dao.ListSeatLabels(ctx, userID, libraryID)
	if err != nil {
		return nil, err
	}
	out := make([]do.SeatLabel, 0, len(rows))
	for _, r := range rows {
		out = append(out, do.SeatLabel{SeatKey: r.SeatKey, SeatName: r.SeatName, Text: r.LabelText})
	}
	return out, nil
}
