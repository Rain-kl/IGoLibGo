// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

import { BaseService } from '@/lib/services/core';
import type {
  CancelReservationRequest,
  ReservationOperationResponse,
  ReservationResponse,
} from './types';

export class IGoReservationService extends BaseService {
  protected static readonly basePath = '/api/v1/igo/reservation';

  static async getReservation(): Promise<ReservationResponse> {
    return this.get<ReservationResponse>('');
  }

  static async refreshReservation(): Promise<ReservationOperationResponse> {
    return this.post<ReservationOperationResponse>('/refresh');
  }

  static async cancelReservation(
    data?: CancelReservationRequest,
  ): Promise<ReservationOperationResponse> {
    return this.post<ReservationOperationResponse>(
      '/cancel',
      data as unknown as Record<string, unknown>,
    );
  }
}
