// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

import { BaseService } from '@/lib/services/core';
import type {
  CoordinatorStatus,
  GlobalLeakBlacklistResponse,
  GlobalLeakLibraryTarget,
  GlobalLeakStartRequest,
  GrabStartRequest,
  OccupyStartRequest,
  SaveGlobalLeakBlacklistRequest,
  SaveGlobalLeakSelectedLibrariesRequest,
  TaskLaunchRecord,
  TaskListResponse,
  TomorrowStartRequest,
} from './types';

export class IGoTaskService extends BaseService {
  protected static readonly basePath = '/api/v1/igo';

  static async listTasks(): Promise<TaskListResponse> {
    return this.get<TaskListResponse>('/tasks');
  }

  static async listTaskRecords(kind?: string): Promise<TaskLaunchRecord[]> {
    return this.get<TaskLaunchRecord[]>('/task-records', {
      params: kind ? { kind } : undefined,
    });
  }

  static async runTomorrowNow(): Promise<CoordinatorStatus> {
    return this.post<CoordinatorStatus>('/tasks/tomorrow/run-now');
  }

  static async startGrab(data: GrabStartRequest): Promise<CoordinatorStatus> {
    return this.post<CoordinatorStatus>(
      '/tasks/grab/start',
      data as unknown as Record<string, unknown>,
    );
  }

  static async cancelGrab(): Promise<CoordinatorStatus> {
    return this.post<CoordinatorStatus>('/tasks/grab/cancel');
  }

  static async startLeak(
    data: GlobalLeakStartRequest,
  ): Promise<CoordinatorStatus> {
    return this.post<CoordinatorStatus>(
      '/tasks/leak/start',
      data as unknown as Record<string, unknown>,
    );
  }

  static async cancelLeak(): Promise<CoordinatorStatus> {
    return this.post<CoordinatorStatus>('/tasks/leak/cancel');
  }

  static async startTomorrow(
    data: TomorrowStartRequest,
  ): Promise<CoordinatorStatus> {
    return this.post<CoordinatorStatus>(
      '/tasks/tomorrow/start',
      data as unknown as Record<string, unknown>,
    );
  }

  static async cancelTomorrow(): Promise<CoordinatorStatus> {
    return this.post<CoordinatorStatus>('/tasks/tomorrow/cancel');
  }

  static async startOccupy(
    data: OccupyStartRequest,
  ): Promise<CoordinatorStatus> {
    return this.post<CoordinatorStatus>(
      '/tasks/occupy/start',
      data as unknown as Record<string, unknown>,
    );
  }

  static async cancelOccupy(): Promise<CoordinatorStatus> {
    return this.post<CoordinatorStatus>('/tasks/occupy/cancel');
  }

  static async getGlobalLeakBlacklist(): Promise<GlobalLeakBlacklistResponse> {
    return this.get<GlobalLeakBlacklistResponse>('/global-leak/blacklist');
  }

  static async saveGlobalLeakBlacklist(
    data: SaveGlobalLeakBlacklistRequest,
  ): Promise<GlobalLeakBlacklistResponse> {
    return this.put<GlobalLeakBlacklistResponse>(
      '/global-leak/blacklist',
      data as unknown as Record<string, unknown>,
    );
  }

  static async getGlobalLeakSelectedLibraries(): Promise<
    GlobalLeakLibraryTarget[]
  > {
    return this.get<GlobalLeakLibraryTarget[]>(
      '/global-leak/selected-libraries',
    );
  }

  static async saveGlobalLeakSelectedLibraries(
    data: SaveGlobalLeakSelectedLibrariesRequest,
  ): Promise<GlobalLeakLibraryTarget[]> {
    return this.put<GlobalLeakLibraryTarget[]>(
      '/global-leak/selected-libraries',
      data as unknown as Record<string, unknown>,
    );
  }
}
