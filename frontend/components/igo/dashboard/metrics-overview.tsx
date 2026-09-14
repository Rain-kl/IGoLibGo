// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Award, Cpu, Shield, Zap } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
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
      <Card className='border-border/60 shadow-sm'>
        <CardContent className='p-4 flex items-center gap-4'>
          <div className='size-11 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0'>
            <Award className='size-5.5' />
          </div>
          <div className='space-y-0.5 min-w-0'>
            <div className='text-xs text-muted-foreground truncate'>
              {t('successCount')}
            </div>
            <div className='text-2xl font-bold tracking-tight text-foreground'>
              {successCount}
              <span className='text-xs font-normal text-muted-foreground ml-1.5'>
                {t('unitTimes')}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card className='border-border/60 shadow-sm'>
        <CardContent className='p-4 flex items-center gap-4'>
          <div className='size-11 rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center shrink-0'>
            <Shield className='size-5.5' />
          </div>
          <div className='space-y-0.5 min-w-0'>
            <div className='text-xs text-muted-foreground truncate'>
              {t('guardTime')}
            </div>
            <div className='text-2xl font-bold tracking-tight text-foreground'>
              {guardHours}
              <span className='text-xs font-normal text-muted-foreground ml-1.5'>
                {t('unitHours')}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card className='border-border/60 shadow-sm'>
        <CardContent className='p-4 flex items-center gap-4'>
          <div className='size-11 rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400 flex items-center justify-center shrink-0'>
            <Cpu className='size-5.5' />
          </div>
          <div className='space-y-0.5 min-w-0'>
            <div className='text-xs text-muted-foreground truncate'>
              {t('activeEngines')}
            </div>
            <div className='text-2xl font-bold tracking-tight text-foreground'>
              {activeEnginesCount}
              <span className='text-xs font-normal text-muted-foreground ml-1.5'>
                / 4 个引擎
              </span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
