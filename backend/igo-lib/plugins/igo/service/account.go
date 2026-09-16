// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"Wavelet/igo-lib/plugins/igo/traceint"
	"context"
	"net/http"
	"strings"
)

const maxReferencedPipelineIDs = 5

// ListAccounts returns accounts owned by the user.
func (s *Service) ListAccounts(ctx context.Context, userID uint64) ([]do.AccountDTO, error) {
	rows, err := dao.ListAccountsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]do.AccountDTO, 0, len(rows))
	for i := range rows {
		out = append(out, toAccountDTO(&rows[i]))
	}
	return out, nil
}

// GetAccount returns one account owned by the user.
func (s *Service) GetAccount(ctx context.Context, userID, id uint64) (*do.AccountDTO, error) {
	row, err := dao.GetAccountByUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到该账户")
	}
	dto := toAccountDTO(row)
	return &dto, nil
}

// CreateAccount creates an account. Cookie and check-in token are optional.
func (s *Service) CreateAccount(ctx context.Context, userID uint64, req do.CreateAccountRequest) (*do.AccountDTO, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, consts.NewError(http.StatusBadRequest, consts.CodeValidationError, "账户名称不能为空")
	}
	row := &entity.Account{UserID: userID, Name: name}
	if strings.TrimSpace(req.Cookie) != "" {
		cookie, exp, err := s.resolveCookie(ctx, userID, req.Cookie)
		if err != nil {
			return nil, err
		}
		row.Cookie = cookie
		row.CookieExpiresAt = exp
	}
	if strings.TrimSpace(req.CheckinToken) != "" {
		token, exp, err := s.resolveCheckinToken(ctx, userID, req.CheckinToken)
		if err != nil {
			return nil, err
		}
		row.CheckinToken = token
		row.CheckinExpiresAt = exp
		s.fillAccountProfile(ctx, userID, row)
	}
	if err := dao.CreateAccount(ctx, row); err != nil {
		return nil, err
	}
	dto := toAccountDTO(row)
	return &dto, nil
}

// UpdateAccount updates the account display name.
func (s *Service) UpdateAccount(ctx context.Context, userID, id uint64, req do.UpdateAccountRequest) (*do.AccountDTO, error) {
	row, err := dao.GetAccountByUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到该账户")
	}
	if name := strings.TrimSpace(req.Name); name != "" {
		row.Name = name
	}
	if err := dao.UpdateAccount(ctx, row); err != nil {
		return nil, err
	}
	dto := toAccountDTO(row)
	return &dto, nil
}

// DeleteAccount removes an account that is not referenced by pipeline cards.
func (s *Service) DeleteAccount(ctx context.Context, userID, id uint64) error {
	row, err := dao.GetAccountByUser(ctx, id, userID)
	if err != nil {
		return err
	}
	if row == nil {
		return consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到该账户")
	}
	refs, err := dao.ListPipelineIDsByAccount(ctx, userID, id)
	if err != nil {
		return err
	}
	if len(refs) > 0 {
		return consts.NewError(http.StatusConflict, consts.CodeConflict, "账户仍被一条龙引用: "+joinPipelineIDs(refs))
	}
	return dao.DeleteAccount(ctx, id, userID)
}

// LoginAccount writes occupy credentials onto the account.
func (s *Service) LoginAccount(ctx context.Context, userID, id uint64, req do.AccountLoginRequest) (*do.AccountDTO, error) {
	row, err := dao.GetAccountByUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到该账户")
	}
	cookie, exp, err := s.resolveCookie(ctx, userID, req.Code)
	if err != nil {
		return nil, err
	}
	row.Cookie = cookie
	row.CookieExpiresAt = exp
	if err := dao.UpdateAccount(ctx, row); err != nil {
		return nil, err
	}
	dto := toAccountDTO(row)
	return &dto, nil
}

// AuthorizeAccountCheckin writes check-in credentials and refreshes the student profile.
func (s *Service) AuthorizeAccountCheckin(ctx context.Context, userID, id uint64, req do.AccountCheckinAuthRequest) (*do.AccountCheckinAuthResponse, error) {
	row, err := dao.GetAccountByUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, consts.NewError(http.StatusNotFound, consts.CodeNotFound, "未找到该账户")
	}
	token, exp, err := s.resolveCheckinToken(ctx, userID, req.Code)
	if err != nil {
		return nil, err
	}
	row.CheckinToken = token
	row.CheckinExpiresAt = exp
	device := s.fillAccountProfile(ctx, userID, row)
	if err := dao.UpdateAccount(ctx, row); err != nil {
		return nil, err
	}
	return &do.AccountCheckinAuthResponse{Account: toAccountDTO(row), Device: device}, nil
}

func (s *Service) fillAccountProfile(ctx context.Context, userID uint64, row *entity.Account) *do.CheckInDeviceResponse {
	if strings.TrimSpace(row.CheckinToken) == "" {
		return nil
	}
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return nil
	}
	device, err := s.api(ctx, userID).GetCheckInDevices(ctx, tpl, row.CheckinToken)
	if err != nil || device == nil {
		return nil
	}
	if device.Nickname != "" {
		row.Nickname = device.Nickname
	}
	if device.School != "" {
		row.School = device.School
	}
	if device.StudentName != "" {
		row.StudentName = device.StudentName
	}
	if device.StudentNumber != "" {
		row.StudentNumber = device.StudentNumber
	}
	return device
}

func toAccountDTO(row *entity.Account) do.AccountDTO {
	if row == nil {
		return do.AccountDTO{}
	}
	return do.AccountDTO{
		ID:               row.ID,
		Name:             row.Name,
		HasCookie:        row.Cookie != "",
		CookieMasked:     traceint.MaskCookie(row.Cookie),
		CookieExpiresAt:  row.CookieExpiresAt,
		HasCheckinToken:  row.CheckinToken != "",
		CheckinExpiresAt: row.CheckinExpiresAt,
		Nickname:         row.Nickname,
		School:           row.School,
		StudentName:      row.StudentName,
		StudentNumber:    row.StudentNumber,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

func joinPipelineIDs(ids []string) string {
	if len(ids) > maxReferencedPipelineIDs {
		ids = ids[:maxReferencedPipelineIDs]
	}
	return strings.Join(ids, ", ")
}
