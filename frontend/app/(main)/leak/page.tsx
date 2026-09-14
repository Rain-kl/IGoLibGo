// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Ban, Radio, Sliders } from 'lucide-react';
import { toast } from 'sonner';

import { IGoService } from '@/lib/services/igo';
import type {
  ActivityLogEntry,
  CoordinatorStatus,
  GlobalLeakLibraryTarget,
  LibrarySummary,
  SeatRef,
} from '@/lib/services/igo/types';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

import { LeakStatusCard } from '@/components/igo/leak/leak-status-card';
import { LibrarySelectionCard } from '@/components/igo/leak/library-selection-card';
import { SeatBlacklistDialog } from '@/components/igo/leak/seat-blacklist-dialog';
import { ActivityLogStream } from '@/components/igo/common/activity-log-stream';

export default function GlobalLeakPage() {
  const t = useTranslations('igo.leak');
  const tCommon = useTranslations('common');

  const [allLibraries, setAllLibraries] = React.useState<LibrarySummary[]>([]);
  const [selectedLibraries, setSelectedLibraries] = React.useState<
    GlobalLeakLibraryTarget[]
  >([]);
  const [scanInterval, setScanInterval] = React.useState<number>(5);
  const [blacklist, setBlacklist] = React.useState<Record<number, SeatRef[]>>(
    {},
  );
  const [blacklistOpen, setBlacklistOpen] = React.useState(false);

  const [status, setStatus] = React.useState<CoordinatorStatus | null>(null);
  const [logs, setLogs] = React.useState<ActivityLogEntry[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [starting, setStarting] = React.useState(false);
  const [cancelling, setCancelling] = React.useState(false);

  const loadData = React.useCallback(async (isSilent = false) => {
    if (!isSilent) setLoading(true);
    try {
      const [libsRes, tasksRes, logsRes, blRes, savedLibsRes] =
        await Promise.all([
          IGoService.venue.listLibraries().catch(() => []),
          IGoService.task.listTasks().catch(() => null),
          IGoService.dashboard.listActivityLogs({ limit: 40 }).catch(() => []),
          IGoService.task.getGlobalLeakBlacklist().catch(() => ({ items: {} })),
          IGoService.task.getGlobalLeakSelectedLibraries().catch(() => []),
        ]);

      setAllLibraries(libsRes || []);
      const leakTask = tasksRes?.tasks?.find((t) => t.kind === 'leak');
      setStatus(leakTask ?? null);
      setLogs(logsRes || []);
      setBlacklist(blRes.items || {});

      if (savedLibsRes && savedLibsRes.length > 0) {
        setSelectedLibraries(savedLibsRes);
      } else if (libsRes && libsRes.length > 0) {
        // 默认全选前两个场馆
        setSelectedLibraries(
          libsRes.slice(0, 3).map((l) => ({
            library_id: l.library_id,
            library_name: l.name,
            floor: l.floor,
          })),
        );
      }
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
    if (selectedLibraries.length === 0) {
      toast.error('请至少勾选一个捡漏场馆');
      return;
    }

    setStarting(true);
    try {
      // 先保存已选场馆
      await IGoService.task.saveGlobalLeakSelectedLibraries({
        libraries: selectedLibraries,
      });

      const res = await IGoService.task.startLeak({
        libraries: selectedLibraries,
        scan_interval_seconds: scanInterval,
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
      const res = await IGoService.task.cancelLeak();
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
        <Radio className='size-5 text-primary' />
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
      <LeakStatusCard
        status={status}
        loading={loading}
        starting={starting}
        cancelling={cancelling}
        onStart={handleStart}
        onCancel={handleCancel}
        canStart={selectedLibraries.length > 0}
      />

      {/* 3. 扫描参数与黑名单配置 */}
      <Card className='border-dashed shadow-none'>
        <CardHeader className='pb-3'>
          <div className='flex items-center justify-between'>
            <div className='flex items-center gap-2'>
              <Sliders className='size-4 text-primary' />
              <CardTitle className='text-base font-semibold'>
                捡漏策略与黑名单
              </CardTitle>
            </div>
            <Button
              variant='outline'
              size='sm'
              onClick={() => setBlacklistOpen(true)}
              className='h-7 text-xs gap-1 border-dashed shadow-none'
            >
              <Ban className='size-3 text-destructive' />
              <span>{t('editBlacklist')}</span>
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className='max-w-xs space-y-1.5'>
            <Label htmlFor='scan-interval' className='text-xs font-medium'>
              {t('scanInterval')}
            </Label>
            <Input
              id='scan-interval'
              type='number'
              value={scanInterval}
              onChange={(e) =>
                setScanInterval(Math.max(1, Number(e.target.value)))
              }
              min={1}
              max={60}
              className='h-8 text-xs font-mono shadow-none bg-background'
            />
            <p className='text-[10px] text-muted-foreground'>
              推荐设为 3~10 秒，在平稳请求与极速命中之间取得平衡
            </p>
          </div>
        </CardContent>
      </Card>

      {/* 4. 监听场馆多选 */}
      <LibrarySelectionCard
        allLibraries={allLibraries}
        selectedLibraries={selectedLibraries}
        onSelectionChange={setSelectedLibraries}
      />

      {/* 5. 实时活动日志 */}
      <ActivityLogStream
        logs={logs}
        loading={loading}
        onRefresh={() => loadData(true)}
      />

      {/* 6. 黑名单编辑弹窗 */}
      <SeatBlacklistDialog
        open={blacklistOpen}
        onOpenChange={setBlacklistOpen}
        libraries={allLibraries}
        blacklist={blacklist}
        onRefreshBlacklist={() => loadData(true)}
      />
    </div>
  );
}
