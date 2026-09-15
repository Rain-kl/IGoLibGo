// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

export interface SeatRef {
  seat_key: string;
  seat_name: string;
}

export interface SeatSnapshot {
  seat_key: string;
  seat_name: string;
  is_occupied: boolean;
  x: number;
  y: number;
  seat_status?: number;
}

export interface SeatLabel {
  seat_key: string;
  seat_name: string;
  text: string;
}

export interface LibrarySummary {
  library_id: number;
  name: string;
  floor: string;
  is_open: boolean;
  total_seats: number;
  used_seats: number;
  booked_seats: number;
}

export interface LibraryLayoutResponse extends LibrarySummary {
  seats: SeatSnapshot[];
  max_x?: number;
  max_y?: number;
  available_seats: number;
  invalid_layout_item_count: number;
}

export interface LibraryRuleResponse {
  library_id: number;
  advance_booking: string;
  seat_ttl_minutes: string;
  hold_ttl_minutes: string;
  renew_time_minutes: string;
  hold_reason_json: string;
  close_start_date?: string;
  close_end_date?: string;
  open_time: number;
  open_time_text: string;
  close_time: number;
  close_time_text: string;
  validate_time: number;
}

export interface BoundLibraryResponse {
  bound: boolean;
  library?: LibrarySummary;
  layout?: LibraryLayoutResponse;
}

export interface SaveFavoritesRequest {
  seats: SeatRef[];
}

export interface SetSeatLabelsRequest {
  seats: SeatRef[];
  text: string;
}

export interface DeleteSeatLabelsRequest {
  seat_keys: string[];
}

export interface SessionResponse {
  authorized: boolean;
  source?: string;
  saved_at?: string;
  expires_at?: string;
  can_auto_restore: boolean;
  cookie_masked?: string;
}

export interface AuthFromCodeRequest {
  code: string;
  remember?: boolean;
}

export interface AuthFromCookieRequest {
  cookie: string;
  remember?: boolean;
}

export interface SessionWorkflowResponse {
  session: SessionResponse;
  libraries?: LibrarySummary[];
  message?: string;
}

export interface QRCodeResponse {
  image_data_url: string;
  auth_url: string;
  expires_at?: string;
}

export interface ReservationResponse {
  has_reservation: boolean;
  reservation_token?: string;
  library_id?: number;
  library_name?: string;
  seat_key?: string;
  seat_name?: string;
  expiration_time?: string;
}

export interface CancelReservationRequest {
  stop_occupy_first?: boolean;
}

export interface ReservationOperationResponse {
  reservation: ReservationResponse;
  message?: string;
}

export interface CoordinatorStatus {
  kind: string;
  state:
    | 'idle'
    | 'starting'
    | 'running'
    | 'stopping'
    | 'completed'
    | 'failed'
    | string;
  title: string;
  message: string;
  started_at?: string;
  last_updated_at?: string;
  poll_count: number;
  request_count: number;
  last_request_at?: string;
  reason?: string;
  is_active: boolean;
}

export interface DashboardResponse {
  authorized: boolean;
  hero_status: string;
  hero_status_detail: string;
  historical_success_count: number;
  total_guard_seconds: number;
  engine_summary: string;
  bound_library?: LibrarySummary;
  reservation: ReservationResponse;
  tasks: CoordinatorStatus[];
}

export interface TaskListResponse {
  tasks: CoordinatorStatus[];
}

export interface GrabStartRequest {
  library_id: number;
  library_name?: string;
  seats: SeatRef[];
  polling_mode?: string;
  reservation_strategy?: string;
  scheduled_start?: string;
  polling_min_delay_ms?: number;
  polling_max_delay_ms?: number;
}

export interface OccupyStartRequest {
  re_reserve_delay_seconds: number;
  check_interval_mode?: string;
}

export interface GlobalLeakLibraryTarget {
  library_id: number;
  library_name: string;
  floor: string;
}

export interface GlobalLeakStartRequest {
  libraries: GlobalLeakLibraryTarget[];
  scan_interval_seconds: number;
}

export interface TomorrowStartRequest {
  library_id: number;
  library_name?: string;
  seat: SeatRef;
  scheduled_start: string;
  execute_immediately?: boolean;
}

export interface TaskLaunchRecord {
  record_id: string;
  kind: string;
  recorded_at: string;
  library_id?: number;
  library_name?: string;
  seats?: SeatRef[];
  libraries?: GlobalLeakLibraryTarget[];
  scan_interval_seconds?: number;
  polling_mode?: string;
  reservation_strategy?: string;
}

export interface GlobalLeakBlacklistResponse {
  items: Record<number, SeatRef[]>;
}

export interface SaveGlobalLeakBlacklistRequest {
  items: Record<number, SeatRef[]>;
}

export interface SaveGlobalLeakSelectedLibrariesRequest {
  libraries: GlobalLeakLibraryTarget[];
}

export interface ActivityLogEntry {
  id: string;
  level: 'info' | 'warn' | 'error' | string;
  kind: string;
  message: string;
  created_at: string;
}

export interface CheckInSessionResponse {
  authorized: boolean;
  saved_at?: string;
  expires_at?: string;
  can_auto_restore: boolean;
}

export interface CheckInAuthFromCodeRequest {
  code: string;
  remember?: boolean;
}

export interface CheckInDeviceResponse {
  nickname: string;
  school: string;
  student_name: string;
  student_number: string;
  beacon_uuids: string[];
}

export interface CheckInSignRequest {
  expected_library_id: number;
  expected_library_name?: string;
  beacon_uuid: string;
  major?: number;
  minor?: number;
  latitude?: number;
  longitude?: number;
}

