// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

import { BaseService } from '@/lib/services/core';
import type {
  CheckInAuthFromCodeRequest,
  CheckInAuthorizationResponse,
  CheckInDeviceResponse,
  CheckInSessionResponse,
  CheckInSignRequest,
  CheckInSignResponse,
  CheckInVenueProfile,
  CheckInVenueProfilesResponse,
  QRCodeResponse,
  SaveCheckInVenueProfileRequest,
} from './types';

export class IGoCheckInService extends BaseService {
  protected static readonly basePath = '/api/v1/igo/checkin';

  static async getSession(): Promise<CheckInSessionResponse> {
    return this.get<CheckInSessionResponse>('/session');
  }

  static async getAuthQRCode(): Promise<QRCodeResponse> {
    return this.get<QRCodeResponse>('/auth-qrcode');
  }

  static async authorizeFromCode(
    data: CheckInAuthFromCodeRequest,
  ): Promise<CheckInAuthorizationResponse> {
    return this.post<CheckInAuthorizationResponse>(
      '/from-code',
      data as unknown as Record<string, unknown>,
    );
  }

  static async getDevices(): Promise<CheckInDeviceResponse> {
    return this.get<CheckInDeviceResponse>('/devices');
  }

  static async signCheckIn(
    data: CheckInSignRequest,
  ): Promise<CheckInSignResponse> {
    return this.post<CheckInSignResponse>(
      '/sign',
      data as unknown as Record<string, unknown>,
    );
  }

  static async clearSession(): Promise<void> {
    return this.delete<void>('/session');
  }

  static async getVenueProfile(
    libraryId: number,
  ): Promise<CheckInVenueProfile | null> {
    return this.get<CheckInVenueProfile | null>(`/profiles/${libraryId}`);
  }

  static async saveVenueProfile(
    libraryId: number,
    data: SaveCheckInVenueProfileRequest,
  ): Promise<CheckInVenueProfile> {
    return this.put<CheckInVenueProfile>(
      `/profiles/${libraryId}`,
      data as unknown as Record<string, unknown>,
    );
  }

  static async listVenueProfiles(): Promise<CheckInVenueProfilesResponse> {
    return this.get<CheckInVenueProfilesResponse>('/profiles');
  }
}
