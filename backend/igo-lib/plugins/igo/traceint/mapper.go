// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package traceint

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"Wavelet/igo-lib/plugins/igo/model/do"
)

func errf(msg string) error { return fmt.Errorf("%s", msg) }

func mapLibraries(raw []byte) ([]do.LibrarySummary, error) {
	root, err := parseRoot(raw)
	if err != nil {
		return nil, err
	}
	libs := dive(root, "data", "userAuth", "reserve", "libs")
	if libs == nil {
		return nil, errf("场馆列表响应无效")
	}
	arr, _ := libs.([]any)
	out := make([]do.LibrarySummary, 0, len(arr))
	for _, item := range arr {
		obj, _ := item.(map[string]any)
		if obj == nil {
			continue
		}
		floor := asString(obj["lib_floor"])
		if floor == "0" {
			continue
		}
		rt, _ := obj["lib_rt"].(map[string]any)
		out = append(out, do.LibrarySummary{
			LibraryID:   asInt(obj["lib_id"]),
			Name:        orDefault(asString(obj["lib_name"]), "Unknown"),
			Floor:       floor,
			IsOpen:      asBool(obj["is_open"]),
			TotalSeats:  asInt(rt["seats_total"]),
			UsedSeats:   asInt(rt["seats_used"]),
			BookedSeats: asInt(rt["seats_booking"]),
		})
	}
	return out, nil
}

func mapLayout(raw []byte) (*do.LibraryLayoutResponse, error) {
	root, err := parseRoot(raw)
	if err != nil {
		return nil, err
	}
	libs, _ := dive(root, "data", "userAuth", "reserve", "libs").([]any)
	if len(libs) == 0 {
		return nil, errf("场馆布局响应无效")
	}
	lib, _ := libs[0].(map[string]any)
	layout, _ := lib["lib_layout"].(map[string]any)
	if lib == nil || layout == nil {
		return nil, errf("场馆布局响应无效")
	}
	seatsRaw, _ := layout["seats"].([]any)
	seats := make([]do.SeatSnapshot, 0, len(seatsRaw))
	invalid := 0
	for _, s := range seatsRaw {
		el, _ := s.(map[string]any)
		if el == nil {
			invalid++
			continue
		}
		typ := 1
		if _, ok := el["type"]; ok {
			typ = asInt(el["type"])
		}
		if typ != 1 {
			continue
		}
		key := strings.TrimSpace(asString(el["key"]))
		name := strings.TrimSpace(asString(el["name"]))
		occupied, ok := asBoolOk(el["status"])
		if key == "" || !ok {
			invalid++
			continue
		}
		if name == "" {
			name = key
		}
		snap := do.SeatSnapshot{
			SeatKey:    key,
			SeatName:   name,
			IsOccupied: occupied,
			X:          asInt(el["x"]),
			Y:          asInt(el["y"]),
		}
		if _, has := el["seat_status"]; has {
			v := asInt(el["seat_status"])
			snap.SeatStatus = &v
		}
		seats = append(seats, snap)
	}
	maxX := asIntPtr(layout["max_x"])
	maxY := asIntPtr(layout["max_y"])
	total := asInt(layout["seats_total"])
	booked := asInt(layout["seats_booking"])
	used := asInt(layout["seats_used"])
	return &do.LibraryLayoutResponse{
		LibrarySummary: do.LibrarySummary{
			LibraryID:   asInt(lib["lib_id"]),
			Name:        orDefault(asString(lib["lib_name"]), "Unknown"),
			Floor:       asString(lib["lib_floor"]),
			IsOpen:      asBool(lib["is_open"]),
			TotalSeats:  total,
			UsedSeats:   used,
			BookedSeats: booked,
		},
		Seats:                  seats,
		MaxX:                   maxX,
		MaxY:                   maxY,
		AvailableSeats:         max(0, total-booked-used),
		InvalidLayoutItemCount: invalid,
	}, nil
}

func mapRule(raw []byte, libraryID int) (*do.LibraryRuleResponse, error) {
	root, err := parseRoot(raw)
	if err != nil {
		return nil, err
	}
	rule, _ := dive(root, "data", "userAuth", "reserve", "libRule").(map[string]any)
	if rule == nil {
		return nil, errf("场馆规则响应无效")
	}
	return &do.LibraryRuleResponse{
		LibraryID:        libraryID,
		AdvanceBooking:   asString(rule["advance_booking"]),
		SeatTTLMinutes:   asString(rule["lib_seat_ttl"]),
		HoldTTLMinutes:   asString(rule["lib_hold_ttl"]),
		RenewTimeMinutes: asString(rule["lib_renew_time"]),
		HoldReasonJSON:   asString(rule["hold_reason"]),
		CloseStartDate:   asString(rule["close_start_date"]),
		CloseEndDate:     asString(rule["close_end_date"]),
		OpenTime:         asInt64(rule["open_time"]),
		OpenTimeText:     asString(rule["open_time_str"]),
		CloseTime:        asInt64(rule["close_time"]),
		CloseTimeText:    asString(rule["close_time_str"]),
		ValidateTime:     asInt(rule["lib_validate_time"]),
	}, nil
}

