// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Rocket } from 'lucide-react';
import { toast } from 'sonner';

import { IGoService } from '@/lib/services/igo';
import type {
  ActivityLogEntry,
  BoundLibraryResponse,
  CoordinatorStatus,
  SeatRef,
} from '@/lib/services/igo/types';

import { GrabStatusCard } from '@/components/igo/grab/grab-status-card';
import { GrabConfigForm } from '@/components/igo/grab/grab-config-form';
import { ActivityLogStream } from '@/components/igo/common/activity-log-stream';

export default function GrabPage() {
  const t = useTranslations('igo.grab');
  const tCommon = useTranslations('common');

  // 目标场馆与配置
  const [boundInfo, setBoundInfo] = React.useState<BoundLibraryResponse | null>(
    null,
  );
  const [seats, setSeats] = React.useState<SeatRef[]>([]);
  const [scheduledStart, setScheduledStart] = React.useState('');
  const [strategy, setStrategy] = React.useState('query_then_reserve');
  const [minDelay, setMinDelay] = React.useState(200);
  const [maxDelay, setMaxDelay] = React.useState(800);

  // 运行状态与日志
  const [status, setStatus] = React.useState<CoordinatorStatus | null>(null);
  const [logs, setLogs] = React.useState<ActivityLogEntry[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [starting, setStarting] = React.useState(false);
  const [cancelling, setCancelling] = React.useState(false);

  // 初始加载场馆信息并尝试从 sessionStorage 恢复来自场馆选座的候选列表
  React.useEffect(() => {
    try {
      const stored = sessionStorage.getItem('igo_grab_candidates');
      if (stored) {
        const parsed: SeatRef[] = JSON.parse(stored);
        if (Array.isArray(parsed) && parsed.length > 0) {
          setSeats(parsed);
          sessionStorage.removeItem('igo_grab_candidates');
          toast.success(`已载入 ${parsed.length} 个候选抢座目标`);
        }
      }
    } catch {
      // 忽略解析错误
    }
  }, []);

  const loadData = React.useCallback(async (isSilent = false) => {
    if (!isSilent) setLoading(true);
    try {
      const [boundRes, tasksRes, logsRes] = await Promise.all([
        IGoService.venue.getBoundLibrary().catch(() => null),
        IGoService.task.listTasks().catch(() => null),
        IGoService.dashboard.listActivityLogs({ limit: 40 }).catch(() => []),
      ]);
      setBoundInfo(boundRes);
      const grabTask = tasksRes?.tasks?.find((task) => task.kind === 'grab');
      setStatus(grabTask ?? null);
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
    if (!boundInfo?.library) {
      toast.error('请先锁定目标场馆');
      return;
    }
    if (seats.length === 0) {
      toast.error('请至少添加一个候选座位');
      return;
    }

    setStarting(true);
    try {
      const res = await IGoService.task.startGrab({
        library_id: boundInfo.library.library_id,
        library_name: boundInfo.library.name,
        seats,
        reservation_strategy: strategy,
        scheduled_start: scheduledStart.trim() || undefined,
        polling_min_delay_ms: minDelay,
        polling_max_delay_ms: maxDelay,
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
      const res = await IGoService.task.cancelGrab();
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
        <Rocket className='size-5 text-primary' />
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
      <GrabStatusCard
        status={status}
        loading={loading}
        starting={starting}
        cancelling={cancelling}
        onStart={handleStart}
        onCancel={handleCancel}
        canStart={!!boundInfo?.library && seats.length > 0}
      />

      {/* 3. 参数配置表单 */}
      <GrabConfigForm
        boundInfo={boundInfo}
        seats={seats}
        onSeatsChange={setSeats}
        scheduledStart={scheduledStart}
        onScheduledStartChange={setScheduledStart}
        strategy={strategy}
        onStrategyChange={setStrategy}
        minDelay={minDelay}
        onMinDelayChange={setMinDelay}
        maxDelay={maxDelay}
        onMaxDelayChange={setMaxDelay}
      />

      {/* 4. 实时活动日志 */}
      <ActivityLogStream
        logs={logs}
        loading={loading}
        onRefresh={() => loadData(true)}
      />
    </div>
  );
}
