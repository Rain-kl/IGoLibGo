// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Award, Cpu, Shield } from 'lucide-react';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import type { DashboardResponse } from '@/lib/services/igo/types';

interface MetricsOverviewProps {
  dashboard: DashboardResponse | null;
}

export function MetricsOverview({ dashboard }: MetricsOverviewProps) {
  const t = useTranslations('igo.dashboard.metrics');

  const successCount = dashboard?.historical_success_count ?? 0;
  const guardHours = React.useMemo(() => {
    const secs = dashboard?.total_guard_seconds ?? 0;
    return (secs / 3600).toFixed(1);
  }, [dashboard?.total_guard_seconds]);

  const activeEnginesCount = React.useMemo(() => {
    return dashboard?.tasks?.filter((t) => t.is_active)?.length ?? 0;
  }, [dashboard?.tasks]);

  return (
    <div className='grid grid-cols-1 sm:grid-cols-3 gap-4'>
      <Card className='border-dashed shadow-none'>
        <CardHeader className='flex flex-row items-center justify-between pb-2'>
          <span className='text-xs font-medium text-muted-foreground'>
            {t('successCount')}
          </span>
          <Award className='size-4 text-primary' />
        </CardHeader>
        <CardContent className='space-y-1'>
          <div className='text-2xl font-semibold tracking-tight text-foreground'>
            {successCount}
            <span className='text-xs font-normal text-muted-foreground ml-1.5'>
              {t('unitTimes')}
            </span>
          </div>
        </CardContent>
      </Card>

      <Card className='border-dashed shadow-none'>
        <CardHeader className='flex flex-row items-center justify-between pb-2'>
          <span className='text-xs font-medium text-muted-foreground'>
            {t('guardTime')}
          </span>
          <Shield className='size-4 text-emerald-500' />
        </CardHeader>
        <CardContent className='space-y-1'>
          <div className='text-2xl font-semibold tracking-tight text-foreground'>
            {guardHours}
            <span className='text-xs font-normal text-muted-foreground ml-1.5'>
              {t('unitHours')}
            </span>
          </div>
        </CardContent>
      </Card>

      <Card className='border-dashed shadow-none'>
        <CardHeader className='flex flex-row items-center justify-between pb-2'>
          <span className='text-xs font-medium text-muted-foreground'>
            {t('activeEngines')}
          </span>
          <Cpu className='size-4 text-amber-500' />
        </CardHeader>
        <CardContent className='space-y-1'>
          <div className='text-2xl font-semibold tracking-tight text-foreground'>
            {activeEnginesCount}
            <span className='text-xs font-normal text-muted-foreground ml-1.5'>
              / 4 个引擎
            </span>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