func mapReservation(raw []byte) (*do.ReservationResponse, error) {
	root, err := parseRoot(raw)
	if err != nil {
		return nil, err
	}
	reserveNode, _ := dive(root, "data", "userAuth", "reserve").(map[string]any)
	if reserveNode == nil {
		return nil, errf("预约响应无效")
	}
	reservation, _ := reserveNode["reserve"].(map[string]any)
	if reservation == nil {
		return &do.ReservationResponse{HasReservation: false}, nil
	}
	exp := asInt64(reservation["exp_date"])
	expTime := time.Unix(exp, 0).UTC().Format(time.RFC3339)
	return &do.ReservationResponse{
		HasReservation:   true,
		ReservationToken: asString(reserveNode["getSToken"]),
		LibraryID:        asInt(reservation["lib_id"]),
		LibraryName:      asString(reservation["lib_name"]),
		SeatKey:          asString(reservation["seat_key"]),
		SeatName:         asString(reservation["seat_name"]),
		ExpirationTime:   expTime,
	}, nil
}

func mapReserveOK(raw []byte) (bool, error) {
	root, err := parseRoot(raw)
	if err != nil {
		return false, err
	}
	v := dive(root, "data", "userAuth", "reserve", "reserueSeat")
	ok, _ := asBoolOk(v)
	return ok, nil
}

func mapCancelOK(raw []byte) bool {
	var root any
	if err := json.Unmarshal(raw, &root); err != nil {
		return false
	}
	if msg, found := findErrorMessage(root); found && strings.Contains(msg, "成功") {
		return true
	}
	if dive(root, "data", "userAuth", "reserve", "reserveCancle") != nil {
		return true
	}
	return false
}

func parseRoot(raw []byte) (any, error) {
	var root any
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, errf("TraceInt 响应不是 JSON")
	}
	if err := throwGraphQLError(root); err != nil {
		return nil, err
	}
	return root, nil
}

func throwGraphQLError(root any) error {
	obj, _ := root.(map[string]any)
	if obj == nil {
		return nil
	}
	errs, _ := obj["errors"].([]any)
	if len(errs) == 0 {
		return nil
	}
	if msg, found := findErrorMessage(errs[0]); found {
		return errf(msg)
	}
	return errf("TraceInt 返回错误")
}

func findErrorMessage(v any) (string, bool) {
	switch t := v.(type) {
	case map[string]any:
		if msg := asString(t["message"]); msg != "" {
			return msg, true
		}
		if msg := asString(t["msg"]); msg != "" {
			return msg, true
		}
		for _, child := range t {
			if msg, ok := findErrorMessage(child); ok {
				return msg, true
			}
		}
	case []any:
		for _, child := range t {
			if msg, ok := findErrorMessage(child); ok {
				return msg, true
			}
		}
	}
	return "", false
}

func dive(v any, keys ...string) any {
	cur := v
	for _, k := range keys {
		switch t := cur.(type) {
		case map[string]any:
			cur = t[k]
		case []any:
			if k == "0" && len(t) > 0 {
				cur = t[0]
			} else {
				return nil
			}
		default:
			return nil
		}
	}
	return cur
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func asInt(v any) int { return int(asInt64(v)) }

func asInt64(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case json.Number:
		i, _ := n.Int64()
		return i
	case int:
		return int64(n)
	case int64:
		return n
	case string:
		var i int64
		for _, r := range n {
			if r < '0' || r > '9' {
				return 0
			}
			i = i*10 + int64(r-'0')
		}
		return i
	default:
		return 0
	}
}

func asIntPtr(v any) *int {
	if v == nil {
		return nil
	}
	n := asInt(v)
	return &n
}

func asBool(v any) bool {
	ok, _ := asBoolOk(v)
	return ok
}

func asBoolOk(v any) (bool, bool) {
	switch t := v.(type) {
	case bool:
		return t, true
	case float64:
		return t != 0, true
	case string:
		s := strings.ToLower(strings.TrimSpace(t))
		if s == "true" || s == "1" || s == "yes" {
			return true, true
		}
		if s == "false" || s == "0" || s == "no" || s == "" {
			return false, true
		}
	}
	return false, false
}

func orDefault(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
