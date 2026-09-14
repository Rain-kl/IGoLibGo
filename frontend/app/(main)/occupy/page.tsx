// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { ShieldCheck } from 'lucide-react';
import { toast } from 'sonner';

import { IGoService } from '@/lib/services/igo';
import type {
  ActivityLogEntry,
  CoordinatorStatus,
  ReservationResponse,
} from '@/lib/services/igo/types';

import { OccupyControlCard } from '@/components/igo/occupy/occupy-control-card';
import { OccupyIntervalConfig } from '@/components/igo/occupy/occupy-interval-config';
import { ReservationProgressCard } from '@/components/igo/dashboard/reservation-progress-card';
import { ActivityLogStream } from '@/components/igo/common/activity-log-stream';

export default function OccupyPage() {
  const t = useTranslations('igo.occupy');
  const tCommon = useTranslations('common');

  const [reservation, setReservation] =
    React.useState<ReservationResponse | null>(null);
  const [status, setStatus] = React.useState<CoordinatorStatus | null>(null);
  const [logs, setLogs] = React.useState<ActivityLogEntry[]>([]);
  const [loading, setLoading] = React.useState(true);

  const [intervalMode, setIntervalMode] = React.useState('fixed_10s');
  const [reReserveDelay, setReReserveDelay] = React.useState(60);

  const [starting, setStarting] = React.useState(false);
  const [cancelling, setCancelling] = React.useState(false);

  const loadData = React.useCallback(async (isSilent = false) => {
    if (!isSilent) setLoading(true);
    try {
      const [resRes, tasksRes, logsRes] = await Promise.all([
        IGoService.reservation.getReservation().catch(() => null),
        IGoService.task.listTasks().catch(() => null),
        IGoService.dashboard.listActivityLogs({ limit: 40 }).catch(() => []),
      ]);

      setReservation(resRes);
      const occupyTask = tasksRes?.tasks?.find(
        (task) => task.kind === 'occupy',
      );
      setStatus(occupyTask ?? null);
      setLogs(logsRes || []);
    } finally {
      if (!isSilent) setLoading(false);
    }
  }, []);

  React.useEffect(() => {
    loadData();
    const timer = setInterval(() => loadData(true), 5000);
    return () => clearInterval(timer);
  }, [loadData]);

  const handleStart = async () => {
    if (!reservation?.has_reservation) {
      toast.error('当前暂无有效预约，请先在【账户与场馆】中完成预约');
      return;
    }

    setStarting(true);
    try {
      const res = await IGoService.task.startOccupy({
        re_reserve_delay_seconds: reReserveDelay,
        check_interval_mode: intervalMode,
      });
      setStatus(res);
      toast.success(tCommon('save'));
      loadData(true);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setStarting(false);
    }
  };

  const handleCancel = async () => {
    setCancelling(true);
    try {
      const res = await IGoService.task.cancelOccupy();
      setStatus(res);
      toast.success(tCommon('save'));
      loadData(true);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setCancelling(false);
    }
  };

  return (
    <div className='py-6 px-1 space-y-6 w-full'>
      {/* 1. 标准标题 */}
      <div className='flex items-center gap-2'>
        <ShieldCheck className='size-5 text-primary' />
        <div>
          <h1 className='text-2xl font-semibold tracking-tight'>
            {t('title')}
          </h1>
          <p className='text-sm text-muted-foreground mt-0.5'>
            {t('description')}
          </p>
        </div>
      </div>

      {/* 2. 续座守护控制面板 */}
      <OccupyControlCard
        status={status}
        loading={loading}
        starting={starting}
        cancelling={cancelling}
        onStart={handleStart}
        onCancel={handleCancel}
        canStart={reservation?.has_reservation ?? false}
      />

      {/* 3. 当前在座预约信息卡片 */}
      <ReservationProgressCard
        reservation={reservation}
        loading={loading}
        onRefresh={() => loadData(true)}
      />

      {/* 4. 续座策略配置 */}
      <OccupyIntervalConfig
        intervalMode={intervalMode}
        onIntervalModeChange={setIntervalMode}
        reReserveDelaySeconds={reReserveDelay}
        onReReserveDelaySecondsChange={setReReserveDelay}
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
