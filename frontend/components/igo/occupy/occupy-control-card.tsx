// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

'use client';

import * as React from 'react';
import { useTranslations } from 'next-intl';
import { Play, ShieldCheck, StopCircle } from 'lucide-react';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import { TaskStatusBadge } from '@/components/igo/common/task-status-badge';
import type { CoordinatorStatus } from '@/lib/services/igo/types';

interface OccupyControlCardProps {
  status: CoordinatorStatus | null;
  loading: boolean;
  starting: boolean;
  cancelling: boolean;
  onStart: () => void;
  onCancel: () => void;
  canStart: boolean;
}

export function OccupyControlCard({
  status,
  loading,
  starting,
  cancelling,
  onStart,
  onCancel,
  canStart,
}: OccupyControlCardProps) {
  const t = useTranslations('igo.occupy');

  const state = status?.state || 'idle';
  const isRunning = state === 'running' || state === 'starting';

  return (
    <Card className='border-border/60 shadow-sm'>
      <CardHeader className='pb-3'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-2'>
            <ShieldCheck className='size-4 text-primary' />
            <CardTitle className='text-base font-semibold'>
              {t('statusTitle')}
            </CardTitle>
          </div>
          <TaskStatusBadge state={state} />
        </div>
        <CardDescription>
          {status?.message || '自动续座守护防被违规暂离释放'}
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        <div className='grid grid-cols-2 sm:grid-cols-4 gap-3 p-3 rounded-lg bg-muted/40 text-xs'>
          <div>
            <div className='text-muted-foreground'>运行状态</div>
            <div className='font-semibold text-foreground mt-0.5 capitalize'>
              {state}
            </div>
          </div>
          <div>
            <div className='text-muted-foreground'>续座检查轮次</div>
            <div className='font-mono font-semibold text-foreground mt-0.5'>
              {status?.poll_count ?? 0} 次
            </div>
          </div>
          <div>
            <div className='text-muted-foreground'>续约执行次数</div>
            <div className='font-mono font-semibold text-foreground mt-0.5'>
              {status?.request_count ?? 0} 次
            </div>
          </div>
          <div>
            <div className='text-muted-foreground'>最近心跳</div>
            <div className='font-mono font-semibold text-foreground mt-0.5'>
              {status?.last_request_at
                ? status.last_request_at.slice(11, 19)
                : '--:--:--'}
            </div>
          </div>
        </div>

        <div className='flex items-center gap-2 pt-1'>
          {!isRunning ? (
            <Button
              variant='default'
              size='sm'
              onClick={onStart}
              disabled={!canStart || starting || loading}
              className='gap-1.5'
            >
              {starting ? (
                <Spinner className='size-3.5' />
              ) : (
                <Play className='size-3.5' />
              )}
              <span>{t('startBtn')}</span>
            </Button>
          ) : (
            <Button
              variant='destructive'
              size='sm'
              onClick={onCancel}
              disabled={cancelling}
              className='gap-1.5'
            >
              {cancelling ? (
                <Spinner className='size-3.5' />
              ) : (
                <StopCircle className='size-3.5' />
              )}
              <span>{t('stopBtn')}</span>
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
