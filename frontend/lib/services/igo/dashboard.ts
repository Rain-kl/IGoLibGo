// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

import { BaseService } from '@/lib/services/core';
import type { ActivityLogEntry, DashboardResponse } from './types';

export class IGoDashboardService extends BaseService {
  protected static readonly basePath = '/api/v1/igo';

  static async getDashboard(): Promise<DashboardResponse> {
    return this.get<DashboardResponse>('/dashboard');
  }

  static async getStatus(): Promise<DashboardResponse> {
    return this.get<DashboardResponse>('/status');
  }

  static async listActivityLogs(params?: {
    limit?: number;
    before_id?: string;
    page?: number;
    per_page?: number;
  }): Promise<ActivityLogEntry[]> {
    const queryParams: Record<string, unknown> = {};
    if (params?.page) queryParams.page = params.page;
    if (params?.per_page || params?.limit) {
      queryParams.per_page = params.per_page || params.limit;
    }
    return this.get<ActivityLogEntry[]>('/activity-logs', queryParams);
  }
}
