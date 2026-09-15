// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

import { IGoCheckInService } from './checkin';
import { IGoConfigService } from './config';
import { IGoDashboardService } from './dashboard';
import { IGoPipelineService } from './pipeline';
import { IGoReservationService } from './reservation';
import { IGoSessionService } from './session';
import { IGoTaskService } from './tasks';
import { IGoVenueService } from './venue';

export * from './types';
export * from './dashboard';
export * from './session';
export * from './venue';
export * from './reservation';
export * from './tasks';
export * from './checkin';
export * from './config';
export * from './pipeline';

export const IGoService = {
  dashboard: IGoDashboardService,
  session: IGoSessionService,
  venue: IGoVenueService,
  reservation: IGoReservationService,
  task: IGoTaskService,
  checkin: IGoCheckInService,
  config: IGoConfigService,
  pipeline: IGoPipelineService,
} as const;

export default IGoService;
