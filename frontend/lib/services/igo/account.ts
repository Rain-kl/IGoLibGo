// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

import { BaseService } from '@/lib/services/core';
import type {
  AccountCheckinAuthResponse,
  AccountDTO,
  CreateAccountRequest,
} from './types';

export class IGoAccountService extends BaseService {
  protected static readonly basePath = '/api/v1/igo/accounts';

  static list(): Promise<AccountDTO[]> {
    return this.get<AccountDTO[]>('');
  }

  static create(data: CreateAccountRequest): Promise<AccountDTO> {
    return this.post<AccountDTO>(
      '',
      data as unknown as Record<string, unknown>,
    );
  }

  static getOne(id: string): Promise<AccountDTO> {
    return this.get<AccountDTO>(`/${id}`);
  }

  static update(id: string, data: { name: string }): Promise<AccountDTO> {
    return this.put<AccountDTO>(
      `/${id}`,
      data as unknown as Record<string, unknown>,
    );
  }

  static remove(id: string): Promise<void> {
    return this.delete<void>(`/${id}`);
  }

  static login(id: string, data: { code: string }): Promise<AccountDTO> {
    return this.post<AccountDTO>(
      `/${id}/login`,
      data as unknown as Record<string, unknown>,
    );
  }

  static checkinAuth(
    id: string,
    data: { code: string },
  ): Promise<AccountCheckinAuthResponse> {
    return this.post<AccountCheckinAuthResponse>(
      `/${id}/checkin-auth`,
      data as unknown as Record<string, unknown>,
    );
  }
}
