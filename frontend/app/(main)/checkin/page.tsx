// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { MapPin } from 'lucide-react';

import { IGoService } from '@/lib/services/igo';
import type {
  ActivityLogEntry,
  BoundLibraryResponse,
  CheckInDeviceResponse,
  CheckInSessionResponse,
} from '@/lib/services/igo/types';

import { CheckInAuthCard } from '@/components/igo/checkin/checkin-auth-card';
import { CheckInDeviceTable } from '@/components/igo/checkin/checkin-device-table';
import { CheckInActionPanel } from '@/components/igo/checkin/checkin-action-panel';
import { ActivityLogStream } from '@/components/igo/common/activity-log-stream';

export default function CheckInPage() {
  const t = useTranslations('igo.checkin');

  const [session, setSession] = React.useState<CheckInSessionResponse | null>(
    null,
  );
  const [device, setDevice] = React.useState<CheckInDeviceResponse | null>(
    null,
  );
  const [boundInfo, setBoundInfo] = React.useState<BoundLibraryResponse | null>(
    null,
  );
  const [logs, setLogs] = React.useState<ActivityLogEntry[]>([]);
  const [loading, setLoading] = React.useState(true);

  const loadData = React.useCallback(async (isSilent = false) => {
    if (!isSilent) setLoading(true);
    try {
      const [sessRes, devRes, boundRes, logsRes] = await Promise.all([
        IGoService.checkin.getSession().catch(() => null),
        IGoService.checkin.getDevices().catch(() => null),
        IGoService.venue.getBoundLibrary().catch(() => null),
        IGoService.dashboard.listActivityLogs({ limit: 40 }).catch(() => []),
      ]);

      setSession(sessRes);
      setDevice(devRes);
      setBoundInfo(boundRes);
      setLogs(logsRes || []);
    } finally {
      if (!isSilent) setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    loadData();
    const timer = setInterval(() => loadData(true), 10000);
    return () => clearInterval(timer);
  }, [loadData]);

  return (
    <div className='py-6 px-1 space-y-6 w-full'>
      {/* 1. 标准页面标题 */}
      <div className='flex items-center gap-2'>
        <MapPin className='size-5 text-primary' />
        <div>
          <h1 className='text-2xl font-semibold tracking-tight'>
            {t('title')}
          </h1>
          <p className='text-sm text-muted-foreground mt-0.5'>
            {t('description')}
          </p>
        </div>
      </div>

      {/* 2. 微信签到授权卡片 */}
      <CheckInAuthCard
        session={session}
        loading={loading}
        onRefresh={() => loadData(true)}
      />

      {/* 3. 授权设备与 Beacon 列表 */}
      <CheckInDeviceTable device={device} loading={loading} />

      {/* 4. 打卡操作与模拟配置 */}
      <CheckInActionPanel
        boundInfo={boundInfo}
        device={device}
        onSuccess={() => loadData(true)}
      />

      {/* 5. 实时活动日志 */}
      <ActivityLogStream
        logs={logs}
        loading={loading}
        onRefresh={() => loadData(true)}
      />
    </div>
  );
}
