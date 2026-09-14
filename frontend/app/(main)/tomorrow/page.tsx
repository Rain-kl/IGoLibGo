// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Calendar } from 'lucide-react';
import { toast } from 'sonner';

import { IGoService } from '@/lib/services/igo';
import type {
  ActivityLogEntry,
  BoundLibraryResponse,
  CoordinatorStatus,
  SeatRef,
  TaskLaunchRecord,
} from '@/lib/services/igo/types';

import { TomorrowStatusCard } from '@/components/igo/tomorrow/tomorrow-status-card';
import { TomorrowConfigForm } from '@/components/igo/tomorrow/tomorrow-config-form';
import { TomorrowHistoryTable } from '@/components/igo/tomorrow/tomorrow-history-table';
import { ActivityLogStream } from '@/components/igo/common/activity-log-stream';

export default function TomorrowPage() {
  const t = useTranslations('igo.tomorrow');
  const tCommon = useTranslations('common');

  const [boundInfo, setBoundInfo] = React.useState<BoundLibraryResponse | null>(
    null,
  );
  const [targetSeat, setTargetSeat] = React.useState<SeatRef | null>(null);
  const [scheduledTime, setScheduledTime] = React.useState('22:00:00');

  const [status, setStatus] = React.useState<CoordinatorStatus | null>(null);
  const [records, setRecords] = React.useState<TaskLaunchRecord[]>([]);
  const [logs, setLogs] = React.useState<ActivityLogEntry[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [starting, setStarting] = React.useState(false);
  const [cancelling, setCancelling] = React.useState(false);
  const [runningNow, setRunningNow] = React.useState(false);

  const loadData = React.useCallback(
    async (isSilent = false) => {
      if (!isSilent) setLoading(true);
      try {
        const [boundRes, tasksRes, recsRes, logsRes] = await Promise.all([
          IGoService.venue.getBoundLibrary().catch(() => null),
          IGoService.task.listTasks().catch(() => null),
          IGoService.task.listTaskRecords('tomorrow').catch(() => []),
          IGoService.dashboard.listActivityLogs({ limit: 40 }).catch(() => []),
        ]);

        setBoundInfo(boundRes);
        const tomorrowTask = tasksRes?.tasks?.find(
          (task) => task.kind === 'tomorrow',
        );
        setStatus(tomorrowTask ?? null);
        setRecords(recsRes || []);
        setLogs(logsRes || []);

        // 默认选择场馆第 1 个空闲座位或收藏
        if (
          boundRes?.layout?.seats &&
          boundRes.layout.seats.length > 0 &&
          !targetSeat
        ) {
          const firstAvail =
            boundRes.layout.seats.find((s) => !s.is_occupied) ||
            boundRes.layout.seats[0];
          setTargetSeat({
            seat_key: firstAvail.seat_key,
            seat_name: firstAvail.seat_name || firstAvail.seat_key,
          });
        }
      } finally {
        if (!isSilent) setLoading(false);
      }
    },
    [targetSeat],
  );

  React.useEffect(() => {
    loadData();
    const timer = setInterval(() => loadData(true), 5000);
    return () => clearInterval(timer);
  }, [loadData]);

  const handleStart = async () => {
    if (!boundInfo?.library) {
      toast.error('请先锁定目标场馆');
      return;
    }
    if (!targetSeat) {
      toast.error('请选择明日目标座位');
      return;
    }
    if (!scheduledTime.trim()) {
      toast.error('请输入定时预约时间');
      return;
    }

    setStarting(true);
    try {
      const res = await IGoService.task.startTomorrow({
        library_id: boundInfo.library.library_id,
        library_name: boundInfo.library.name,
        seat: targetSeat,
        scheduled_start: scheduledTime.trim(),
        execute_immediately: false,
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
      const res = await IGoService.task.cancelTomorrow();
      setStatus(res);
      toast.success(tCommon('save'));
      loadData(true);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setCancelling(false);
    }
  };

  const handleRunNow = async () => {
    setRunningNow(true);
    try {
      const res = await IGoService.task.runTomorrowNow();
      setStatus(res);
      toast.success('已触发立即执行！');
      loadData(true);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : tCommon('unknownError'));
    } finally {
      setRunningNow(false);
    }
  };

  return (
    <div className='py-6 px-1 space-y-6 w-full'>
      {/* 1. 标准页面标题 */}
      <div className='flex items-center gap-2'>
        <Calendar className='size-5 text-primary' />
        <div>
          <h1 className='text-2xl font-semibold tracking-tight'>
            {t('title')}
          </h1>
          <p className='text-sm text-muted-foreground mt-0.5'>
            {t('description')}
          </p>
        </div>
      </div>

      {/* 2. 状态看板 */}
      <TomorrowStatusCard
        status={status}
        loading={loading}
        starting={starting}
        cancelling={cancelling}
        runningNow={runningNow}
        onStart={handleStart}
        onCancel={handleCancel}
        onRunNow={handleRunNow}
        canStart={!!boundInfo?.library && !!targetSeat}
      />

      {/* 3. 参数配置表单 */}
      <TomorrowConfigForm
        boundInfo={boundInfo}
        seat={targetSeat}
        scheduledTime={scheduledTime}
        onScheduledTimeChange={setScheduledTime}
      />

      {/* 4. 历史记录表格 */}
      <TomorrowHistoryTable records={records} />

      {/* 5. 实时活动日志 */}
      <ActivityLogStream
        logs={logs}
        loading={loading}
        onRefresh={() => loadData(true)}
      />
    </div>
  );
}
