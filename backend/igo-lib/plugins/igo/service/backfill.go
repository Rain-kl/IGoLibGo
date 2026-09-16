// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"context"
	"strconv"
	"strings"
)

const defaultAccountName = "默认"

// BackfillAccountsAndCheckinInfos copies legacy sessions, pipeline credentials,
// and venue profiles into igo_accounts / igo_checkin_infos. It is idempotent.
func (s *Service) BackfillAccountsAndCheckinInfos(ctx context.Context) error {
	s.backfillOnce.Do(func() {
		s.backfillErr = s.backfillSessionsAndProfiles(ctx)
	})
	if s.backfillErr != nil {
		return s.backfillErr
	}
	return s.backfillPipelineRows(ctx)
}

func (s *Service) backfillSessionsAndProfiles(ctx context.Context) error {
	if err := s.backfillSessions(ctx); err != nil {
		return err
	}
	if err := s.backfillCheckInSessions(ctx); err != nil {
		return err
	}
	return s.backfillSettingsProfiles(ctx)
}

func (s *Service) backfillSessions(ctx context.Context) error {
	sessions, err := dao.ListSessions(ctx)
	if err != nil {
		return err
	}
	for i := range sessions {
		row := sessions[i]
		if strings.TrimSpace(row.Cookie) == "" {
			continue
		}
		existing, err := dao.ListAccountsByUser(ctx, row.UserID)
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			continue
		}
		acc := &entity.Account{
			UserID:          row.UserID,
			Name:            defaultAccountName,
			Cookie:          row.Cookie,
			CookieExpiresAt: row.ExpiresAt,
			CookieSource:    row.Source,
		}
		if err := dao.CreateAccount(ctx, acc); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) backfillCheckInSessions(ctx context.Context) error {
	rows, err := dao.ListAllCheckInSessions(ctx)
	if err != nil {
		return err
	}
	for i := range rows {
		sess := rows[i]
		if strings.TrimSpace(sess.Token) == "" {
			continue
		}
		acc, err := s.ensureDefaultAccount(ctx, sess.UserID)
		if err != nil {
			return err
		}
		if acc.CheckinToken != "" {
			continue
		}
		acc.CheckinToken = sess.Token
		acc.CheckinExpiresAt = sess.ExpiresAt
		if err := dao.UpdateAccount(ctx, acc); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) backfillPipelineRows(ctx context.Context) error {
	cards, err := dao.ListAllPipelineConfigs(ctx)
	if err != nil {
		return err
	}
	for i := range cards {
		if err := s.backfillOnePipeline(ctx, &cards[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) backfillOnePipeline(ctx context.Context, card *entity.PipelineConfig) error {
	if card.AccountID != 0 {
		return nil
	}
	acc, err := dao.FindAccountByCookie(ctx, card.UserID, card.Cookie)
	if err != nil {
		return err
	}
	if acc == nil {
		name := strings.TrimSpace(card.Name)
		if name == "" {
			name = defaultAccountName
		}
		acc = &entity.Account{
			UserID:           card.UserID,
			Name:             name,
			Cookie:           card.Cookie,
			CookieExpiresAt:  card.CookieExpiresAt,
			CheckinToken:     card.CheckinToken,
			CheckinExpiresAt: card.CheckinExpiresAt,
		}
		if err := dao.CreateAccount(ctx, acc); err != nil {
			return err
		}
	} else if acc.CheckinToken == "" && strings.TrimSpace(card.CheckinToken) != "" {
		acc.CheckinToken = card.CheckinToken
		acc.CheckinExpiresAt = card.CheckinExpiresAt
		if err := dao.UpdateAccount(ctx, acc); err != nil {
			return err
		}
	}
	card.AccountID = acc.ID
	if hasPipelineBeacon(card) && card.CheckinInfoID == 0 {
		info := &entity.CheckInInfo{
			UserID:     card.UserID,
			Name:       strings.TrimSpace(card.Name) + " 签到",
			BeaconUUID: card.BeaconUUID,
			Major:      card.Major,
			Minor:      card.Minor,
			Latitude:   card.Latitude,
			Longitude:  card.Longitude,
		}
		if err := dao.CreateCheckInInfo(ctx, info); err != nil {
			return err
		}
		card.CheckinInfoID = info.ID
	}
	return dao.UpdatePipelineConfig(ctx, card)
}

func (s *Service) backfillSettingsProfiles(ctx context.Context) error {
	// Walk users that have settings rows via pipeline/session accounts is incomplete;
	// load settings through known user IDs from sessions, accounts, and pipelines.
	userIDs := map[uint64]struct{}{}
	sessions, err := dao.ListSessions(ctx)
	if err != nil {
		return err
	}
	for i := range sessions {
		userIDs[sessions[i].UserID] = struct{}{}
	}
	cards, err := dao.ListAllPipelineConfigs(ctx)
	if err != nil {
		return err
	}
	for i := range cards {
		userIDs[cards[i].UserID] = struct{}{}
	}
	for userID := range userIDs {
		st := s.settings(ctx, userID)
		if len(st.CheckInProfiles) == 0 {
			continue
		}
		existing, err := dao.ListCheckInInfosByUser(ctx, userID)
		if err != nil {
			return err
		}
		have := map[string]struct{}{}
		for i := range existing {
			have[existing[i].Name] = struct{}{}
		}
		for libID, profile := range st.CheckInProfiles {
			name := strings.TrimSpace(profile.LibraryName)
			if name == "" {
				name = "场馆 " + strconv.Itoa(libID)
			}
			if _, ok := have[name]; ok {
				continue
			}
			info := &entity.CheckInInfo{
				UserID:     userID,
				Name:       name,
				BeaconUUID: profile.BeaconUUID,
				Major:      profile.Major,
				Minor:      profile.Minor,
				Latitude:   strconv.FormatFloat(profile.Latitude, 'f', -1, 64),
				Longitude:  strconv.FormatFloat(profile.Longitude, 'f', -1, 64),
			}
			if err := dao.CreateCheckInInfo(ctx, info); err != nil {
				return err
			}
			have[name] = struct{}{}
		}
	}
	return nil
}

func (s *Service) ensureDefaultAccount(ctx context.Context, userID uint64) (*entity.Account, error) {
	list, err := dao.ListAccountsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].Name == defaultAccountName {
			return &list[i], nil
		}
	}
	if len(list) > 0 {
		return &list[0], nil
	}
	acc := &entity.Account{UserID: userID, Name: defaultAccountName}
	if err := dao.CreateAccount(ctx, acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func hasPipelineBeacon(card *entity.PipelineConfig) bool {
	return strings.TrimSpace(card.BeaconUUID) != "" ||
		strings.TrimSpace(card.Latitude) != "" ||
		strings.TrimSpace(card.Longitude) != "" ||
		card.Major != 0 ||
		card.Minor != 0
}
