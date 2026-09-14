// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

import { BaseService } from '@/lib/services/core';
import type {
  ActivityLogEntry,
  DashboardResponse,
  TaskListResponse,
} from './types';

export class IGoDashboardService extends BaseService {
  protected static readonly basePath = '/api/v1/igo';

  static async getDashboard(): Promise<DashboardResponse> {
    return this.get<DashboardResponse>('/dashboard');
  }

  static async getStatus(): Promise<TaskListResponse> {
    return this.get<TaskListResponse>('/status');
  }

  static async listActivityLogs(params?: {
    limit?: number;
    before_id?: string;
  }): Promise<ActivityLogEntry[]> {
    return this.get<ActivityLogEntry[]>('/activity-logs', {
      params: params as Record<string, unknown>,
    });
  }
}
