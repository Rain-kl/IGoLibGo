// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Compass } from 'lucide-react';
import { useUser } from '@/contexts/user-context';
import { IGoService } from '@/lib/services/igo';
import type {
  ActivityLogEntry,
  DashboardResponse,
} from '@/lib/services/igo/types';

import { HeroStatusCard } from '@/components/igo/dashboard/hero-status-card';
import { MetricsOverview } from '@/components/igo/dashboard/metrics-overview';
import { ReservationProgressCard } from '@/components/igo/dashboard/reservation-progress-card';
import { QuickTaskControls } from '@/components/igo/dashboard/quick-task-controls';
import { ActivityLogStream } from '@/components/igo/common/activity-log-stream';

export default function DashboardRootPage() {
  const t = useTranslations('igo.dashboard');
  const { user } = useUser();

  const [dashboard, setDashboard] = React.useState<DashboardResponse | null>(
    null,
  );
  const [dashboardLoading, setDashboardLoading] = React.useState(true);

  const [logs, setLogs] = React.useState<ActivityLogEntry[]>([]);
  const [logsLoading, setLogsLoading] = React.useState(true);

  const loadData = React.useCallback(async (isSilent = false) => {
    if (!isSilent) {
      setDashboardLoading(true);
      setLogsLoading(true);
    }
    try {
      const [dashRes, logsRes] = await Promise.all([
        IGoService.dashboard.getDashboard().catch(() => null),
        IGoService.dashboard.listActivityLogs({ limit: 50 }).catch(() => []),
      ]);
      setDashboard(dashRes);
      setLogs(logsRes || []);
    } finally {
      if (!isSilent) {
        setDashboardLoading(false);
        setLogsLoading(false);
      }
    }
  }, []);

  React.useEffect(() => {
    loadData();

    // 10 秒轻量轮询更新引擎状态与活动流水
    const interval = setInterval(() => {
      loadData(true);
    }, 10000);

    return () => clearInterval(interval);
  }, [loadData]);

  return (
    <div className='py-6 px-1 space-y-6 w-full'>
      {/* 1. 标准页面标题 */}
      <div className='flex items-center gap-2'>
        <Compass className='size-5 text-primary' />
        <div>
          <h1 className='text-2xl font-semibold tracking-tight'>
            {t('title')}
          </h1>
          <p className='text-sm text-muted-foreground mt-0.5'>
            {t('description')}
          </p>
        </div>
      </div>

      {/* 2. 问候与时钟看板 */}
      <HeroStatusCard
        dashboard={dashboard}
        userName={user?.nickname || user?.username || '同学'}
      />

      {/* 3. 统计核心指标 */}
      <MetricsOverview dashboard={dashboard} />

      {/* 4. 今日在座预约进度条卡片 */}
      <ReservationProgressCard
        reservation={dashboard?.reservation ?? null}
        loading={dashboardLoading}
        onRefresh={() => loadData(true)}
      />

      {/* 5. 四大自动化引擎快捷调度 */}
      <QuickTaskControls tasks={dashboard?.tasks ?? []} />

      {/* 6. 实时活动日志流 */}
      <ActivityLogStream
        logs={logs}
        loading={logsLoading}
        onRefresh={() => loadData(true)}
      />
    </div>
  );
}