export interface CheckInSignResponse {
  message: string;
  status?: number;
  library_id?: number;
  library_name?: string;
  library_floor?: string;
  seat_key?: string;
  seat_name?: string;
  signed_at?: string;
  expiration_time?: string;
}

export interface CheckInAuthorizationResponse {
  session: CheckInSessionResponse;
  device?: CheckInDeviceResponse;
  device_refresh_warning?: string;
}

export interface CheckInVenueProfile {
  library_id: number;
  library_name: string;
  beacon_uuid: string;
  major: number;
  minor: number;
  latitude: number;
  longitude: number;
  updated_at?: string;
}

export interface CheckInVenueProfilesResponse {
  profiles: CheckInVenueProfile[];
}

export interface SaveCheckInVenueProfileRequest {
  library_name?: string;
  beacon_uuid: string;
  major: number;
  minor: number;
  latitude: number;
  longitude: number;
}

export interface ProtocolTemplatesResponse {
  get_cookie_url_template: string;
  cookie_authorization_return_url: string;
  graphql_endpoint_url: string;
  graphql_default_referer_url: string;
  graphql_default_origin_url: string;
  graphql_tomorrow_referer_url: string;
  graphql_tomorrow_origin_url: string;
  tomorrow_reservation_queue_url_template: string;
  remote_checkin_auth_url_template: string;
  remote_checkin_authorization_return_url: string;
  remote_checkin_auth_referer_url: string;
  remote_checkin_devices_endpoint_url: string;
  remote_checkin_time_endpoint_url: string;
  remote_checkin_sign_endpoint_url: string;
  remote_checkin_api_referer_url: string;
  query_libraries_template: string;
  query_library_layout_template: string;
  query_library_rule_template: string;
  query_reservation_info_template: string;
  reserve_seat_template: string;
  cancel_reservation_template: string;
  tomorrow_reservation_warmup_template: string;
  tomorrow_reservation_save_template: string;
  tomorrow_reservation_info_template: string;
}

export interface SaveProtocolTemplatesRequest {
  overrides: ProtocolTemplatesResponse;
}

export interface SettingsResponse {
  request_timeout_seconds: number;
  network_max_retries: number;
  traceint_graphql_overrides_enabled: boolean;
  grab_reservation_strategy: string;
  optimal_grab_strategy_reminder_enabled: boolean;
  grab_scheduled_start_default?: string;
  tomorrow_scheduled_start_default?: string;
  occupy_re_reserve_delay_seconds: number;
  occupy_check_interval_mode: string;
  global_leak_scan_interval_seconds: number;
  auto_release_enabled: boolean;
  auto_release_lead_seconds: number;
  home_reservation_progress_mode?: string;
}

export interface SaveSettingsRequest {
  request_timeout_seconds?: number;
  network_max_retries?: number;
  traceint_graphql_overrides_enabled?: boolean;
  grab_reservation_strategy?: string;
  optimal_grab_strategy_reminder_enabled?: boolean;
  grab_scheduled_start_default?: string;
  tomorrow_scheduled_start_default?: string;
  occupy_re_reserve_delay_seconds?: number;
  occupy_check_interval_mode?: string;
  global_leak_scan_interval_seconds?: number;
  auto_release_enabled?: boolean;
  auto_release_lead_seconds?: number;
  home_reservation_progress_mode?: string;
}

export interface BackupExportRequest {
  password: string;
}

export interface BackupExportResponse {
  filename: string;
  content: string;
}

export interface BackupImportRequest {
  password: string;
  content: string;
}

export interface PipelineConfigDTO {
  id: string;
  name: string;
  cookie?: string;
  cookie_masked?: string;
  has_cookie?: boolean;
  cookie_expires_at?: string;
  library_id: number;
  library_name: string;
  floor: string;
  seat_key: string;
  seat_name: string;
  auto_checkin: boolean;
  checkin_token?: string;
  has_checkin_token?: boolean;
  checkin_expires_at?: string;
  beacon_uuid?: string;
  major?: number;
  minor?: number;
  latitude?: string;
  longitude?: string;
  created_at: string;
  updated_at: string;
}

export interface CreatePipelineConfigRequest {
  id: string;
  name: string;
  cookie: string;
  library_id: number;
  library_name: string;
  floor: string;
  seat_key: string;
  seat_name: string;
  auto_checkin: boolean;
  checkin_token?: string;
  beacon_uuid?: string;
  major?: number;
  minor?: number;
  latitude?: string;
  longitude?: string;
}

export interface UpdatePipelineConfigRequest {
  name?: string;
  cookie?: string;
  library_id?: number;
  library_name?: string;
  floor?: string;
  seat_key?: string;
  seat_name?: string;
  auto_checkin?: boolean;
  checkin_token?: string;
  beacon_uuid?: string;
  major?: number;
  minor?: number;
  latitude?: string;
  longitude?: string;
}

export interface RunPipelineRequest {
  cookie?: string;
  checkin_token?: string;
}

export interface PipelineExecutionResult {
  config_id: string;
  name: string;
  success: boolean;
  need_auth?: string; // "LOGIN" | "CHECKIN"
  auth_url?: string;
  message: string;
  reservation_status?: string;
  checkin_status?: string;
  executed_at: string;
}

export interface PipelineHelperVerifySessionRequest {
  cookie?: string;
  auth_code?: string;
  auth_url?: string;
}

export interface PipelineHelperVerifySessionResponse {
  valid?: boolean;
  cookie: string;
  expires_at?: string;
  libraries?: LibrarySummary[];
}

export interface PipelineHelperVerifyCheckinRequest {
  checkin_token?: string;
  checkin_code?: string;
  checkin_url?: string;
}

export interface PipelineHelperVerifyCheckinResponse {
  valid?: boolean;
  token: string;
  expires_at?: string;
}
